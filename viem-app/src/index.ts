// src/index.ts
import { http, Client, createClient, HttpTransport, PublicRpcSchema, PrivateKeyAccount, TypedDataDomain, createWalletClient, TypedDataParameter } from 'viem';
import { privateKeyToAccount } from 'viem/accounts';
import dotenv from 'dotenv';
import { OrbyActions, orbyActions } from '@orb-labs/orby-viem-extension';
import { Account, getOrbyChainId, OperationDataFormat, OperationType, QuoteType } from '@orb-labs/orby-core';

// Load environment variables
dotenv.config();

// Load params from environment variables
const PRIVATE_KEY = process.env.PRIVATE_KEY ?? "";
const ORBY_ENGINE_ADMIN_URL = process.env.ORBY_ENGINE_ADMIN_URL ?? "";
const ORBY_INSTANCE_NAME = process.env.ORBY_INSTANCE_NAME ?? "";
const INPUT_TOKEN_ADDRESS = process.env.INPUT_TOKEN_ADDRESS ?? "";
const OUTPUT_TOKEN_ADDRESS = process.env.OUTPUT_TOKEN_ADDRESS ?? "";
const INPUT_TOKEN_CHAIN_ID = process.env.INPUT_TOKEN_CHAIN_ID ?? "";
const OUTPUT_TOKEN_CHAIN_ID = process.env.OUTPUT_TOKEN_CHAIN_ID ?? "";
const AMOUNT = process.env.AMOUNT ?? "0";

// Ensure everything is set
if (!PRIVATE_KEY) {
  throw new Error("Missing PRIVATE_KEY in .env file");
}
if (!ORBY_ENGINE_ADMIN_URL) {
  throw new Error("Missing ORBY_ENGINE_ADMIN_URL in .env file");
}
if (!ORBY_INSTANCE_NAME) {
  throw new Error("Missing ORBY_INSTANCE_NAME in .env file");
}
if (!INPUT_TOKEN_ADDRESS) {
  throw new Error("Missing INPUT_TOKEN_ADDRESS in .env file");
}
if (!OUTPUT_TOKEN_ADDRESS) {
  throw new Error("Missing OUTPUT_TOKEN_ADDRESS in .env file");
}
if (!INPUT_TOKEN_CHAIN_ID) {
  throw new Error("Missing INPUT_TOKEN_CHAIN_ID in .env file");
}
if (!OUTPUT_TOKEN_CHAIN_ID) {
  throw new Error("Missing OUTPUT_TOKEN_CHAIN_ID in .env file");
}
if (!AMOUNT) {
  throw new Error("Missing AMOUNT in .env file");
}

export const notEmpty = <T>(value: T): value is NonNullable<typeof value> => value != undefined && value != null;

