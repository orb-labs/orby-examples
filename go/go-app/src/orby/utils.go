// utils.go provides utility functions used throughout the app
package orby

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// GetEnvWithDefault returns the value of the environment variable or the default value
func GetEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Version for use with go-ethereum's apitypes.TypedData
func AddEIP712DomainTypeToTypedData(typedData *apitypes.TypedData) {
	// Check if Types exists, initialize if not
	if typedData.Types == nil {
		typedData.Types = make(map[string][]apitypes.Type)
	}

	// Define EIP712Domain type
	eip712Domain := []apitypes.Type{
		{Name: "name", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	}

	// Add to Types
	typedData.Types["EIP712Domain"] = eip712Domain
}

// GetExternalChainIdFromInternalChainId converts an internal chain ID to the external format
// Returns a string in the format "eip155-{chainId}" or empty string if the input is 0
func GetExternalChainIdFromInternalChainId(chainId int64) string {
	if chainId == 0 {
		return ""
	}
	return fmt.Sprintf("eip155-%d", chainId)
}

func GetPrivateKey() *ecdsa.PrivateKey {
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatalf("[ERROR] PRIVATE_KEY environment variable is required")
	}

	// If private key starts with "0x", remove it
	if len(privateKeyHex) > 2 && privateKeyHex[:2] == "0x" {
		privateKeyHex = privateKeyHex[2:]
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		log.Fatalf("[ERROR] Failed to parse private key: %v", err)
	}

	return privateKey
}

func SignTransaction(operation Operation) (string, error) {
	// Get private key
	privateKey := GetPrivateKey()

	// Output the incoming data for debugging
	fmt.Println("Transaction Data to sign:")
	fmt.Println(operation.Data)
	fmt.Println("Chain ID:", operation.ChainId)
	fmt.Println("To Address:", operation.To)

	// Check if the txData is a hex string or JSON
	var txMap map[string]interface{}
	isJSON := false

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(operation.Data), &txMap); err == nil {
		isJSON = true
		fmt.Println("Transaction data is in JSON format")
	} else {
		fmt.Println("Transaction data is in hex format")
	}

	// Parse the chain ID
	chainIDStr := strings.TrimPrefix(operation.ChainId, "eip155-")
	chainID, ok := new(big.Int).SetString(operation.ChainId, 10)
	if !ok {
		return "", fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}

	// Create the transaction
	toAddr := common.HexToAddress(operation.To)

	// Default values
	gasLimit := uint64(300_000) // default gas limit
	value := big.NewInt(0)      // default value is 0 ETH
	nonce := uint64(0)          // default nonce is 0

	var txData_bytes []byte

	// If we have JSON data, extract values
	if isJSON {
		// Extract the necessary fields
		if gasLimitStr, ok := txMap["gasLimit"].(string); ok {
			gasLimit_temp, _ := strconv.ParseUint(strings.TrimPrefix(gasLimitStr, "0x"), 16, 64)
			if gasLimit_temp > 0 {
				gasLimit = gasLimit_temp
			}
		}

		if valueStr, ok := txMap["value"].(string); ok && valueStr != "" {
			value_temp, ok := new(big.Int).SetString(strings.TrimPrefix(valueStr, "0x"), 16)
			if ok {
				value = value_temp
			}
		}

		if nonceStr, ok := txMap["nonce"].(string); ok && nonceStr != "" {
			nonce_temp, _ := strconv.ParseUint(strings.TrimPrefix(nonceStr, "0x"), 16, 64)
			nonce = nonce_temp
		}

		// Extract data field
		if dataStr, ok := txMap["data"].(string); ok && dataStr != "" {
			txData_bytes = common.FromHex(dataStr)
		} else {
			txData_bytes = []byte{}
		}
	} else {
		// Raw hex data
		txData_bytes = common.FromHex(operation.Data)
	}

	fmt.Printf("Transaction details: gasLimit=%d, nonce=%d\n", gasLimit, nonce)

	// Create a transaction based on the available fields
	var tx *types.Transaction

	// Check for EIP-1559 fields
	if isJSON && txMap["maxFeePerGas"] != nil {
		maxFeePerGasStr, _ := txMap["maxFeePerGas"].(string)
		maxPriorityFeePerGasStr, _ := txMap["maxPriorityFeePerGas"].(string)

		maxFeePerGas, _ := new(big.Int).SetString(strings.TrimPrefix(maxFeePerGasStr, "0x"), 16)
		maxPriorityFeePerGas, _ := new(big.Int).SetString(strings.TrimPrefix(maxPriorityFeePerGasStr, "0x"), 16)

		fmt.Printf("EIP-1559 transaction: maxFeePerGas=%s, maxPriorityFeePerGas=%s\n",
			maxFeePerGas.String(), maxPriorityFeePerGas.String())

		// Create an EIP-1559 transaction
		tx = types.NewTx(&types.DynamicFeeTx{
			ChainID:   chainID,
			Nonce:     nonce,
			To:        &toAddr,
			Value:     value,
			Gas:       gasLimit,
			GasFeeCap: maxFeePerGas,
			GasTipCap: maxPriorityFeePerGas,
			Data:      txData_bytes,
		})
	} else {
		// Create a legacy transaction
		gasPrice := big.NewInt(1000000000) // default gas price is 1 Gwei
		if isJSON && txMap["gasPrice"] != nil {
			gasPriceStr, _ := txMap["gasPrice"].(string)
			gasPrice, _ = new(big.Int).SetString(strings.TrimPrefix(gasPriceStr, "0x"), 16)
		}

		fmt.Printf("Legacy transaction: gasPrice=%s\n", gasPrice.String())

		tx = types.NewTransaction(
			nonce,
			toAddr,
			value,
			gasLimit,
			gasPrice,
			txData_bytes,
		)
	}

	// Sign the transaction
	signer := types.NewEIP155Signer(chainID)
	signedTx, err := types.SignTx(tx, signer, privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Verify the signature by recovering the sender address
	sender, err := signer.Sender(signedTx)
	if err != nil {
		return "", fmt.Errorf("failed to recover sender from signed transaction: %v", err)
	}

	// Get the original signer's address for comparison
	originalAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Verify that the recovered sender matches the original signer's address
	if sender != originalAddress {
		return "", fmt.Errorf("transaction signature verification failed: recovered sender %s does not match original address %s",
			sender.Hex(), originalAddress.Hex())
	}

	fmt.Printf("Transaction signature verification successful!\n")
	fmt.Printf("  Original signer address: %s\n", originalAddress.Hex())
	fmt.Printf("  Recovered sender address: %s\n", sender.Hex())

	// Convert the signed transaction to raw bytes
	signedTxBytes, err := signedTx.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to marshal signed transaction: %v", err)
	}

	return "0x" + hex.EncodeToString(signedTxBytes), nil
}

