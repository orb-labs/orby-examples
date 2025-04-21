const { ethers } = require("ethers");
const { PRIVATE_KEY } = require("./config");
const ERC20ABI = require("./abi/erc20.json");

// ******************************************** for signing *******************************************

// TO-DO: create example that signs non-typed data
const signTransaction = async (operation) => {
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

const signTypedData = async (operation) => {
  const wallet = new ethers.Wallet(PRIVATE_KEY);

  console.log(`\n[INFO] Signing operation...`);
  console.log(`         Type: ${operation.type}`);
  console.log(`         Format: ${operation.format}`);
  console.log(`         From: ${operation.from}`);
  console.log(`         To: ${operation.to}`);
  console.log(`         Chain ID: ${operation.chainId}`);
  console.log(`         TX RPC URL: ${operation.txRpcUrl}`);

  const parsedData = JSON.parse(operation.data);
  delete parsedData.types['EIP712Domain'];

  return await wallet.signTypedData(parsedData.domain, parsedData.types, parsedData.message);
}

// ****************************************************************************************************

const iface = new ethers.Interface(ERC20ABI);

module.exports = { signTransaction, signTypedData, iface };