async function main() {
  // Connect to orby admin client
  const orbyAdminProvider: Client<
      HttpTransport,
      undefined,
      undefined,
      PublicRpcSchema,
      OrbyActions
  > = createClient({
  transport: http(ORBY_ENGINE_ADMIN_URL),
  }).extend(orbyActions);

  // Create private instance
  const { success, orbyInstancePrivateUrl } = await orbyAdminProvider.createInstance(ORBY_INSTANCE_NAME);
  if (!success || !orbyInstancePrivateUrl) {
    throw new Error("failed to create instance");
  } else {
    console.log(`[INFO] successfully created instance`);
  }

  // Connect to the private instance
  console.log("[INFO] creating private instance using url: ", orbyInstancePrivateUrl);
  const privateInstanceProvider: Client<
      HttpTransport,
      undefined,
      undefined,
      PublicRpcSchema,
      OrbyActions
  > = createClient({
  transport: http(orbyInstancePrivateUrl),
  }).extend(orbyActions);

  // Get address from private key
  const wallet = privateKeyToAccount(`0x${PRIVATE_KEY}`);
  const address = wallet.address;
  console.log("\n[INFO] derived address from private key: ", address)

  // Create account cluster
  console.log("\n[INFO] creating account cluster...");
  const accounts = [Account.toAccount({ vmType: "EVM", address, accountType: "EOA" })];
  const { accountClusterId } = await privateInstanceProvider.createAccountCluster(accounts);
  if (!accountClusterId) {
    throw new Error("failed to create account cluster");
  }
  
  // Get virtual node RPC URL
  console.log("\n[INFO] getting virtual node RPC URL...");
  console.log(`[INFO]   accountClusterId: ${accountClusterId}`);
  console.log(`[INFO]   chainId: ${BigInt(INPUT_TOKEN_CHAIN_ID)}`);
  console.log(`[INFO]   address: ${address}`);
  const virtualNodeRpcUrl = await privateInstanceProvider.getVirtualNodeRpcUrl(
    accountClusterId,
    BigInt(INPUT_TOKEN_CHAIN_ID),
    address
  );
  if (!virtualNodeRpcUrl) {
    throw new Error("failed to get virtual node RPC URL");
  } else {
    console.log(`[INFO] You can now use this URL to interact with the virtual node: ${virtualNodeRpcUrl}`);
  }

  // Get virtual node client
  const virtualNodeProvider: Client<
      HttpTransport,
      undefined,
      undefined,
      PublicRpcSchema,
      OrbyActions
  > = createClient({
  transport: http(virtualNodeRpcUrl),
  }).extend(orbyActions);

  // Get standardized token ids using the virtual node
  const tokens = [
    { chainId: BigInt(INPUT_TOKEN_CHAIN_ID), tokenAddress: INPUT_TOKEN_ADDRESS },
    { chainId: BigInt(OUTPUT_TOKEN_CHAIN_ID), tokenAddress: OUTPUT_TOKEN_ADDRESS },
  ]
  console.log("\n[INFO] getting standardized token IDs for:");
  tokens.forEach((token, i) => {
      console.log(`         Token ${i + 1}:`);
      console.log(`           Chain ID: ${token.chainId}`);
      console.log(`           Address: ${token.tokenAddress}`);
  });
  const standardizedTokenIds = await virtualNodeProvider.getStandardizedTokenIds(tokens);
  if (!standardizedTokenIds) {
    throw new Error("[ERROR] failed to get standardized token IDs");
  } else {
    console.log(`[INFO] standardized token ids: ${standardizedTokenIds}`);
  }

  // Get operations needed to swap using the virtual node and standardized token ids
  console.log("\n[INFO] calling getOperationsToSwap...");
  const input = {
    standardizedTokenId: standardizedTokenIds[0], 
    amount: BigInt(AMOUNT),
    tokenSources: [{ chainId: BigInt(INPUT_TOKEN_CHAIN_ID) }]
  };
  const output = {
    standardizedTokenId: standardizedTokenIds[standardizedTokenIds.length - 1],
    tokenDestination: { chainId: BigInt(OUTPUT_TOKEN_CHAIN_ID) }
  };
  const gasToken = { standardizedTokenId: standardizedTokenIds[0] };
  const swapResponse = await virtualNodeProvider.getOperationsToSwap(
    accountClusterId,
    QuoteType.EXACT_INPUT,
    input,
    output,
    gasToken
  );
  if (!swapResponse) {
    throw new Error("[ERROR] failed to get operations to swap");
  }

  // Display structured response information
  console.log("\n[INFO] Swap Operations Response:");
  console.log(`         Status: ${swapResponse.status}`);
  console.log(`         Estimated Time: ${swapResponse.aggregateEstimatedTimeInMs}`);
  console.log(`         Number of Operations: ${swapResponse.intents?.length ?? 0}`);

  const allOperations = swapResponse.intents
    ?.map((intent) => intent.intentOperations)
    .flat()
    ?.concat(swapResponse?.primaryOperation)
    .filter(notEmpty) ?? [];

  const promises = allOperations.map(async (operation, i) => {
      console.log(`\n         Operation: ${i}`);
      console.log(`         Type: ${operation.type}`);
      console.log(`         Format: ${operation.format}`);
      console.log(`         From: ${operation.from}`);
      console.log(`         To: ${operation.to}`);
      console.log(`         Chain ID: ${operation.chainId}`);
      console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

      // Sign the operations
      const signature = await signOperation(wallet, operation);
      return { 
        type: operation.type, 
        signature, 
        data: operation.data, 
        chainId: getOrbyChainId(operation.chainId), 
        from: operation.from
      };
  });

  const signedOperations = (await Promise.all(promises)).filter(notEmpty);
  
  if (signedOperations) {
    console.log("\n[INFO] finished signing transactions");

    const finalOperation = allOperations.find(
      (operation) => operation.type == OperationType.FINAL_TRANSACTION
    );
    if (!finalOperation) {
      throw new Error("\n[INFO] failed to find final transaction");
    }
    const finalTxOrbyProvider: Client<
        HttpTransport,
        undefined,
        undefined,
        PublicRpcSchema,
        OrbyActions
    > = createClient({
    transport: http(finalOperation.txRpcUrl),
    }).extend(orbyActions);

    console.log(`\n[INFO] sending transactions to ${finalOperation.txRpcUrl}`);

    const { operationResponses } =
    signedOperations.length > 0
      ? await finalTxOrbyProvider.sendSignedOperations(accountClusterId, signedOperations)
      : { operationResponses: [] };

    console.log(`[INFO] successfully submited transactions`, operationResponses);
  }
}

async function signOperation(account: PrivateKeyAccount, operation: any): Promise<string> {
  const { from, to, data, gasPrice, maxPriorityFeePerGas, maxFeePerGas, format, gasLimit, value, nonce, chainId } =
    operation;
  
  createWalletClient
  if (format?.trim() == OperationDataFormat.TRANSACTION) {
    const txData = {
      from,
      to,
      value,
      input: data,
      nonce: Number(nonce),
      gas: gasLimit,
      chainId,
      gasPrice,
      maxFeePerGas,
      maxPriorityFeePerGas
    };
    
    return await account.signTransaction(txData);
  } else {
    const parsedData = JSON.parse(data) as {
      domain: TypedDataDomain;
      types: Record<string, Array<TypedDataParameter>>;
      message: Record<string, any>;
      primaryType: string;
    };

    const eip712Domain = [
      { name: "name", type: "string" },
      { name: "chainId", type: "uint256" },
      { name: "verifyingContract", type: "address" },
    ];
    parsedData.types["EIP712Domain"] = eip712Domain;

    return await account.signTypedData(parsedData);
  }
}

// Run the application
main().catch((err) => console.error("[ERROR] ", err));