// main.go is the entry point of the application.
// It handles the logic to create, sign, and send transactions via Orby.

package main

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
	"github.com/joho/godotenv"
)

func main() {
	// Set up account cluster, virtual node, and private key based on env vars
	accountClusterId, virtualNodeClient, privateKey, externalInputTokenChainId, externalOutputTokenChainId := setup()
	if accountClusterId == "" || virtualNodeClient == nil || privateKey == nil {
		log.Fatalf("[ERROR] Error setting up account cluster, virtual node, or private key")
	}

	// Define input and output tokens
	inputTokenAddress := getEnvWithDefault("INPUT_TOKEN_ADDRESS", "")
	outputTokenAddress := getEnvWithDefault("OUTPUT_TOKEN_ADDRESS", "")

	// Create token parameters
	tokens := []TokenParams{
		{
			ChainId:      externalInputTokenChainId,
			TokenAddress: inputTokenAddress,
		},
		{
			ChainId:      externalOutputTokenChainId,
			TokenAddress: outputTokenAddress,
		},
	}

	// Log the request parameters
	fmt.Println("\nGetting standardized token IDs for:")
	for i, token := range tokens {
		fmt.Printf("  Token %d:\n", i+1)
		fmt.Printf("    Chain ID: %s\n", token.ChainId)
		fmt.Printf("    Address: %s\n", token.TokenAddress)
	}

	// Call orby_getStandardizedTokenIds using the virtual node RPC URL
	tokenIdsResult, err := virtualNodeClient.GetStandardizedTokenIds(tokens)
	if err != nil {
		log.Printf("[ERROR] Error getting standardized token IDs: %v", err)
		return
	}

	// Parse and display the standardized token IDs response
	var tokenIdsResponse StandardizedTokenIdsResponse
	if err := json.Unmarshal(tokenIdsResult, &tokenIdsResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_getStandardizedTokenIds response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(tokenIdsResult, &rawResponse) == nil {
			fmt.Printf("          Raw standardized token IDs response: %v\n", rawResponse)
		}
		return
	}

	jsonFormatted, err := json.MarshalIndent(tokenIdsResponse, "", "  ")
	if err == nil {
		fmt.Printf("\n[INFO] Standardized token IDs:\n%s\n", string(jsonFormatted))
	} else {
		fmt.Printf("\n[INFO] Standardized token IDs:\n%v\n", tokenIdsResponse)
	}

	// Now call orby_getOperationsToSwap with the virtual node RPC URL and standardized token IDs
	fmt.Println("\nCalling orby_getOperationsToSwap...")
	amount := getEnvWithDefault("AMOUNT", "1") // 1 token with 18 decimals
	swapResult, err := virtualNodeClient.GetOperationsToSwap(
		accountClusterId,
		tokenIdsResponse.StandardizedTokenIds,
		amount,
		externalInputTokenChainId,
		externalOutputTokenChainId)
	if err != nil {
		log.Printf("[ERROR] Error getting operations to swap: %v", err)
	}

	// Parse the response into our structured type
	var swapResponse GetOperationsToSwapResponse
	if err := json.Unmarshal(swapResult, &swapResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_getOperationsToSwap response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(swapResult, &rawResponse) == nil {
			fmt.Printf("          Raw swap operations response: %v\n", rawResponse)
		}
		return
	}

	// Display structured response information
	fmt.Printf("\n[INFO] Swap Operations Response:\n")
	fmt.Printf("        Status: %s\n", swapResponse.Status)
	fmt.Printf("        Estimated Time: %d ms\n", swapResponse.AggregateEstimatedTimeInMs)

	if len(swapResponse.Intents) > 0 {
		fmt.Printf("        Number of Operations: %d\n", len(swapResponse.Intents[0].IntentOperations))

		// Collection of signed operations to send
		var signedOperations []SignedOperation

		// Display each operation and sign based on format
		for i, op := range swapResponse.Intents[0].IntentOperations {
			fmt.Printf("\n        Operation %d:\n", i+1)
			fmt.Printf("          Type: %s\n", op.Type)
			fmt.Printf("          Format: %s\n", op.Format)
			fmt.Printf("          From: %s\n", op.From)
			fmt.Printf("          To: %s\n", op.To)
			fmt.Printf("          Chain ID: %s\n", op.ChainId)
			fmt.Printf("          TX RPC URL: %s\n", op.TxRpcUrl)

			// Sign the operation based on its format
			var signature string
			var signErr error

			if op.Format == "TYPED_DATA" {
				// For TYPED_DATA operations, use signTypedData
				signature, signErr = signTypedData(op.Data, privateKey)
				if signErr != nil {
					log.Printf("[ERROR] Error signing typed data: %v", signErr)
					continue
				}
				fmt.Printf("          Signed TYPED_DATA: %s\n", signature)
			} else if op.Format == "TRANSACTION" {
				// For TRANSACTION operations, use signTransaction
				signature, signErr = signTransaction(op.Data, op.ChainId, op.To, privateKey)
				if signErr != nil {
					log.Printf("[ERROR] Error signing transaction: %v", signErr)
					continue
				}
				fmt.Printf("          Signed TRANSACTION: %s\n", signature)
			} else {
				fmt.Printf("          Unknown format, no signature generated\n")
				continue
			}

			// Create a signed operation for sending
			signedOp := SignedOperation{
				Type:      op.Type,
				Signature: signature,
				Data:      op.Data,
				ChainId:   op.ChainId,
				From:      op.From,
			}

			signedOperations = append(signedOperations, signedOp)
		}

		// If we have signed operations, send them
		if len(signedOperations) > 0 {
			fmt.Printf("\nSending %d signed operations to orby_sendSignedOperations...\n", len(signedOperations))

			// Send the signed operations
			sendResult, sendErr := virtualNodeClient.SendSignedOperations(signedOperations, accountClusterId)
			if sendErr != nil {
				log.Printf("[ERROR] Error sending signed operations: %v", sendErr)
			} else {
				// Parse and display the send result
				var sendResponse any
				if jsonErr := json.Unmarshal(sendResult, &sendResponse); jsonErr != nil {
					log.Printf("[ERROR] Error parsing orby_sendSignedOperations response: %v", jsonErr)
					// Try to display raw response
					fmt.Printf("[INFO] Raw response: %s\n", string(sendResult))
				} else {
					jsonFormatted, _ := json.MarshalIndent(sendResponse, "", "  ")
					fmt.Printf("\n[INFO] Signed operations sent successfully:\n%s\n", string(jsonFormatted))
				}
			}
		}
	}

	// Display the full response as JSON if needed
	if verbose := os.Getenv("VERBOSE"); verbose == "1" || verbose == "true" {
		jsonFormatted, err := json.MarshalIndent(swapResponse, "", "  ")
		if err != nil {
			fmt.Printf("\n[INFO] Full response (unformatted): %v\n", swapResponse)
		} else {
			fmt.Printf("\n[INFO] Full response:\n%s\n", string(jsonFormatted))
		}
	}
}

