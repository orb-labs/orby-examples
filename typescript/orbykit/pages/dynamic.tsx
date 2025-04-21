import { useCallback, useEffect, useMemo, useState } from "react";
import {
  ConnectedWallet,
  OrbyKitProvider,
  useTransaction,
  OrbyKitConfig,
  useOrbyKit,
} from "@orb-labs/orbykit";
import { http, createConfig, WagmiProvider, Config } from "wagmi";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import {
  mainnet,
  base,
  optimism,
  arbitrum,
  polygon,
} from "wagmi/chains";
import { OrbyActions } from "@orb-labs/orby-viem-extension";

import { DynamicWagmiConnector } from "@dynamic-labs/wagmi-connector";
import {
  DynamicContextProvider,
  DynamicConnectButton,
  useDynamicContext,
} from "@dynamic-labs/sdk-react-core";
import { EthereumWalletConnectors } from "@dynamic-labs/ethereum";
import {
  Account,
  AccountCluster,
  AccountType,
  BlockchainEnvironment,
  Category,
  OnchainOperation,
  OperationDataFormat,
  OperationStatus,
  OperationStatusType,
  OperationType,
  SignedOperation,
  VMType,
} from "@orb-labs/orby-core";
import {
  Client,
  HttpTransport,
  PublicRpcSchema,
  TypedData,
  TypedDataDomain,
} from "viem";
import _ from "lodash";
import {
  sendTransaction,
  signTypedData,
  SendTransactionParameters,
} from "wagmi/actions";
import React from "react";

const config = createConfig({
  chains: [mainnet, base, polygon, optimism, arbitrum],
  // chains: [
  //   holesky,
  //   sepolia,
  //   optimismSepolia,
  //   baseSepolia,
  //   arbitrumSepolia,
  //   polygonAmoy,
  // ],
  multiInjectedProviderDiscovery: false,
  transports: {
    [mainnet.id]: http(),
    [base.id]: http(),
    [optimism.id]: http(),
    [arbitrum.id]: http(),
    [polygon.id]: http(),
  },
  // transports: {
  //   [holesky.id]: http(),
  //   [sepolia.id]: http(),
  //   [optimismSepolia.id]: http(),
  //   [baseSepolia.id]: http(),
  //   [arbitrumSepolia.id]: http(),
  //   [polygonAmoy.id]: http(),
  // },
});
const queryClient = new QueryClient();

