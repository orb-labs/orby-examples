package orbyfunctions

import (
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"

	"go-app/src/orby"
)

type GetOperationsToSignTypedData struct {
	VirtualNodeProvider orby.OrbyClient
	AccountClusterId    string
}

func NewGetOperationsToSignTypedData(client orby.OrbyClient, accountClusterId string) *GetOperationsToSignTypedData {
	return &GetOperationsToSignTypedData{
		VirtualNodeProvider: client,
		AccountClusterId:    accountClusterId,
	}
}

func (g *GetOperationsToSignTypedData) Run() error {
	// 0. Check for env variables
	inputTokenAddress := orby.GetEnvWithDefault("INPUT_TOKEN_ADDRESS", "")
	inputTokenChainId, err := strconv.ParseInt(orby.GetEnvWithDefault("INPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		return err
	}
	amount, err := strconv.ParseInt(orby.GetEnvWithDefault("AMOUNT", "0"), 10, 64)
	if err != nil {
		return err
	}

	// 1. Format operation request
	data, err := g.GetParams(inputTokenAddress, inputTokenChainId, amount)
	if err != nil {
		return err
	}

	// 2. Call operation
	fmt.Println("\n[INFO] calling GetOperationsToSignTypedData...")
	result, err := g.VirtualNodeProvider.GetOperationsToSignTypedData(
		g.AccountClusterId,
		data)
	if err != nil {
		log.Printf("[ERROR] Error getting operations to sign typed data: %v", err)
	}

	// Parse the response into our structured type
	var response orby.OperationSet
	if err := json.Unmarshal(result, &response); err != nil {
		log.Printf("[ERROR] Error parsing orby_GetOperationsToSignTypedData response: %v", err)
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
			fmt.Printf("\n[ERROR] Failed to get operations to sign typed data")
			fmt.Printf("\n          Code: %v", errorResponse.Code)
			fmt.Printf("\n          Message: %s", errorResponse.Message)
			return err
		}
		return err
	}

	fmt.Printf("\n[INFO] Operations To Sign Typed Data Response:\n")
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

func (g *GetOperationsToSignTypedData) GetParams(
	inputTokenAddress string,
	inputTokenChainId int64,
	amount int64) (string, error) {
	// 1. Format Permit2 structure
	permit2 := orby.Permit2Object{
		Types: map[string][]orby.EIP712Type{
			"PermitTransferFrom": {
				{Name: "permitted", Type: "TokenPermissions"},
				{Name: "spender", Type: "address"},
				{Name: "nonce", Type: "uint256"},
				{Name: "deadline", Type: "uint256"},
				{Name: "witness", Type: "ExclusiveDutchOrder"},
			},
			"TokenPermissions": {
				{Name: "token", Type: "address"},
				{Name: "amount", Type: "uint256"},
			},
			"EIP712Domain": {
				{Name: "name", Type: "string"},
				{Name: "chainId", Type: "uint256"},
				{Name: "verifyingContract", Type: "address"},
			},
		},
		Domain: orby.EIP712Domain{
			Name:              "Permit2",
			ChainId:           big.NewInt(inputTokenChainId),
			VerifyingContract: "0x000000000022d473030f116ddee9f6b43ac78ba3",
		},
		PrimaryType: "PermitTransferFrom",
		Message: orby.PermitTransferFrom{
			Permitted: orby.TokenPermissions{
				Token:  inputTokenAddress,
				Amount: big.NewInt(amount),
			},
			Spender:  "0xSpenderAddressHere",
			Nonce:    big.NewInt(10),
			Deadline: big.NewInt(100),
		},
	}

	jsonBytes, err := json.Marshal(permit2)
	if err != nil {
		fmt.Println("Error marshaling permit2 JSON:", err)
		return "", err
	}

	return string(jsonBytes), nil
}
