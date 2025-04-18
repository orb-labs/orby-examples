const { CreateOperationsStatus } = require("@orb-labs/orby-core");
const { INPUT_TOKEN_ADDRESS, AMOUNT } = require("../config");
const { signTransaction, signTypedData, iface } = require("../constants");

class GetOperationsToExecuteTransaction {
  constructor(
    virtualNodeProvider,
    accountCluster
  ) {
    this.virtualNodeProvider = virtualNodeProvider;
    this.accountCluster = accountCluster;
  }

  async run() {
    // 0. Check for env variables
    if (!INPUT_TOKEN_ADDRESS) {
      throw new Error("INPUT_TOKEN_ADDRESS must be set in .env file for GetOperationsToExecuteTransaction");
    }
    if (!AMOUNT) {
      throw new Error("AMOUNT must be set in .env file for GetOperationsToExecuteTransaction");
    }

    // 1. Format operation request
    const to = INPUT_TOKEN_ADDRESS;
    const data = iface.encodeFunctionData("transfer", [
      this.accountCluster.accounts[0].address,  // recipient address
      AMOUNT
    ]);
    
    // 2. Call operation
    console.log("\n[INFO] calling getOperationsToExecuteTransaction...");
    const response = await this.virtualNodeProvider.getOperationsToExecuteTransaction(
      this.accountCluster.accountClusterId,
      data,
      to
    );
    if (!response || response.status != CreateOperationsStatus.SUCCESS) {
      throw new Error("failed to get operations to execute transaction");
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
}

module.exports = { GetOperationsToExecuteTransaction };