// setup creates an account cluster, virtual node, and private key based on the defined environment variables
func setup() (string, *OrbyClient, *ecdsa.PrivateKey, string, string) {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARN] .env file not found, using environment variables")
	}

	// Get Orby-related URLs from environment variables
	orbyEngineAdminURL := getEnvWithDefault("ORBY_ENGINE_ADMIN_URL", "")
	orbyURL := getEnvWithDefault("ORBY_URL", "")

	// Create Orby Admin client
	orbyClient := NewOrbyClient(orbyEngineAdminURL, orbyURL)
	fmt.Printf("[INFO] Orby Engine Admin URL: %s\n", orbyEngineAdminURL)
	fmt.Printf("[INFO] Orby URL: %s\n", orbyURL)

	// Create Orby instance
	instanceName := getEnvWithDefault("ORBY_INSTANCE_NAME", "")
	fmt.Printf("\n[INFO] Creating Orby instance with name: %s\n", instanceName)
	instanceResponse, err := orbyClient.CreateOrbyInstance(instanceName)
	if err != nil {
		log.Fatalf("[ERROR] Error creating Orby instance: %v", err)
	}

	fmt.Printf("[INFO] Orby instance created successfully:\n")
	fmt.Printf("         Success: %v\n", instanceResponse.Success)
	fmt.Printf("         Private URL: %s\n", instanceResponse.OrbyInstancePrivateUrl)
	fmt.Printf("         Public URL: %s\n", instanceResponse.OrbyInstancePublicUrl)

	// Create a private Orby client using the private URL from the response
	privateOrbyClient := NewOrbyClient(instanceResponse.OrbyInstancePrivateUrl, instanceResponse.OrbyInstancePrivateUrl)

	// Get private key from environment variable to derive address
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("[ERROR] PRIVATE_KEY environment variable is required")
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

	// Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("[ERROR] Error casting public key to ECDSA")
	}

	// Get address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Printf("\n[INFO] Derived address from private key: %s\n", address)

	// Create an account cluster using the private Orby client and address
	fmt.Println("\nCreating account cluster...")

	// Create the accounts array with a single EVM EOA account
	accounts := []AccountParams{
		{
			VMType:      "EVM",
			Address:     address,
			AccountType: "EOA",
		},
	}

	// Call orby_createAccountCluster
	clusterResult, err := privateOrbyClient.CreateAccountCluster(accounts)
	if err != nil {
		log.Fatalf("[ERROR] Error creating account cluster: %v", err)
	}

	// Parse the account cluster response into a structured type
	var clusterResponse AccountClusterResponse
	if err := json.Unmarshal(clusterResult, &clusterResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_createAccountCluster response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(clusterResult, &rawResponse) == nil {
			fmt.Printf("          Raw account cluster response: %v\n", rawResponse)
		}
		return "", nil, nil, "", ""
	}

	fmt.Println("\n[INFO] Account cluster created successfully:")
	fmt.Printf("         Account Cluster ID: %s\n", clusterResponse.AccountClusterId)
	fmt.Printf("         ID: %s\n", clusterResponse.Id)
	fmt.Println("         Accounts:")
	for i, account := range clusterResponse.Accounts {
		fmt.Printf("           Account %d:\n", i+1)
		fmt.Printf("             Address: %s\n", account.Address)
		fmt.Printf("             Type: %s\n", account.AccountType)
		fmt.Printf("             VM Type: %s\n", account.VMType)
		fmt.Printf("             Chain ID: %s\n", account.ChainId)
	}

	// Get source / destination chain ids
	inputTokenChainId, err := strconv.ParseInt(getEnvWithDefault("INPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		log.Fatalf("[ERROR] Error getting token chain id: %v", err)
	}
	outputTokenChainId, err := strconv.ParseInt(getEnvWithDefault("OUTPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		log.Fatalf("[ERROR] Error getting token chain id: %v", err)
	}

	// Format chain IDs to external format
	externalInputTokenChainId := GetExternalChainIdFromInternalChainId(inputTokenChainId)
	externalOutputTokenChainId := GetExternalChainIdFromInternalChainId(outputTokenChainId)

	fmt.Printf("\n[INFO] Input chain ID: %d (external format: %s)\n", inputTokenChainId, externalInputTokenChainId)
	fmt.Printf("\n[INFO] Output chain ID: %d (external format: %s)\n", outputTokenChainId, externalOutputTokenChainId)

	// Get virtual node RPC URL
	fmt.Println("\nGetting virtual node RPC URL...")
	virtualNodeResult, err := privateOrbyClient.GetVirtualNodeRpcUrl(
		clusterResponse.AccountClusterId,
		externalInputTokenChainId,
		address,
	)

	if err != nil {
		log.Fatalf("[ERROR] Error getting virtual node RPC URL: %v", err)
	}

	// Parse and display the virtual node RPC URL response
	var virtualNodeResponse VirtualNodeRpcUrlResponse
	var virtualNodeRpcUrl string
	if err := json.Unmarshal(virtualNodeResult, &virtualNodeResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_getVirtualNodeRpcUrl response: %v", err)
		// Try to display raw response
		var rawResponse interface{}
		if json.Unmarshal(virtualNodeResult, &rawResponse) == nil {
			fmt.Printf("          Raw virtual node RPC URL response: %v\n", rawResponse)
		}
	} else {
		virtualNodeRpcUrl = virtualNodeResponse.VirtualNodeRpcUrl
		fmt.Printf("\n[INFO] Virtual Node RPC URL: %s\n", virtualNodeRpcUrl)
		fmt.Printf("\n[INFO] You can now use this URL to interact with the virtual node:\n%s\n", virtualNodeRpcUrl)
	}

	// Create a client using the virtual node RPC URL for standardized token IDs
	virtualNodeClient := NewOrbyClient(virtualNodeRpcUrl, virtualNodeRpcUrl)

	// Return information on account cluster, virtual node, and private key
	return clusterResponse.AccountClusterId, virtualNodeClient, privateKey, externalInputTokenChainId, externalOutputTokenChainId
}

// signTypedData signs EIP-712 typed data with the provided private key and returns the signature and recovered address
func signTypedData(typedDataJSON string, privateKey *ecdsa.PrivateKey) (string, error) {
	// Log the typed data JSON for inspection
	fmt.Println("Raw typed data JSON:")
	fmt.Println(typedDataJSON)

	// Parse the JSON into TypedData struct from go-ethereum
	var typedData apitypes.TypedData
	if err := json.Unmarshal([]byte(typedDataJSON), &typedData); err != nil {
		return "", fmt.Errorf("failed to parse typed data JSON: %v", err)
	}

	addEIP712DomainTypeToTypedData(&typedData)

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

// signTransaction signs a transaction using the provided private key and verifies the signature
func signTransaction(txData string, chainIDStr string, to string, privateKey *ecdsa.PrivateKey) (string, error) {
	// Output the incoming data for debugging
	fmt.Println("Transaction Data to sign:")
	fmt.Println(txData)
	fmt.Println("Chain ID:", chainIDStr)
	fmt.Println("To Address:", to)

	// Check if the txData is a hex string or JSON
	var txMap map[string]interface{}
	isJSON := false

	// Try to parse as JSON
	if err := json.Unmarshal([]byte(txData), &txMap); err == nil {
		isJSON = true
		fmt.Println("Transaction data is in JSON format")
	} else {
		fmt.Println("Transaction data is in hex format")
	}

	// Parse the chain ID
	chainIDStr = strings.TrimPrefix(chainIDStr, "eip155-")
	chainID, ok := new(big.Int).SetString(chainIDStr, 10)
	if !ok {
		return "", fmt.Errorf("invalid chain ID: %s", chainIDStr)
	}

	// Create the transaction
	toAddr := common.HexToAddress(to)

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
		txData_bytes = common.FromHex(txData)
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
