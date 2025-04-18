package examples

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"go-app/src/orby"
)

type GetOperationsToSwap struct {
	VirtualNodeProvider orby.OrbyClient
	AccountClusterId    string
}

func NewGetOperationsToSwap(client orby.OrbyClient, accountClusterId string) *GetOperationsToSwap {
	return &GetOperationsToSwap{
		VirtualNodeProvider: client,
		AccountClusterId:    accountClusterId,
	}
}

func (g *GetOperationsToSwap) Run() error {
	// 0. Check for env variables
	inputTokenAddress := orby.GetEnvWithDefault("INPUT_TOKEN_ADDRESS", "")
	outputTokenAddress := orby.GetEnvWithDefault("OUTPUT_TOKEN_ADDRESS", "")
	inputTokenChainId, err := strconv.ParseInt(orby.GetEnvWithDefault("INPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		return err
	}
	externalInputTokenChainId := orby.GetExternalChainIdFromInternalChainId(inputTokenChainId)
	outputTokenChainId, err := strconv.ParseInt(orby.GetEnvWithDefault("OUTPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		return err
	}
	externalOutputTokenChainId := orby.GetExternalChainIdFromInternalChainId(outputTokenChainId)
	amount := orby.GetEnvWithDefault("AMOUNT", "0")

	// 1. Format operation request
	standardizedTokenIds, err := g.GetParams(
		inputTokenAddress,
		outputTokenAddress,
		externalInputTokenChainId,
		externalOutputTokenChainId)
	if err != nil {
		return err
	}

	// 2. Call operation
	fmt.Println("\n[INFO] calling getOperationsToSwap...")
	swapResult, err := g.VirtualNodeProvider.GetOperationsToSwap(
		g.AccountClusterId,
		*standardizedTokenIds,
		amount,
		externalInputTokenChainId,
		externalOutputTokenChainId)
	if err != nil {
		log.Printf("[ERROR] Error getting operations to swap: %v", err)
	}

	// Parse the response into our structured type
	var swapResponse orby.OperationSet
	if err := json.Unmarshal(swapResult, &swapResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_getOperationsToSwap response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(swapResult, &rawResponse) == nil {
			fmt.Printf("          Raw swap operations response: %v\n", rawResponse)
		}
		return err
	}

	if swapResponse.Status == "" {
		var errorResponse orby.ErrorResponse
		if err := json.Unmarshal(swapResult, &errorResponse); err == nil {
			fmt.Printf("\n[ERROR] Failed to get operations to swap:")
			fmt.Printf("\n          Code: %v", errorResponse.Code)
			fmt.Printf("\n          Message: %s", errorResponse.Message)
			return err
		}
		return err
	}

	fmt.Printf("\n[INFO] Swap Operations Response:\n")
	fmt.Printf("        Status: %s\n", swapResponse.Status)
	fmt.Printf("        Estimated Time: %d ms\n", swapResponse.AggregateEstimatedTimeInMs)

	// 3. Sign and send the operations
	if len(swapResponse.Intents) > 0 {
		fmt.Printf("        Number of Operations: %d\n", len(swapResponse.Intents[0].IntentOperations))

		// Collection of signed operations to send
		var signedOperations []orby.SignedOperation

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

func (g *GetOperationsToSwap) GetParams(
	inputTokenAddress string,
	outputTokenAddress string,
	externalInputTokenChainId string,
	externalOutputTokenChainId string) (*[]string, error) {
	// Create token parameters
	tokens := []orby.TokenParams{
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
	fmt.Println("\n[INFO] getting standardized token IDs...")
	tokenIdsResult, err := g.VirtualNodeProvider.GetStandardizedTokenIds(tokens)
	if err != nil {
		log.Printf("[ERROR] Error getting standardized token IDs: %v", err)
		return nil, err
	}

	// Parse and display the standardized token IDs response
	var tokenIdsResponse orby.StandardizedTokenIdsResponse
	if err := json.Unmarshal(tokenIdsResult, &tokenIdsResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_getStandardizedTokenIds response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(tokenIdsResult, &rawResponse) == nil {
			fmt.Printf("          Raw standardized token IDs response: %v\n", rawResponse)
		}
		return nil, err
	}

	jsonFormatted, err := json.MarshalIndent(tokenIdsResponse, "", "  ")
	if err == nil {
		fmt.Printf("\n[INFO] Standardized token IDs:\n%s\n", string(jsonFormatted))
	} else {
		fmt.Printf("\n[INFO] Standardized token IDs:\n%v\n", tokenIdsResponse)
	}

	return &tokenIdsResponse.StandardizedTokenIds, nil
}
