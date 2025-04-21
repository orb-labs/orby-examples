import { OnchainOperation } from "@orb-labs/orby-core";
import { ethers, TypedDataDomain, TypedDataField } from "ethers";
import { PRIVATE_KEY } from ".";
import ERC20ABI from "./abi/erc20.json";

// ******************************************** for signing *******************************************

export const noOpSigner = async () => undefined;

// TO-DO: create example that signs non-typed data
export const signTransaction = async (operation: OnchainOperation): Promise<string | undefined> => {
  const wallet = new ethers.Wallet(PRIVATE_KEY);

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);
  
  const { from, to, data, gasPrice, maxPriorityFeePerGas, maxFeePerGas, gasLimit, value, nonce, chainId } = operation;
      
  const txData = {
    from,
    to,
    value,
    data,
    nonce: Number(nonce),
    gasLimit,
    chainId,
    gasPrice,
    maxFeePerGas,
    maxPriorityFeePerGas,
  };

  const tx = await wallet.populateTransaction(txData);
  return await wallet.signTransaction(tx);
}

export const signTypedData = async (operation: OnchainOperation): Promise<string | undefined> => {
  const wallet = new ethers.Wallet(PRIVATE_KEY);

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const parsedData = JSON.parse(operation.data) as {
    domain: TypedDataDomain;
    types: Record<string, Array<TypedDataField>>;
    message: Record<string, any>;
  };
  delete parsedData.types['EIP712Domain'];

  return await wallet.signTypedData(parsedData.domain, parsedData.types, parsedData.message);
}

// ****************************************************************************************************

export const iface = new ethers.Interface(ERC20ABI);