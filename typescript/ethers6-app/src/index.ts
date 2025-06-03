import * as dotenv from "dotenv";
import { GetOperationsToSwap } from "./orby_functions/get_operations_to_swap";
import { GetOperationsToExecuteTransaction } from "./orby_functions/get_operations_to_execute_transaction";
import { GetOperationsToSignTypedData } from "./orby_functions/get_operations_to_sign_typed_data";
import { GetFungibleTokenPortfolio } from "./orby_functions/get_fungible_token_portfolio";
import { setup } from "./utils";

dotenv.config();

const {
  PRIVATE_KEY,
  ORBY_PRIVATE_INSTANCE_URL,
  INPUT_TOKEN_CHAIN_ID,
  EXAMPLE_TYPE,
  INPUT_TOKEN_ADDRESS,
  OUTPUT_TOKEN_ADDRESS,
  OUTPUT_TOKEN_CHAIN_ID,
  AMOUNT,
  DESTINATION_ACCOUNT_ADDRESS,
} = process.env;

if (!PRIVATE_KEY) {
  throw new Error("Missing PRIVATE_KEY in .env file");
} else if (!ORBY_PRIVATE_INSTANCE_URL) {
  throw new Error("Missing ORBY_ENGINE_ADMIN_URL in .env file");
} else if (!INPUT_TOKEN_CHAIN_ID) {
  throw new Error("Missing INPUT_TOKEN_CHAIN_ID in .env file");
} else if (!EXAMPLE_TYPE) {
  throw new Error("Missing EXAMPLE_TYPE in .env file");
}

async function main() {
  // 1. Setup virtual node provider and account cluster
  const { virtualNodeProvider, accountCluster } = await setup(
    ORBY_PRIVATE_INSTANCE_URL!,
    BigInt(INPUT_TOKEN_CHAIN_ID!)
  );

  // 2. Run example
  let example;

  if (EXAMPLE_TYPE == "getOperationsToSwap") {
    example = new GetOperationsToSwap(
      virtualNodeProvider,
      accountCluster,
      INPUT_TOKEN_ADDRESS!,
      BigInt(INPUT_TOKEN_CHAIN_ID!),
      OUTPUT_TOKEN_ADDRESS!,
      BigInt(OUTPUT_TOKEN_CHAIN_ID!),
      BigInt(AMOUNT!)
    );
  } else if (EXAMPLE_TYPE == "getOperationsToExecuteTransaction") {
    example = new GetOperationsToExecuteTransaction(
      virtualNodeProvider,
      accountCluster,
      INPUT_TOKEN_ADDRESS!,
      BigInt(AMOUNT!),
      DESTINATION_ACCOUNT_ADDRESS!
    );
  } else if (EXAMPLE_TYPE == "getOperationsToSignTypedData") {
    example = new GetOperationsToSignTypedData(
      virtualNodeProvider,
      accountCluster,
      INPUT_TOKEN_ADDRESS!,
      BigInt(AMOUNT!),
      BigInt(INPUT_TOKEN_CHAIN_ID!)
    );
  } else if (EXAMPLE_TYPE == "getFungibleTokenPortfolio") {
    example = new GetFungibleTokenPortfolio(
      virtualNodeProvider,
      accountCluster
    );
  } else {
    throw new Error(`invalid example type: ${EXAMPLE_TYPE}`);
  }

  await example.run();
}

main().catch((err) => console.error("[ERROR] ", err));
