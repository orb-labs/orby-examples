const { Account } = require("@orb-labs/orby-core");
const { ethers } = require("ethers");
const { OrbyProvider } = require("@orb-labs/orby-ethers6");
const { GetOperationsToSwap } = require("./orby_functions/get_operations_to_swap");
const { GetOperationsToExecuteTransaction } = require("./orby_functions/get_operations_to_execute_transaction");
const { GetOperationsToSignTypedData } = require("./orby_functions/get_operations_to_sign_typed_data");
const { GetFungibleTokenPortfolio } = require("./orby_functions/get_fungible_token_portfolio");
const { PRIVATE_KEY, ORBY_ENGINE_ADMIN_URL, ORBY_INSTANCE_NAME, INPUT_TOKEN_CHAIN_ID, EXAMPLE_TYPE } = require("./config");

if (!PRIVATE_KEY) {
  throw new Error("Missing PRIVATE_KEY in .env file");
}
if (!ORBY_ENGINE_ADMIN_URL) {
  throw new Error("Missing ORBY_ENGINE_ADMIN_URL in .env file");
}
if (!ORBY_INSTANCE_NAME) {
  throw new Error("Missing ORBY_INSTANCE_NAME in .env file");
}
if (!INPUT_TOKEN_CHAIN_ID) {
  throw new Error("Missing INPUT_TOKEN_CHAIN_ID in .env file");
}
if (!EXAMPLE_TYPE) {
  throw new Error("Missing EXAMPLE_TYPE in .env file");
}

async function main() {
  // 1. Setup virtual node provider and account cluster
  const { virtualNodeProvider, accountCluster } = await setup();

  // 2. Run example
  let example;

  if (EXAMPLE_TYPE == "getOperationsToSwap") {
    example = new GetOperationsToSwap(virtualNodeProvider, accountCluster);
  } else if (EXAMPLE_TYPE == "getOperationsToExecuteTransaction") {
    example = new GetOperationsToExecuteTransaction(virtualNodeProvider, accountCluster);
  } else if (EXAMPLE_TYPE == "getOperationsToSignTypedData") {
    example = new GetOperationsToSignTypedData(virtualNodeProvider, accountCluster);
  } else if (EXAMPLE_TYPE == "getFungibleTokenPortfolio") {
    example = new GetFungibleTokenPortfolio(virtualNodeProvider, accountCluster);
  } else {
    throw new Error(`invalid example type: ${EXAMPLE_TYPE}`)
  }

  await example.run();
}

async function setup() {
  // ******************************** Creating a private instance using orby engine admin ********************************

  // 1. Connect to orby engine admin
  const orbyAdminProvider = new OrbyProvider(ORBY_ENGINE_ADMIN_URL);

  // 2. Create a private instance
  console.log(`[INFO] creating orby instance using name: ${ORBY_INSTANCE_NAME}`);
  const { success, orbyInstancePrivateUrl } = await orbyAdminProvider.createInstance(ORBY_INSTANCE_NAME);
  if (!success || !orbyInstancePrivateUrl) {
    throw new Error("failed to create instance");
  } else {
    console.log(`[INFO] successfully created instance`);
  }

  // 3. Connect to the private instance
  console.log("[INFO] creating private instance using url: ", orbyInstancePrivateUrl);
  const privateInstanceProvider = new OrbyProvider(orbyInstancePrivateUrl);

  // ********************************** Use private instance to create account cluster ***********************************

  // 4. Get address from private key
  const wallet = new ethers.Wallet(PRIVATE_KEY);
  const address = wallet.address;
  console.log("\n[INFO] derived address from private key: ", address)

  // 5. Create account cluster
  console.log("\n[INFO] creating account cluster...");
  const accounts = [Account.toAccount({ vmType: "EVM", address, accountType: "EOA" })];
  const accountCluster = await privateInstanceProvider.createAccountCluster(accounts);
  if (!accountCluster) {
    throw new Error("failed to create account cluster");
  }
  const accountClusterId = accountCluster.accountClusterId;

  // ******************************* Create virtual node to interact with account cluster ********************************

  // 6. Get virtual node RPC URL
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

  // 7. Create virtual node client
  const virtualNodeProvider = new OrbyProvider(virtualNodeRpcUrl);

  return { virtualNodeProvider, accountCluster };
}

