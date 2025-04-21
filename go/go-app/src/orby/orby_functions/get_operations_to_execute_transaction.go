package orbyfunctions

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"

	"go-app/src/orby"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type GetOperationsToExecuteTransaction struct {
	VirtualNodeProvider orby.OrbyClient
	AccountClusterId    string
}

func NewGetOperationsToExecuteTransaction(client orby.OrbyClient, accountClusterId string) *GetOperationsToExecuteTransaction {
	return &GetOperationsToExecuteTransaction{
		VirtualNodeProvider: client,
		AccountClusterId:    accountClusterId,
	}
}

func (g *GetOperationsToExecuteTransaction) Run() error {
	// 0. Check for env variables
	inputTokenAddress := orby.GetEnvWithDefault("INPUT_TOKEN_ADDRESS", "")
	amount := orby.GetEnvWithDefault("AMOUNT", "0")

	// 1. Format operation request
	data, err := g.GetParams(amount)
	if err != nil {
		return err
	}

	// 2. Call operation
	fmt.Println("\n[INFO] calling GetOperationsToExecuteTransaction...")
	result, err := g.VirtualNodeProvider.GetOperationsToExecuteTransaction(
		g.AccountClusterId,
		data,
		inputTokenAddress)
	if err != nil {
		log.Printf("[ERROR] Error getting operations to execute transaction: %v", err)
	}

	// Parse the response into our structured type
	var response orby.OperationSet
	if err := json.Unmarshal(result, &response); err != nil {
		log.Printf("[ERROR] Error parsing orby_GetOperationsToExecuteTransaction response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(result, &rawResponse) == nil {
			fmt.Printf("          Raw response: %v\n", rawResponse)
		}
		return err
	}

	if response.Status == "" {
		var errorResponse orby.ErrorResponse
		if err := json.Unmarshal(result, &errorResponse); err == nil {
			fmt.Printf("\n[ERROR] Failed to get operations to execute transaction:")
			fmt.Printf("\n          Code: %v", errorResponse.Code)
			fmt.Printf("\n          Message: %s", errorResponse.Message)
			return err
		}
		return err
	}

	fmt.Printf("\n[INFO] Operations To Execute Transaction Response:\n")
	fmt.Printf("        Status: %s\n", response.Status)
	fmt.Printf("        Estimated Time: %d ms\n", response.AggregateEstimatedTimeInMs)

	// 3. Sign and send the operations
	if len(response.Intents) > 0 {
		fmt.Printf("        Number of Operations: %d\n", len(response.Intents[0].IntentOperations))

		// Collection of signed operations to send
		var signedOperations []orby.SignedOperation

		// Display each operation and sign based on format
		for i, op := range response.Intents[0].IntentOperations {
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
				signature, signErr = orby.SignTypedData(op)
				if signErr != nil {
					log.Printf("[ERROR] Error signing typed data: %v", signErr)
					continue
				}
				fmt.Printf("          Signed TYPED_DATA: %s\n", signature)
			} else if op.Format == "TRANSACTION" {
				// For TRANSACTION operations, use signTransaction
				signature, signErr = orby.SignTransaction(op)
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
			signedOp := orby.SignedOperation{
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
			sendResult, sendErr := g.VirtualNodeProvider.SendSignedOperations(signedOperations, g.AccountClusterId)
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

	return nil
}

func (g *GetOperationsToExecuteTransaction) GetParams(amount string) (string, error) {
	// 1. Get ERC20 abi
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
	}
	fmt.Println("Current working directory:", cwd)

	erc20AbiJson, err := os.ReadFile("src/abi/erc20.json")
	if err != nil {
		log.Fatalf("Failed to read ABI file: %v", err)
	}

	erc20Abi, err := abi.JSON(bytes.NewReader(erc20AbiJson))
	if err != nil {
		log.Fatalf("Failed to parse ABI: %v", err)
	}

	// 2. Get private key from environment variable to derive address
	privateKey := orby.GetPrivateKey()
	if privateKey == nil {
		log.Fatalf("[ERROR] Failed to get private key: %v", err)
	}

	// 3. Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("[ERROR] Error casting public key to ECDSA")
	}

	// 4. Get recipient address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("\n[INFO] Derived address from private key: %s\n", address)

	// 5. Convert amount to bigint
	bigIntValue := new(big.Int)
	bigIntValue, ok = bigIntValue.SetString(amount, 10) // Base 10 for decimal numbers
	if !ok {
		log.Fatal("Error converting amount to big.Int")
	}

	// 6. Encode transaction data
	data, err := erc20Abi.Pack("transfer", address, bigIntValue)
	if err != nil {
		log.Fatalf("Failed to encode transfer data: %v", err)
	}
	dataHex := hexutil.Encode(data)

	return dataHex, nil
}
