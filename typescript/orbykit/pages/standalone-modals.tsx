import {
  OrbyKitProvider,
  useTransaction,
  OrbyKitConfig,
  useOrbyKit,
  AccountModal,
  TransactionModal,
  useModal,
} from "@orb-labs/orbykit";
import { useEffect, useState } from "react";
import { http, createConfig, WagmiProvider } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { mainnet, base, optimism, arbitrum, polygon } from "wagmi/chains";
import { DynamicWagmiConnector } from "@dynamic-labs/wagmi-connector";
import {
  DynamicContextProvider,
  DynamicConnectButton,
} from "@dynamic-labs/sdk-react-core";
import { EthereumWalletConnectors } from "@dynamic-labs/ethereum";
import {
  BlockchainEnvironment,
  OperationDataFormat,
  OperationType,
} from "@orb-labs/orby-core";

const config = createConfig({
  chains: [mainnet, base, polygon, optimism, arbitrum],
  multiInjectedProviderDiscovery: false,
  transports: {
    [mainnet.id]: http(),
    [base.id]: http(),
    [optimism.id]: http(),
    [arbitrum.id]: http(),
    [polygon.id]: http(),
  },
});
const queryClient = new QueryClient();

function StandaloneModalsExample() {
  const orbyKitConfig: OrbyKitConfig = {
    instancePrivateAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
    instancePublicAPIKey: "f1c1d996-8df4-4d23-b926-ca702173021d",
    environment: BlockchainEnvironment.MAINNET,
    appName: "Example",
  };

  return (
    <OrbyKitProvider config={orbyKitConfig}>
      <TransactionView />
    </OrbyKitProvider>
  );
}

function TransactionView({
  // NOTE: ask imti for funds back if you need them
  toAddress = "0xe00180243cdC909208298aB8433785E8a31F287b",
}: {
  toAddress?: string;
}) {
  const {
    openAccountModal,
    isAccountModalOpen,
    closeAccountModal,
    isTransactionModalOpen,
    closeTransactionModal,
  } = useModal();
  const amountToTransfer = 0.00005;
  const { network, account, isConnected, loading } = useOrbyKit();
  const { previewTransaction, updateTransactionDetails } = useTransaction();

  useEffect(() => {
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

      const onSuccess = (isSuccessful, txHash, error) => {
        console.log("onSuccess", isSuccessful, txHash, error);
      };

      updateTransactionDetails(tx, onSuccess);
    }
  }, [network.chain, account, loading, toAddress]);

  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const handleClick = () => {
    previewTransaction();
  };

  return (
    <div className="w-full flex flex-col items-center justify-center gap-2">
      <h1 className="text-xl font-bold">OrbyKit using our standalone modals</h1>
      {mounted && isConnected ? (
        <>
          <button onClick={openAccountModal}>Open Account Modal</button>
          {isAccountModalOpen && (
            <div className="standalone-modals-account-modal-container">
              <AccountModal onClose={closeAccountModal} />
            </div>
          )}
          <div className="mt-4">
            <button
              className="bg-green-800 text-white px-4 py-2 rounded-full w-full cursor-pointer hover:bg-green-950"
              onClick={handleClick}
            >
              Send testnet ETH to OrbyPool
            </button>
          </div>
          {isTransactionModalOpen && (
            <div className="standalone-modals-transaction-modal-container">
              <TransactionModal onClose={closeTransactionModal} />
            </div>
          )}
        </>
      ) : (
        <DynamicConnectButton>Connect Wallet</DynamicConnectButton>
      )}
    </div>
  );
}

export default function StandaloneModalsOrbyKit() {
  return (
    <DynamicContextProvider
      settings={{
        environmentId: "b7c670c1-fb25-405d-b390-9fb7a60282c1",
        walletConnectors: [EthereumWalletConnectors],
      }}
    >
      <QueryClientProvider client={queryClient}>
        <WagmiProvider config={config}>
          <DynamicWagmiConnector>
            <StandaloneModalsExample />
          </DynamicWagmiConnector>
        </WagmiProvider>
      </QueryClientProvider>
    </DynamicContextProvider>
  );
}
