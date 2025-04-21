import { ReactNode, useEffect, useState } from "react";
import {
  ConnectedWallet,
  OrbyKitProvider,
  useTransaction,
  useOrbyKit,
  OrbyKitConfig,
} from "@orb-labs/orbykit";
import { PrivyProvider, usePrivy } from "@privy-io/react-auth";
import { http } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  holesky,
  sepolia,
  optimismSepolia,
  baseSepolia,
  arbitrumSepolia,
} from "wagmi/chains";

import { createConfig, WagmiProvider } from "@privy-io/wagmi";
import {
  Account,
  AccountType,
  BlockchainEnvironment,
  OperationDataFormat,
  OperationType,
  VMType,
} from "@orb-labs/orby-core";

const config = createConfig({
  chains: [holesky, sepolia, optimismSepolia, baseSepolia, arbitrumSepolia],
  transports: {
    [holesky.id]: http(),
    [sepolia.id]: http(),
    [optimismSepolia.id]: http(),
    [baseSepolia.id]: http(),
    [arbitrumSepolia.id]: http(),
  },
});
const queryClient = new QueryClient();

const PRIVY_TEST_APP_ID = "cltf9s2kz0goyk10zy60ysikl";

const orbyKitConfig: OrbyKitConfig = {
  instancePrivateAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  instancePublicAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  environment: BlockchainEnvironment.TESTNET,
  appName: "Example",
};

const Providers = ({ children }: { children: ReactNode }) => {
  return (
    <PrivyProvider
      appId={PRIVY_TEST_APP_ID}
      config={{
        loginMethods: ["email", "wallet"],
        appearance: {
          theme: "light",
          accentColor: "#676FFF",
          logo: "https://your-logo-url",
        },
        defaultChain: sepolia,
        supportedChains: [holesky, sepolia, optimismSepolia, baseSepolia],
      }}
    >
      <QueryClientProvider client={queryClient}>
        <WagmiProvider config={config}>
          <OrbyKitProvider config={orbyKitConfig}>{children}</OrbyKitProvider>
        </WagmiProvider>
      </QueryClientProvider>
    </PrivyProvider>
  );
};

function PrivyExample({
  // NOTE: ask imti for funds back if you need them
  toAddress = "0xe00180243cdC909208298aB8433785E8a31F287b",
}: {
  toAddress?: string;
}) {
  const amountToTransfer = 0.001;
  const { login } = usePrivy();
  const { previewTransaction, updateTransactionDetails } = useTransaction();
  const { network, account, isConnected, loading } = useOrbyKit();

  if (network.chain && account && !loading) {
    const tx = {
      chainId: BigInt(network.chain.id),
      to: toAddress,
      value: BigInt(amountToTransfer * 10 ** 18),
      data: "",
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
      <h1 className="text-xl font-bold">Privy + OrbyKit</h1>
      {isConnected && mounted ? (
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
        <button className="cursor-pointer" onClick={login}>
          Connect Wallet
        </button>
      )}
    </div>
  );
}

export default function PrivyOrbyKit() {
  return (
    <Providers>
      <PrivyExample />
    </Providers>
  );
}
