import "@rainbow-me/rainbowkit/styles.css";

import { ReactNode, useEffect, useState } from "react";
import {
  ConnectedWallet,
  OrbyKitProvider,
  useTransaction,
  OrbyKitConfig,
  useOrbyKit,
} from "@orb-labs/orbykit";
import { WagmiProvider } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { holesky, sepolia, optimismSepolia, baseSepolia } from "wagmi/chains";
import {
  getDefaultConfig,
  RainbowKitProvider,
  ConnectButton,
} from "@rainbow-me/rainbowkit";
import {
  BlockchainEnvironment,
  OperationDataFormat,
  OperationType,
} from "@orb-labs/orby-core";

const config = getDefaultConfig({
  appName: "imti test",
  projectId: "fc7c6778abc12d0930b59137135d092b",
  chains: [holesky, sepolia, optimismSepolia, baseSepolia],
  ssr: true, // If your dApp uses server side rendering (SSR)
});

const queryClient = new QueryClient();

const orbyKitConfig: OrbyKitConfig = {
  instancePrivateAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  instancePublicAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  environment: BlockchainEnvironment.TESTNET,
  appName: "Example",
};

const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <QueryClientProvider client={queryClient}>
      <WagmiProvider config={config}>
        <RainbowKitProvider>
          <OrbyKitProvider config={orbyKitConfig}>{children}</OrbyKitProvider>
        </RainbowKitProvider>
      </WagmiProvider>
    </QueryClientProvider>
  );
};

function RainbowKitExample({
  // NOTE: ask imti for funds back if you need them
  toAddress = "0xe00180243cdC909208298aB8433785E8a31F287b",
}: {
  toAddress?: string;
}) {
  const amountToTransfer = 0.001;
  const { network, account, isConnected, loading } = useOrbyKit();
  const { previewTransaction, updateTransactionDetails } = useTransaction();

  if (network.chain && account && !loading) {
    const tx = {
      chainId: BigInt(network.chain.id),
      to: toAddress,
      value: BigInt(amountToTransfer * 10 ** 18),
      data: "0x",
      format: OperationDataFormat.TRANSACTION,
      txRpcUrl: "",
      type: OperationType.FINAL_TRANSACTION,
    };

    updateTransactionDetails(tx);
  }

  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const handleClick = () => {
    previewTransaction();
  };

  const formattedToAddress =
    toAddress.slice(0, 6) + "..." + toAddress.slice(-4);

  return (
    <div className="w-full flex flex-col items-center justify-center gap-2">
      <h1 className="text-xl font-bold">RainbowKit + OrbyKit</h1>
      {mounted && isConnected ? (
        <>
          <ConnectedWallet />
          <div className="mt-4">
            <button
              className="bg-green-800 text-white px-4 py-2 rounded-full w-full cursor-pointer hover:bg-green-950"
              onClick={handleClick}
            >
              Send testnet ETH to OrbyPool
            </button>
          </div>
        </>
      ) : (
        <ConnectButton />
      )}
    </div>
  );
}

export default function RainbowKitOrbyKit() {
  return (
    <Providers>
      <RainbowKitExample />
    </Providers>
  );
}
