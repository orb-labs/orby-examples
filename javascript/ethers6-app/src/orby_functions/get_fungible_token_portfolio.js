class GetFungibleTokenPortfolio {
  constructor(
    virtualNodeProvider,
    accountCluster
  ) {
    this.virtualNodeProvider = virtualNodeProvider;
    this.accountCluster = accountCluster;
  }

  async run() {
    // 1. Call operation
    console.log("\n[INFO] calling getFungibleTokenPortfolio...");
    const response = await this.virtualNodeProvider.getFungibleTokenPortfolio(
      this.accountCluster.accountClusterId
    );
    if (!response) {
      throw new Error("failed to get fungible token portfolio");
    }

    console.log("\n[INFO] Fungible Token Portfolio:");
    response.map((standardizedBalance) => {
      console.log(`\n         Standardized Token ID: ${standardizedBalance.standardizedTokenId}`);
      console.log(`         Total: ${standardizedBalance.total.toRawAmount()}`);
      console.log(`         Token Balances:`);
      standardizedBalance.tokenBalances.map((tokenBalance) => {
        console.log(`           Address: ${tokenBalance.token.address}`);
        console.log(`           Chain ID: ${tokenBalance.token.chainId}`);
        console.log(`           Raw Amount: ${tokenBalance.toRawAmount()}\n`);
      });
      console.log(`         Token Balances on Chains:`);
      standardizedBalance.tokenBalancesOnChains.map((tokenBalance) => {
        console.log(`           Address: ${tokenBalance.token.address}`);
        console.log(`           Chain ID: ${tokenBalance.token.chainId}`);
        console.log(`           Raw Amount: ${tokenBalance.toRawAmount()}\n`);
      });
    });
  }
}

module.exports = { GetFungibleTokenPortfolio };