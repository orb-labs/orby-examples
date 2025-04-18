import { OnchainOperation } from "@orb-labs/orby-core";
import { PRIVATE_KEY } from ".";
import { PrivateKeyAccount, TransactionSerializableEIP1559, TypedDataDomain, TypedDataParameter } from "viem";
import { privateKeyToAccount } from "viem/accounts";

// ******************************************** for signing *******************************************

export const noOpSigner = async () => undefined;

// TO-DO: create example that signs non-typed data
export const signTransaction = async (operation: OnchainOperation): Promise<string | undefined> => {
  const wallet: PrivateKeyAccount = privateKeyToAccount(`0x${PRIVATE_KEY}`);

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const { to, data, gasPrice, maxPriorityFeePerGas, maxFeePerGas, gasLimit, value, nonce, chainId } = operation;

  const txData: TransactionSerializableEIP1559 = {
    type: 'eip1559',
    to: to as `0x${string}`,
    data: data as `0x${string}`,
    value: value ?? 0n,
    nonce: Number(nonce),
    gas: gasLimit ?? 21000n,
    chainId: Number(chainId),
    maxFeePerGas: maxFeePerGas ?? gasPrice ?? 0n,
    maxPriorityFeePerGas: maxPriorityFeePerGas ?? 0n,
  };

  return await wallet.signTransaction(txData);
}

export const signTypedData = async (operation: OnchainOperation): Promise<string | undefined> => {
  const wallet = privateKeyToAccount(`0x${PRIVATE_KEY}`);

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const parsedData = JSON.parse(operation.data) as {
    domain: TypedDataDomain;
    types: Record<string, Array<TypedDataParameter>>;
    message: Record<string, any>;
    primaryType: string;
  };

  const eip712Domain = [
    { name: "name", type: "string" },
    { name: "chainId", type: "uint256" },
    { name: "verifyingContract", type: "address" },
  ];
  parsedData.types["EIP712Domain"] = eip712Domain;

  return await wallet.signTypedData(parsedData);
}

// ****************************************************************************************************