// async function getSwapParams(virtualNodeProvider) {
//   // 1. Get standardized token id's for the desired input/output tokens
//   const tokens = [
//     { chainId: BigInt(INPUT_TOKEN_CHAIN_ID), tokenAddress: INPUT_TOKEN_ADDRESS },
//     { chainId: BigInt(OUTPUT_TOKEN_CHAIN_ID), tokenAddress: OUTPUT_TOKEN_ADDRESS },
//   ]

//   console.log("\n[INFO] getting standardized token IDs for:");
//   tokens.forEach((token, i) => {
//       console.log(`         Token ${i + 1}:`);
//       console.log(`           Chain ID: ${token.chainId}`);
//       console.log(`           Address: ${token.tokenAddress}`);
//   });

//   const standardizedTokenIds = await virtualNodeProvider.getStandardizedTokenIds(tokens);
//   if (!standardizedTokenIds) {
//     throw new Error("failed to get standardized token IDs");
//   } else {
//     console.log(`[INFO] standardized token ids: ${standardizedTokenIds}`);
//   }

//   // 2. Format input/output/gasToken for getOperationsToSwap
//   const input = { 
//     standardizedTokenId: standardizedTokenIds[0], 
//     amount: BigInt(AMOUNT),  // Note: set input amount for EXACT_INPUT quotes
//     tokenSources: [{ chainId: BigInt(INPUT_TOKEN_CHAIN_ID) }]
//   };
//   const output = { 
//     standardizedTokenId: standardizedTokenIds[standardizedTokenIds.length - 1],
//     amount: undefined,  // Note: set output amount for EXACT_OUTPUT quotes
//     tokenDestination: { chainId: BigInt(OUTPUT_TOKEN_CHAIN_ID) }
//   };
//   const gasToken = { standardizedTokenId: standardizedTokenIds[0] };

//   console.log(`\n[INFO] getSwapParams result`);
//   console.log(`         Input Token:`);
//   console.log(`           Standardized Token ID: ${input.standardizedTokenId}`);
//   console.log(`           Amount: ${input.amount}`);
//   console.log(`           Token Sources (Chain ID): ${input.tokenSources[0].chainId}`);
//   console.log(`         Output Token:`);
//   console.log(`           Standardized Token ID: ${output.standardizedTokenId}`);
//   console.log(`           Token Destination (Chain ID): ${output.tokenDestination.chainId}`);
//   console.log(`         Gas Token:`);
//   console.log(`           Standardized Token ID: ${gasToken.standardizedTokenId}`);

//   return { input, output, gasToken };
// }

// // TO-DO: create example that signs non-typed data
// async function signTransaction(operation) {
//   const wallet = new ethers.Wallet(PRIVATE_KEY);

//   console.log(`\n[INFO] Signing operation...`);
//   console.log(`         Type: ${operation.type}`);
//   console.log(`         Format: ${operation.format}`);
//   console.log(`         From: ${operation.from}`);
//   console.log(`         To: ${operation.to}`);
//   console.log(`         Chain ID: ${operation.chainId}`);
//   console.log(`         TX RPC URL: ${operation.txRpcUrl}`);
  
//   const { from, to, data, gasPrice, maxPriorityFeePerGas, maxFeePerGas, gasLimit, value, nonce, chainId } = operation;
      
//   const txData = {
//     from,
//     to,
//     value,
//     data,
//     nonce: Number(nonce),
//     gasLimit,
//     chainId,
//     gasPrice,
//     maxFeePerGas,
//     maxPriorityFeePerGas,
//   };

//   const tx = await wallet.populateTransaction(txData);
//   return await wallet.signTransaction(tx);
// }

// async function signTypedData(operation) {
//   const wallet = new ethers.Wallet(PRIVATE_KEY);

//   console.log(`\n[INFO] Signing operation...`);
//   console.log(`         Type: ${operation.type}`);
//   console.log(`         Format: ${operation.format}`);
//   console.log(`         From: ${operation.from}`);
//   console.log(`         To: ${operation.to}`);
//   console.log(`         Chain ID: ${operation.chainId}`);
//   console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

//   const parsedData = JSON.parse(operation.data);
//   delete parsedData.types['EIP712Domain'];

//   return await wallet.signTypedData(parsedData.domain, parsedData.types, parsedData.message);
// }

main().catch((err) => console.error("[ERROR] ", err));