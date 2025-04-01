// utils.go provides general-purpose utility functions

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/signer/core/apitypes"
)

// NewOrbyClient creates a new OrbyClient instance
func NewOrbyClient(engineAdminURL, orbyURL string) *OrbyClient {
	return &OrbyClient{
		EngineAdminURL: engineAdminURL,
		OrbyURL:        orbyURL,
		HTTPClient:     &http.Client{},
	}
}

// getEnvWithDefault returns the value of the environment variable or the default value
func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// Version for use with go-ethereum's apitypes.TypedData
func addEIP712DomainTypeToTypedData(typedData *apitypes.TypedData) {
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
