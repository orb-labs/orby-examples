// main.go is the entry point of the application.
// It handles the logic to create, sign, and send transactions via Orby.

package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"go-app/src/orby"
	"go-app/src/orby/examples"
	"log"
	"strconv"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
)

type ExampleRunner interface {
	Run() error
}

func main() {
	// Set up account cluster, virtual node, and private key based on env vars
	accountClusterId, virtualNodeClient := setup()
	if accountClusterId == "" {
		log.Fatalf("[ERROR] Error setting up account cluster")
	}
	if virtualNodeClient == nil {
		log.Fatalf("[ERROR] Error setting up virtual node")
	}

	// Run example
	var example ExampleRunner

	switch orby.GetEnvWithDefault("EXAMPLE_TYPE", "") {
	case "getOperationsToSwap":
		example = examples.NewGetOperationsToSwap(*virtualNodeClient, accountClusterId)
	case "getOperationsToExecuteTransaction":
		example = examples.NewGetOperationsToExecuteTransaction(*virtualNodeClient, accountClusterId)
	case "getOperationsToSignTypedData":
		example = examples.NewGetOperationsToSignTypedData(*virtualNodeClient, accountClusterId)
	case "getFungibleTokenPortfolio":
		example = examples.NewGetFungibleTokenPortfolio(*virtualNodeClient, accountClusterId)
	default:
		log.Fatalf("invalid example type: %s", orby.GetEnvWithDefault("EXAMPLE_TYPE", ""))
	}

	// Call Run
	if err := example.Run(); err != nil {
		log.Fatalf("failed to run example: %v", err)
	}
}

// setup creates an account cluster, virtual node, and private key based on the defined environment variables
func setup() (string, *orby.OrbyClient) {
	// 0. Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("[WARN] .env file not found, using environment variables")
	}

	// ******************************** Creating a private instance using orby engine admin ********************************

	// 1. Get Orby-related URLs from environment variables
	orbyEngineAdminURL := orby.GetEnvWithDefault("ORBY_ENGINE_ADMIN_URL", "")
	orbyURL := orby.GetEnvWithDefault("ORBY_URL", "")

	// 2. Create Orby Admin client
	orbyClient := orby.NewOrbyClient(orbyEngineAdminURL, orbyURL)
	fmt.Printf("[INFO] Orby Engine Admin URL: %s\n", orbyEngineAdminURL)
	fmt.Printf("[INFO] Orby URL: %s\n", orbyURL)

	// 3. Create Orby instance
	instanceName := orby.GetEnvWithDefault("ORBY_INSTANCE_NAME", "")
	fmt.Printf("\n[INFO] Creating Orby instance with name: %s\n", instanceName)
	instanceResponse, err := orbyClient.CreateOrbyInstance(instanceName)
	if err != nil {
		log.Fatalf("[ERROR] Error creating Orby instance: %v", err)
	}

	fmt.Printf("[INFO] Orby instance created successfully:\n")
	fmt.Printf("         Success: %v\n", instanceResponse.Success)
	fmt.Printf("         Private URL: %s\n", instanceResponse.OrbyInstancePrivateUrl)
	fmt.Printf("         Public URL: %s\n", instanceResponse.OrbyInstancePublicUrl)

	// 4. Create a private Orby client using the private URL from the response
	privateOrbyClient := orby.NewOrbyClient(instanceResponse.OrbyInstancePrivateUrl, instanceResponse.OrbyInstancePrivateUrl)

	// ********************************** Use private instance to create account cluster ***********************************

	// 5. Get private key from environment variable to derive address
	privateKey := orby.GetPrivateKey()
	if privateKey == nil {
		log.Fatalf("[ERROR] Failed to get private key: %v", err)
	}

	// 6. Get public key
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("[ERROR] Error casting public key to ECDSA")
	}

	// 7. Get address from public key
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	fmt.Printf("\n[INFO] Derived address from private key: %s\n", address)

	// 8. Create the accounts array with a single EVM EOA account
	fmt.Println("\nCreating account cluster...")
	accounts := []orby.AccountParams{
		{
			VMType:      "EVM",
			Address:     address,
			AccountType: "EOA",
		},
	}

	// 9. Call orby_createAccountCluster
	clusterResult, err := privateOrbyClient.CreateAccountCluster(accounts)
	if err != nil {
		log.Fatalf("[ERROR] Error creating account cluster: %v", err)
	}

	// 10. Parse the account cluster response into a structured type
	var clusterResponse orby.AccountClusterResponse
	if err := json.Unmarshal(clusterResult, &clusterResponse); err != nil {
		log.Printf("[ERROR] Error parsing orby_createAccountCluster response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(clusterResult, &rawResponse) == nil {
			fmt.Printf("          Raw account cluster response: %v\n", rawResponse)
		}
		return "", nil
	}

	fmt.Println("\n[INFO] Account cluster created successfully:")
	fmt.Printf("         Account Cluster ID: %s\n", clusterResponse.AccountClusterId)
	fmt.Println("         Accounts:")
	for i, account := range clusterResponse.Accounts {
		fmt.Printf("           Account %d:\n", i+1)
		fmt.Printf("             Address: %s\n", account.Address)
		fmt.Printf("             Type: %s\n", account.AccountType)
		fmt.Printf("             VM Type: %s\n", account.VMType)
		fmt.Printf("             Chain ID: %s\n", account.ChainId)
	}

	// ******************************* Create virtual node to interact with account cluster ********************************

	// 11. Get source chain ids
	inputTokenChainId, err := strconv.ParseInt(orby.GetEnvWithDefault("INPUT_TOKEN_CHAIN_ID", ""), 10, 64)
	if err != nil {
		log.Fatalf("[ERROR] Error getting token chain id: %v", err)
	}

	// 12. Format chain IDs to external format
	externalInputTokenChainId := orby.GetExternalChainIdFromInternalChainId(inputTokenChainId)
	fmt.Printf("\n[INFO] Input chain ID: %d (external format: %s)\n", inputTokenChainId, externalInputTokenChainId)

	// 13. Get virtual node RPC URL
	fmt.Println("\nGetting virtual node RPC URL...")
	virtualNodeResult, err := privateOrbyClient.GetVirtualNodeRpcUrl(
		clusterResponse.AccountClusterId,
		externalInputTokenChainId,
		address,
	)
	if err != nil {
		log.Fatalf("[ERROR] Error getting virtual node RPC URL: %v", err)
	}

	// 14. Parse and display the virtual node RPC URL response
	var virtualNodeResponse orby.VirtualNodeRpcUrlResponse
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

	// 15. Create a client using the virtual node RPC URL for standardized token IDs
	virtualNodeClient := orby.NewOrbyClient(virtualNodeRpcUrl, virtualNodeRpcUrl)

	return clusterResponse.AccountClusterId, virtualNodeClient
}
