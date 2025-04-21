import { ReactNode, useEffect, useState } from "react";
import {
  ConnectButton,
  useTransaction,
  lightTheme,
  OrbyKitProvider,
  OrbyKitConfig,
  useOrbyKit,
} from "@orb-labs/orbykit";
import { createConfig, http, WagmiProvider, useSimulateContract } from "wagmi";
import { erc20Abi, encodeFunctionData } from "viem";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { holesky, sepolia, optimismSepolia, baseSepolia } from "wagmi/chains";
import {
  BlockchainEnvironment,
  OperationDataFormat,
  OperationType,
} from "@orb-labs/orby-core";

const config = createConfig({
  chains: [holesky, sepolia, optimismSepolia, baseSepolia],
  transports: {
    [holesky.id]: http(),
    [sepolia.id]: http(),
    [optimismSepolia.id]: http(),
    [baseSepolia.id]: http(),
  },
});

const queryClient = new QueryClient();

const orbyKitConfig: OrbyKitConfig = {
  instancePrivateAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  instancePublicAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
  appName: "Example",
  environment: BlockchainEnvironment.TESTNET,
};

function Providers({ children }: { children: ReactNode }) {
  return (
    <QueryClientProvider client={queryClient}>
      <WagmiProvider config={config}>
        <OrbyKitProvider
          config={orbyKitConfig}
          theme={lightTheme({
            accentColor: "#9933FF",
            borderRadius: "small",
            fontStack: "rounded",
          })}
        >
          {children}
        </OrbyKitProvider>
      </WagmiProvider>
    </QueryClientProvider>
  );
}

function VanillaNativeTokenExample({
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
      data: "",
      to: toAddress,
      format: OperationDataFormat.TRANSACTION,
      from: account.address,
      txRpcUrl: "",
      type: OperationType.FINAL_TRANSACTION,
      value: BigInt(amountToTransfer * 10 ** 18),
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
      <h1 className="text-xl font-bold">Vanilla OrbyKit (eth tx)</h1>
      <ConnectButton />
      {mounted && isConnected && (
        <div className="mt-4">
          <button
            className="bg-green-800 text-white px-4 py-2 rounded-full w-full cursor-pointer hover:bg-green-950"
            onClick={handleClick}
          >
            Send testnet ETH to OrbyPool
          </button>
        </div>
      )}
    </div>
  );
}

function VanillaFungibleTokenExample({
  // NOTE: ask imti for funds back if you need them
  toAddress = "0xe00180243cdC909208298aB8433785E8a31F287b",
}: {
  toAddress?: string;
}) {
  console.log("foo");
  const { network, account, isConnected, loading } = useOrbyKit();
  const { previewTransaction, updateTransactionDetails } = useTransaction();

  const result = useSimulateContract({
    abi: erc20Abi,
    address: "0x036CbD53842c5426634e7929541eC2318f3dCF7e",
    functionName: "transfer",
    args: ["0xe00180243cdC909208298aB8433785E8a31F287b", BigInt(1000000)],
  });

  console.log("result", result);

  if (network.chain && account && result?.data?.request) {
    console.log("result.data.request", result.data.request);
    const data = encodeFunctionData(result.data.request);
    console.log("data", data);

    // TODO(imti): check why this isn't typed, I can add arbitrary fields and it doesn't complain

    const tx = {
      chainId: BigInt(network.chain.id),
      to: toAddress,
      data: data,
      format: OperationDataFormat.TRANSACTION,
      txRpcUrl: "",
      type: OperationType.FINAL_TRANSACTION,
    };

    updateTransactionDetails(tx);
  }

  //const data = inputTokenContract?.interface.encodeFunctionData("approve", [protocolAddress, quote.inputAmount]);

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
      <h1 className="text-xl font-bold">Vanilla OrbyKit (USDC tx)</h1>
      <ConnectButton />
      {mounted && isConnected && (
        <div className="mt-4">
          <button
            className="bg-green-800 text-white px-4 py-2 rounded-full w-full cursor-pointer hover:bg-green-950"
            onClick={handleClick}
          >
            Send 1 USDC to {formattedToAddress}
          </button>
        </div>
      )}
    </div>
  );
}

export default function VanillaOrbyKit() {
  return (
    <Providers>
      <VanillaNativeTokenExample />
      {/* <VanillaFungibleTokenExample /> */}
    </Providers>
  );
}