function DynamicExample() {
  const { primaryWallet } = useDynamicContext();

  const signerConfigs = useMemo(() => {
    const addressToConfigMap = new Map<string, Config>();
    // TODO: ADD THE CONFIGS FOR THE WALLET ADDRESSES YOU WANT TO USE

    return addressToConfigMap;
  }, []);

  const sendOperations = useCallback(
    async (
      _accountCluster: AccountCluster,
      operations: {
        operation: OnchainOperation;
        virtualNode: Client<
          HttpTransport,
          undefined,
          undefined,
          PublicRpcSchema,
          OrbyActions
        >;
      }[]
    ): Promise<OperationStatus[]> => {
      const signatures = operations.map(async ({ operation }) => {
        if (operation.format == OperationDataFormat.TYPED_DATA) {
          if (!operation.from) {
            throw new Error("From address is required for signing operations");
          } else if (!signerConfigs.has(operation.from)) {
            throw new Error("Wagmi config is required for signing operations");
          }

          const wagmiConfig = signerConfigs.get(operation.from)!;

          const txHash = await sendTransaction(wagmiConfig, {
            nonce: Number(operation.nonce),
            chainId: Number(operation.chainId),
            to: operation.to as `0x${string}`,
            data: operation.data as `0x${string}`,
            value: operation.value,
          } as SendTransactionParameters<Config, Config["chains"][number]["id"]>);

          return {
            hash: txHash,
            id: txHash,
            status: OperationStatusType.PENDING,
            type: OperationType.FINAL_TRANSACTION,
          } as OperationStatus;
        }

        return undefined;
      });

      return await Promise.all(signatures).then((signatures) => {
        return signatures.filter((signature) => signature != undefined);
      });
    },
    [signerConfigs]
  );

  const signOperations = useCallback(
    async (
      _accountCluster: AccountCluster,
      operations: {
        operation: OnchainOperation;
        virtualNode: Client<
          HttpTransport,
          undefined,
          undefined,
          PublicRpcSchema,
          OrbyActions
        >;
      }[]
    ): Promise<SignedOperation[]> => {
      const signatures = operations.map(async ({ operation }) => {
        if (operation.format == OperationDataFormat.TYPED_DATA) {
          const parsedData = JSON.parse(operation.data) as {
            domain: TypedDataDomain;
            types: Record<string, Array<TypedData>>;
            message: Record<string, any>;
            primaryType: string;
          };

          if (!operation.from) {
            throw new Error("From address is required for signing operations");
          } else if (!signerConfigs.has(operation.from)) {
            throw new Error("Wagmi config is required for signing operations");
          }

          const wagmiConfig = signerConfigs.get(operation.from)!;
          const signature = await signTypedData(wagmiConfig, parsedData);

          return {
            type: OperationType.FINAL_TRANSACTION,
            signature,
            category: Category.TYPED_DATA_SIGNATURE,
            data: operation.data,
          };
        }

        return undefined;
      });

      return await Promise.all(signatures).then((signatures) => {
        return signatures.filter((signature) => signature != undefined);
      });
    },
    [signerConfigs]
  );

  const accounts = useMemo(() => {
    if (!primaryWallet) {
      return undefined;
    }

    return [
      // NOTE: user connected wallet
      new Account(
        primaryWallet?.address,
        AccountType.EOA,
        VMType.EVM,
        undefined
      ),

      // NOTE: dev created wallet
      // new Account(
      //   "0xe00180243cdC909208298aB8433785E8a31F287b",
      //   AccountType.EOA,
      //   VMType.EVM,
      //   undefined
      // ),

      // TODO: add more accounts here that you want to use but make sure
      // add more accounts here that you want to use but make sure
      // you have the keys for signing operations for the accounts
    ];
  }, [primaryWallet]);

  const transactingAccount = primaryWallet
    ? new Account(
        primaryWallet?.address,
        AccountType.EOA,
        VMType.EVM,
        undefined
      )
    : undefined;

  const orbyKitConfig: OrbyKitConfig = {
    instancePrivateAPIKey: process.env.ORBY_PRIVATE_API_KEY ?? "",
    instancePublicAPIKey: process.env.ORBY_PUBLIC_API_KEY ?? "",
    transactingAccount,
    environment: BlockchainEnvironment.MAINNET,
    appName: "Example",
    accounts,
    sendOperations,
    signOperations,
  };

  // if (!accounts) return <div>loading...</div>;

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
  const { primaryWallet } = useDynamicContext();

  const amountToTransfer = 0.00005;
  const { network, account, isConnected, loading, updateAppName } =
    useOrbyKit();
  const { previewTransaction, updateTransactionDetails } = useTransaction();

  useEffect(() => {
    if (network.chain && account && !loading) {
      const data = {};

      const tx = {
        chainId: BigInt(network.chain.id),
        to: toAddress,
        value: BigInt(amountToTransfer * 10 ** 18),
        // NOTE(imti): for typed data operations, this will be json stringified data
        // data: JSON.stringify(data),
        data: `0x`,
        format: OperationDataFormat.TRANSACTION,
        txRpcUrl: "",
        type: OperationType.FINAL_TRANSACTION,
      };

      updateAppName("foo bar");

      const onSuccess = (isSuccessful, txHash, error) => {
        console.log("onSuccess", isSuccessful, txHash, error);
      };

      updateTransactionDetails(tx, onSuccess);
    }
  }, [network.chain, account, loading]);

  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  const handleClick = () => {
    previewTransaction();
  };

  return (
    <div className="w-full flex flex-col items-center justify-center gap-2">
      <h1 className="text-xl font-bold">Dynamic + OrbyKit</h1>
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
        <DynamicConnectButton>Connect Wallet</DynamicConnectButton>
      )}
    </div>
  );
}

export default function DynamicOrbyKit() {
  return (
    <DynamicContextProvider
      settings={{
        environmentId: "b7c670c1-fb25-405d-b390-9fb7a60282c1",
        walletConnectors: [
          EthereumWalletConnectors,
          // ZeroDevSmartWalletConnectorsWithConfig({
          //   bundlerRpc: `https://polygon-mainnet.g.alchemy.com/v2/${process.env.NEXT_PUBLIC_ALCHEMY_API_KEY}`,
          //   paymasterRpc: process.env.NEXT_PUBLIC_ZERODEV_PAYMASTER_RPC,
          // }),
        ],
      }}
    >
      <QueryClientProvider client={queryClient}>
        <WagmiProvider config={config}>
          <DynamicWagmiConnector>
            <DynamicExample />
          </DynamicWagmiConnector>
        </WagmiProvider>
      </QueryClientProvider>
    </DynamicContextProvider>
  );
}
