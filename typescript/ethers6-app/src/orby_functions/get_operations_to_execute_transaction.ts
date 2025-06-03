import { OrbyProvider } from "@orb-labs/orby-ethers6";
import { AccountCluster, CreateOperationsStatus } from "@orb-labs/orby-core";
import {
  signTransaction,
  signTypedData,
  iface,
  onOperationSetStatusUpdateCallback,
} from "../utils";

export class GetOperationsToExecuteTransaction {
  private virtualNodeProvider: OrbyProvider;
  private accountCluster: AccountCluster;
  private inputTokenAddress: string;
  private amount: bigint;
  private destination: string;

  constructor(
    virtualNodeProvider: OrbyProvider,
    accountCluster: AccountCluster,
    inputTokenAddress: string,
    amount: bigint,
    destination: string
  ) {
    this.virtualNodeProvider = virtualNodeProvider;
    this.accountCluster = accountCluster;
    this.inputTokenAddress = inputTokenAddress;
    this.amount = amount;
    this.destination = destination;
  }

  async run() {
    // 1. Create the data for the transaction
    const data = iface.encodeFunctionData("transfer", [
      this.destination,
      this.amount,
    ]);

    // 2. Call getOperationsToExecuteTransaction
    console.log("\n[INFO] calling getOperationsToExecuteTransaction...");
    const response =
      await this.virtualNodeProvider.getOperationsToExecuteTransaction(
        this.accountCluster.accountClusterId, // accountClusterId
        data, // data
        this.inputTokenAddress // to
      );

    console.log("\n", response);

    if (!response || response.status != CreateOperationsStatus.SUCCESS) {
      throw new Error("failed to get operations to execute transaction");
    }

    console.log("\n[INFO] Operations Response:");
    console.log(`         Status: ${response.status}`);
    console.log(
      `         Estimated Time: ${response.aggregateEstimatedTimeInMs}`
    );
    console.log(
      `         Number of Operations: ${response.intents?.length ?? 0}`
    );

    // 3. Sign and send the operations
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
}
