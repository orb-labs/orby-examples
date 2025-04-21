import { Client, HttpTransport, PublicRpcSchema } from 'viem';
import { OrbyActions } from '@orb-labs/orby-viem-extension';
import { AccountCluster, CreateOperationsStatus, QuoteType } from "@orb-labs/orby-core";
import { INPUT_TOKEN_CHAIN_ID, INPUT_TOKEN_ADDRESS, OUTPUT_TOKEN_CHAIN_ID, OUTPUT_TOKEN_ADDRESS, AMOUNT } from "..";
import { signTransaction, signTypedData } from "../utils";

export class GetOperationsToSwap {
  private virtualNodeProvider: Client<HttpTransport, undefined, undefined, PublicRpcSchema, OrbyActions>;
  private accountCluster: AccountCluster;

  constructor(
    virtualNodeProvider: Client<HttpTransport, undefined, undefined, PublicRpcSchema, OrbyActions>, 
    accountCluster: AccountCluster
  ) {
    this.virtualNodeProvider = virtualNodeProvider;
    this.accountCluster = accountCluster;
  }

  async run() {
    // 0. Check for env variables
    if (!INPUT_TOKEN_ADDRESS) {
      throw new Error("INPUT_TOKEN_ADDRESS must be set in .env file for GetOperationsToSwap");
    }
    if (!OUTPUT_TOKEN_ADDRESS) {
      throw new Error("OUTPUT_TOKEN_ADDRESS must be set in .env file for GetOperationsToSwap");
    }
    if (!INPUT_TOKEN_CHAIN_ID) {
      throw new Error("INPUT_TOKEN_CHAIN_ID must be set in .env file for GetOperationsToSwap");
    }
    if (!OUTPUT_TOKEN_CHAIN_ID) {
      throw new Error("OUTPUT_TOKEN_CHAIN_ID must be set in .env file for GetOperationsToSwap");
    }
    if (!AMOUNT) {
      throw new Error("AMOUNT must be set in .env file for GetOperationsToSwap");
    }

    // 1. Format operation request
    const { input, output, gasToken } = await this.getParams();
    
    // 2. Call operation
    console.log("\n[INFO] calling getOperationsToSwap...");
    const swapResponse = await this.virtualNodeProvider.getOperationsToSwap(
      this.accountCluster.accountClusterId,
      QuoteType.EXACT_INPUT,
      input,
      output,
      gasToken
    );
    if (!swapResponse || swapResponse.status != CreateOperationsStatus.SUCCESS) {
      throw new Error("failed to get operations to swap");
    }

    console.log("\n[INFO] Swap Operations Response:");
    console.log(`         Status: ${swapResponse.status}`);
    console.log(`         Estimated Time: ${swapResponse.aggregateEstimatedTimeInMs}`);
    console.log(`         Number of Operations: ${swapResponse.intents?.length ?? 0}`);

    // 3. Call sendOperationSet to sign and send the operations
    console.log("\n[INFO] calling sendOperationSet...");
    const sendResponse = await this.virtualNodeProvider.sendOperationSet(
      this.accountCluster,
      swapResponse,
      signTransaction,
      undefined,  // signUserOperation
      signTypedData
    );
    if (!sendResponse) {
      throw new Error("failed to send operation set");
    }

    console.log("\n[INFO] Send Operation Set Response:");
    console.log(`         Success: ${sendResponse.success}`);
    console.log(`         Operation Set ID: ${sendResponse.operationSetId}`);
  }

  async getParams(): Promise<{
    input: any, 
    output: any, 
    gasToken: any
  }> {
    // 1. Get standardized token id's for the desired input/output tokens
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

    const standardizedTokenIds = await this.virtualNodeProvider.getStandardizedTokenIds(tokens);
    if (!standardizedTokenIds) {
      throw new Error("failed to get standardized token IDs");
    } else {
      console.log(`[INFO] standardized token ids: ${standardizedTokenIds}`);
    }

    // 2. Format input/output/gasToken for getOperationsToSwap
    const input = { 
      standardizedTokenId: standardizedTokenIds[0], 
      amount: BigInt(AMOUNT),  // Note: set input amount for EXACT_INPUT quotes
      tokenSources: [{ chainId: BigInt(INPUT_TOKEN_CHAIN_ID) }]
    };
    const output = { 
      standardizedTokenId: standardizedTokenIds[standardizedTokenIds.length - 1],
      amount: undefined,  // Note: set output amount for EXACT_OUTPUT quotes
      tokenDestination: { chainId: BigInt(OUTPUT_TOKEN_CHAIN_ID) }
    };
    const gasToken = { standardizedTokenId: standardizedTokenIds[0] };

    console.log(`\n[INFO] getSwapParams result`);
    console.log(`         Input Token:`);
    console.log(`           Standardized Token ID: ${input.standardizedTokenId}`);
    console.log(`           Amount: ${input.amount}`);
    console.log(`           Token Sources (Chain ID): ${input.tokenSources[0].chainId}`);
    console.log(`         Output Token:`);
    console.log(`           Standardized Token ID: ${output.standardizedTokenId}`);
    console.log(`           Token Destination (Chain ID): ${output.tokenDestination.chainId}`);
    console.log(`         Gas Token:`);
    console.log(`           Standardized Token ID: ${gasToken.standardizedTokenId}`);

    return { input, output, gasToken };
  }
}