func SignTypedData(operation Operation) (string, error) {
	// Get private key
	privateKey := GetPrivateKey()

	// Log the typed data JSON for inspection
	fmt.Println("Raw typed data JSON:")
	fmt.Println(operation.Data)

	// Parse the JSON into TypedData struct from go-ethereum
	var typedData apitypes.TypedData
	if err := json.Unmarshal([]byte(operation.Data), &typedData); err != nil {
		return "", fmt.Errorf("failed to parse typed data JSON: %v", err)
	}

	AddEIP712DomainTypeToTypedData(&typedData)

	// Validate that this looks like typed data
	if typedData.Domain == (apitypes.TypedDataDomain{}) {
		return "", fmt.Errorf("typed data missing 'domain' field")
	}
	if len(typedData.Types) == 0 {
		return "", fmt.Errorf("typed data missing 'types' field")
	}
	if typedData.PrimaryType == "" {
		return "", fmt.Errorf("typed data missing 'primaryType' field")
	}
	if len(typedData.Message) == 0 {
		return "", fmt.Errorf("typed data missing 'message' field")
	}

	fmt.Println("Valid EIP-712 typed data structure detected")

	// 1. Compute the typed data hash according to EIP-712
	// Hash the struct data
	typedDataHash, err := typedData.HashStruct(typedData.PrimaryType, typedData.Message)
	if err != nil {
		return "", fmt.Errorf("failed to hash struct data: %v", err)
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return "", fmt.Errorf("failed to create domain separator: %v", err)
	}

	// 3. Combine with EIP-712 prefix
	rawData := []byte(fmt.Sprintf("\x19\x01%s%s", string(domainSeparator), string(typedDataHash)))
	finalHash := crypto.Keccak256Hash(rawData)

	fmt.Printf("EIP-712 encoded hash: %s\n", finalHash.Hex())

	// 4. Sign the hash
	signature, err := crypto.Sign(finalHash.Bytes(), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign typed data hash: %v", err)
	}

	// Store the original signature for verification
	originalSignature := make([]byte, len(signature))
	copy(originalSignature, signature)

	// 5. Adjust v value (add 27) for Ethereum compatibility
	if signature[64] < 27 {
		signature[64] += 27
	}

	// 6. Verify the signature by recovering the public key
	pubKey, err := crypto.Ecrecover(finalHash.Bytes(), originalSignature)
	if err != nil {
		return "", fmt.Errorf("failed to recover public key from signature: %v", err)
	}

	// Convert the recovered public key bytes to an Ethereum address
	recoveredAddress := common.BytesToAddress(crypto.Keccak256(pubKey[1:])[12:])

	// Get the original signer's address for comparison
	originalAddress := crypto.PubkeyToAddress(privateKey.PublicKey)

	// Verify that the recovered address matches the original signer's address
	if recoveredAddress != originalAddress {
		return "", fmt.Errorf("signature verification failed: recovered address %s does not match original address %s",
			recoveredAddress.Hex(), originalAddress.Hex())
	}

	fmt.Printf("Signature verification successful!\n")
	fmt.Printf("  Original signer address: %s\n", originalAddress.Hex())
	fmt.Printf("  Recovered signer address: %s\n", recoveredAddress.Hex())

	// Return the signature in hex format
	return "0x" + hex.EncodeToString(signature), nil
}
