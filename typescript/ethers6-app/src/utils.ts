import {
  Account,
  AccountCluster,
  Activity,
  ActivityStatus,
  OnchainOperation,
} from "@orb-labs/orby-core";
import { ethers, TypedDataDomain, TypedDataField } from "ethers";
import ERC20ABI from "./abi/erc20.json";
import { OrbyProvider } from "@orb-labs/orby-ethers6";

// ******************************************** for settin up *******************************************
export async function setup(
  orbyInstancePrivateUrl: string,
  chainId: bigint
): Promise<{
  virtualNodeProvider: OrbyProvider;
  accountCluster: AccountCluster;
  wallet: ethers.Wallet;
}> {
  // 1. Connect to the private instance
  const privateInstanceProvider = new OrbyProvider(orbyInstancePrivateUrl);

  // 2. Get address from private key
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY ?? "");
  const address = wallet.address;
  console.log("\n[INFO] derived address from private key: ", address);

  // 3. Create account cluster
  console.log("\n[INFO] creating account cluster...");
  const accounts = [
    Account.toAccount({ vmType: "EVM", address, accountType: "EOA" }),
  ];
  const accountCluster = await privateInstanceProvider.createAccountCluster(
    accounts
  );
  if (!accountCluster) {
    throw new Error("failed to create account cluster");
  }
  const accountClusterId = accountCluster.accountClusterId;

  // 4. Get virtual node RPC URL
  console.log("\n[INFO] getting virtual node RPC URL...");
  console.log(`[INFO]   accountClusterId: ${accountClusterId}`);
  console.log(`[INFO]   chainId: ${chainId}`);
  console.log(`[INFO]   address: ${address}`);
  const virtualNodeRpcUrl = await privateInstanceProvider.getVirtualNodeRpcUrl(
    accountClusterId,
    chainId,
    address
  );

  if (!virtualNodeRpcUrl) {
    throw new Error("failed to get virtual node RPC URL");
  } else {
    console.log(
      `[INFO] You can now use this URL to interact with the virtual node: ${virtualNodeRpcUrl}`
    );
  }

  // 5. Create virtual node client
  const virtualNodeProvider = new OrbyProvider(virtualNodeRpcUrl);

  return { virtualNodeProvider, accountCluster, wallet };
}

// ******************************************** for signing *******************************************

// TO-DO: create example that signs non-typed data
export const signTransaction = async (
  operation: OnchainOperation
): Promise<string | undefined> => {
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY ?? "");

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const {
    from,
    to,
    data,
    gasPrice,
    maxPriorityFeePerGas,
    maxFeePerGas,
    gasLimit,
    value,
    nonce,
    chainId,
  } = operation;

  const txData = {
    from,
    to,
    value,
    data,
    nonce: Number(nonce),
    gasLimit,
    chainId,
    gasPrice,
    maxFeePerGas,
    maxPriorityFeePerGas,
  };

  return await wallet.signTransaction(txData);
};

export const signTypedData = async (
  operation: OnchainOperation
): Promise<string | undefined> => {
  const wallet = new ethers.Wallet(process.env.PRIVATE_KEY ?? "");

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const parsedData = JSON.parse(operation.data) as {
    domain: TypedDataDomain;
    types: Record<string, Array<TypedDataField>>;
    message: Record<string, any>;
  };
  delete parsedData.types["EIP712Domain"];

  return await wallet.signTypedData(
    parsedData.domain,
    parsedData.types,
    parsedData.message
  );
};

// ******************************************** for getting operation set status *******************************************

export const onOperationSetStatusUpdateCallback = async (
  activity?: Activity
): Promise<void> => {
  console.log(
    "[INFO] [onOperationSetStatusUpdateCallback] response: ",
    activity
  );

  if (!activity) {
    // handle not found
  } else if (
    [ActivityStatus.SUCCESSFUL, ActivityStatus.PENDING].includes(
      activity?.overallStatus
    )
  ) {
    // handle success
  } else if (
    [ActivityStatus.FAILED, ActivityStatus.NOT_FOUND].includes(
      activity?.overallStatus
    )
  ) {
    // handle other cases
  }
};

// ****************************************************************************************************

export const iface = new ethers.Interface(ERC20ABI);
