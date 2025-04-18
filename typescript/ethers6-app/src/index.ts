import { Account, AccountCluster } from "@orb-labs/orby-core";
import { ethers } from "ethers";
import { OrbyProvider } from "@orb-labs/orby-ethers6";
import * as dotenv from "dotenv";
import { GetOperationsToSwap } from "./orby_functions/get_operations_to_swap";
import { GetOperationsToExecuteTransaction } from "./orby_functions/get_operations_to_execute_transaction";
import { GetOperationsToSignTypedData } from "./orby_functions/get_operations_to_sign_typed_data";
import { GetFungibleTokenPortfolio } from "./orby_functions/get_fungible_token_portfolio";

dotenv.config();

// Load params from environment variables
export const PRIVATE_KEY = process.env.PRIVATE_KEY ?? "";
export const ORBY_ENGINE_ADMIN_URL = process.env.ORBY_ENGINE_ADMIN_URL ?? "";
export const ORBY_INSTANCE_NAME = process.env.ORBY_INSTANCE_NAME ?? "";
export const INPUT_TOKEN_ADDRESS = process.env.INPUT_TOKEN_ADDRESS ?? "";
export const OUTPUT_TOKEN_ADDRESS = process.env.OUTPUT_TOKEN_ADDRESS ?? "";
export const INPUT_TOKEN_CHAIN_ID = process.env.INPUT_TOKEN_CHAIN_ID ?? "";
export const OUTPUT_TOKEN_CHAIN_ID = process.env.OUTPUT_TOKEN_CHAIN_ID ?? "";
export const AMOUNT = process.env.AMOUNT ?? "0";
export const EXAMPLE_TYPE = process.env.EXAMPLE_TYPE ?? "";

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

async function setup(): Promise<{
  virtualNodeProvider: OrbyProvider;
  accountCluster: AccountCluster
}> {
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

main().catch((err) => console.error("[ERROR] ", err));