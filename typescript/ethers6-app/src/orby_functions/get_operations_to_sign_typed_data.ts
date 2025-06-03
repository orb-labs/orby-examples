import { OrbyProvider } from "@orb-labs/orby-ethers6";
import { AccountCluster, CreateOperationsStatus } from "@orb-labs/orby-core";
import { onOperationSetStatusUpdateCallback, signTransaction, signTypedData } from "../utils";

export class GetOperationsToSignTypedData {
  private virtualNodeProvider: OrbyProvider;
  private accountCluster: AccountCluster;
  private inputTokenAddress: string;
  private amount: bigint;
  private inputTokenChainId: bigint;

  constructor(
    virtualNodeProvider: OrbyProvider,
    accountCluster: AccountCluster,
    inputTokenAddress: string,
    amount: bigint,
    inputTokenChainId: bigint
  ) {
    this.virtualNodeProvider = virtualNodeProvider;
    this.accountCluster = accountCluster;
    this.inputTokenAddress = inputTokenAddress;
    this.amount = amount;
    this.inputTokenChainId = inputTokenChainId;
  }

  async run() {
    // 1. Format operation request
    const data = this.createPermit2Object();

    // 2. Call operation
    console.log("\n[INFO] calling getOperationsToSignTypedData...");
    const response = await this.virtualNodeProvider.getOperationsToSignTypedData(
      this.accountCluster.accountClusterId,
      data
    );
    if (!response || response.status != CreateOperationsStatus.SUCCESS) {
      throw new Error("failed to get operations to sign typed data");
    }

    console.log("\n[INFO] Operations Response:");
    console.log(`         Status: ${response.status}`);
    console.log(`         Estimated Time: ${response.aggregateEstimatedTimeInMs}`);
    console.log(`         Number of Operations: ${response.intents?.length ?? 0}`);

    // 3. Call sendOperationSet to sign and send the operations
    console.log("\n[INFO] calling sendOperationSet...");
    const sendResponse = await this.virtualNodeProvider.sendOperationSet(
      this.accountCluster,
      response,
      signTransaction,
      undefined, // signUserOperation
      signTypedData
    );
    if (!sendResponse) {
      throw new Error("failed to send operation set");
    }

    console.log("\n[INFO] Send Operation Set Response:");
    console.log(`         Success: ${sendResponse.success}`);
    console.log(`         Operation Set ID: ${sendResponse.operationSetId}`);

    // 4. Subscribe to operation set status updates
    this.virtualNodeProvider?.subscribeToOperationSetStatus(
      sendResponse.operationSetId,
      onOperationSetStatusUpdateCallback
    );
  }

  // Creates a Permit 2 object and returns it as a string
  createPermit2Object(): string {
    const permit2Object = {
      types: {
        PermitTransferFrom: [
          { name: "permitted", type: "TokenPermissions" },
          { name: "spender", type: "address" },
          { name: "nonce", type: "uint256" },
          { name: "deadline", type: "uint256" },
          { name: "witness", type: "ExclusiveDutchOrder" },
        ],
        TokenPermissions: [
          { name: "token", type: "address" },
          { name: "amount", type: "uint256" },
        ],
        EIP712Domain: [
          { name: "name", type: "string" },
          { name: "chainId", type: "uint256" },
          { name: "verifyingContract", type: "address" },
        ],
      },
      domain: {
        name: "Permit2",
        chainId: this.inputTokenChainId,
        verifyingContract: "0x000000000022d473030f116ddee9f6b43ac78ba3", // Uniswap Permit2 address
      },
      primaryType: "PermitTransferFrom",
      message: {
        permitted: {
          token: this.inputTokenAddress,
          amount: this.amount.toString(),
        },
        spender: this.accountCluster.accounts[0].address,
        nonce: "10",
        deadline: "100",
      },
    };

    return JSON.stringify(permit2Object);
  }
}
