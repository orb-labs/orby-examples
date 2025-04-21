import { useEffect, useState } from "react";
import {
  ConnectedWallet,
  OrbyKitProvider,
  useTransaction,
  OrbyKitConfig,
  useOrbyKit,
} from "@orb-labs/orbykit";
import { defaultWagmiConfig } from "@web3modal/wagmi/react/config";
import { cookieStorage, createStorage, WagmiProvider } from "wagmi";
import { mainnet, sepolia, baseSepolia } from "wagmi/chains";
import { createWeb3Modal } from "@web3modal/wagmi/react";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  BlockchainEnvironment,
  OperationDataFormat,
  OperationType,
} from "@orb-labs/orby-core";

// Your WalletConnect Cloud project ID
export const projectId = "364d8e862158bea61dd6f0a4ab65220e";

// Create a metadata object
const metadata = {
  name: "example",
  description: "AppKit Example",
  url: "https://web3modal.com", // origin must match your domain & subdomain
  icons: ["https://avatars.githubusercontent.com/u/37784886"],
};

// Create wagmiConfig
const chains = [mainnet, sepolia, baseSepolia] as const;
const config = defaultWagmiConfig({
  chains,
  projectId,
  metadata,
  ssr: true,
  storage: createStorage({
    storage: cookieStorage,
  }),
});

// Setup queryClient
const queryClient = new QueryClient();

if (!projectId) throw new Error("Project ID is not defined");

// Create modal
createWeb3Modal({
  wagmiConfig: config,
  projectId,
  enableAnalytics: false, // Optional - defaults to your Cloud configuration
  enableOnramp: false, // Optional - false as default
});

const orbyKitConfig: OrbyKitConfig = {
  instancePrivateAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  instancePublicAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  environment: BlockchainEnvironment.TESTNET,
  appName: "Example",
};

const Providers = ({ children }: { children: any }) => {
  return (
    <WagmiProvider config={config}>
      <QueryClientProvider client={queryClient}>
        <OrbyKitProvider config={orbyKitConfig}>{children}</OrbyKitProvider>
      </QueryClientProvider>
    </WagmiProvider>
  );
};

function ConnectButton() {
  return <w3m-button />;
}

function AppKitExample({
  // NOTE: ask imti for funds back if you need them
  toAddress = "0xe00180243cdC909208298aB8433785E8a31F287b",
}: {
  toAddress?: string;
}) {
  const amountToTransfer = 0.001;
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
      <h1 className="text-xl font-bold">AppKit + OrbyKit (WIP)</h1>
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

export default function AppKitOrbyKit() {
  return (
    <Providers>
      <AppKitExample />
    </Providers>
  );
}
