// orby_client.go makes the rpc calls to Orby
package orby

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// Define OrbyClient struct to interact with Orby Engine API
type OrbyClient struct {
	EngineAdminURL string
	OrbyURL        string
	HTTPClient     *http.Client
}

// NewOrbyClient creates a new OrbyClient instance
func NewOrbyClient(engineAdminURL, orbyURL string) *OrbyClient {
	return &OrbyClient{
		EngineAdminURL: engineAdminURL,
		OrbyURL:        orbyURL,
		HTTPClient:     &http.Client{},
	}
}

// SendJSONRPCRequest sends a JSON-RPC request to the specified URL
func (c *OrbyClient) SendJSONRPCRequest(url string, method string, params []interface{}) (json.RawMessage, error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, fmt.Errorf("RPC error: %s (code: %d)", result.Error.Message, result.Error.Code)
	}

	return result.Result, nil
}

// CreateOrbyInstance creates a new Orby instance with the given name
func (c *OrbyClient) CreateOrbyInstance(name string) (*OrbyInstanceResponse, error) {
	params := []interface{}{
		map[string]interface{}{
			"name": name,
		},
	}

	resultBytes, err := c.SendJSONRPCRequest(c.EngineAdminURL, "orby_createInstance", params)
	if err != nil {
		return nil, err
	}

	var response OrbyInstanceResponse
	if err := json.Unmarshal(resultBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse orby_createInstance response: %v", err)
	}

	return &response, nil
}

// CreateAccountCluster creates an account cluster with the given accounts
func (c *OrbyClient) CreateAccountCluster(accounts []AccountParams) (json.RawMessage, error) {
	params := []interface{}{
		CreateAccountClusterParams{
			Accounts: accounts,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_createAccountCluster", params)
}

// GetVirtualNodeRpcUrl gets the virtual node RPC URL for the given parameters
func (c *OrbyClient) GetVirtualNodeRpcUrl(accountClusterId string, chainId string, entrypointAccountAddress string) (json.RawMessage, error) {
	params := []any{
		GetVirtualNodeRpcUrlParams{
			AccountClusterId:         accountClusterId,
			ChainId:                  chainId,
			EntrypointAccountAddress: entrypointAccountAddress,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getVirtualNodeRpcUrl", params)
}

// GetStandardizedTokenIds gets standardized token IDs for the given tokens
func (c *OrbyClient) GetStandardizedTokenIds(tokens []TokenParams) (json.RawMessage, error) {
	params := []any{
		GetStandardizedTokenIdsParams{
			Tokens: tokens,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getStandardizedTokenIds", params)
}

// Call orby_getOperationsToSwap with the provided parameters
func (c *OrbyClient) GetOperationsToSwap(
	accountClusterId string,
	standardizedTokenIds []string,
	amount string,
	externalInputTokenChainId string,
	externalOutputTokenChainId string) (json.RawMessage, error) {

	// Check if standardizedTokenIds is empty
	if len(standardizedTokenIds) == 0 {
		log.Fatal("StandardizedTokenIds is empty. Make sure tokenIdsResponse contains token IDs")
	}

	// Prepare the parameters for orby_getOperationsToSwap
	params := GetOperationsToSwapParams{
		AccountClusterId: accountClusterId,
		SwapType:         "EXACT_INPUT",
		Input: InputSwapParam{
			StandardizedTokenId: standardizedTokenIds[0],
			Amount:              amount,
			TokenSources: []TokenSource{
				{
					ChainID: externalInputTokenChainId,
				},
			},
		},
		Output: OutputSwapParam{
			StandardizedTokenId: standardizedTokenIds[len(standardizedTokenIds)-1],
			TokenDestination: TokenSource{
				ChainID: externalOutputTokenChainId,
			},
		},
	}

	// Log the request parameters
	fmt.Println("\nCalling orby_getOperationsToSwap with parameters:")
	jsonParams, _ := json.MarshalIndent(params, "", "  ")
	fmt.Println(string(jsonParams))

	// Call orby_getOperationsToSwap
	rpcParams := []any{params}
	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getOperationsToSwap", rpcParams)
}

// Call orby_getOperationsToExecuteTransaction with the params
func (c *OrbyClient) GetOperationsToExecuteTransaction(
	accountClusterId string,
	data string,
	to string) (json.RawMessage, error) {
	params := []interface{}{
		GetOperationsToExecuteTransactionParams{
			AccountClusterId: accountClusterId,
			Data:             data,
			To:               to,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getOperationsToExecuteTransaction", params)
}

// Call orby_getOperationsToSignTypedData with the params
func (c *OrbyClient) GetOperationsToSignTypedData(
	accountClusterId string,
	data string) (json.RawMessage, error) {
	params := []interface{}{
		GetOperationsToSignTypedDataParams{
			AccountClusterId: accountClusterId,
			Data:             data,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getOperationsToSignTypedData", params)
}

// Call orby_getFungibleTokenPortfolio with the params
func (c *OrbyClient) GetFungibleTokenPortfolio(
	accountClusterId string) (json.RawMessage, error) {
	params := []interface{}{
		GetFungibleTokenPortfolioParams{
			AccountClusterId: accountClusterId,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_getFungibleTokenPortfolio", params)
}

// SendOperationSet sends signed operations to the virtual node
func (c *OrbyClient) SendSignedOperations(signedOperations []SignedOperation, accountClusterId string) (json.RawMessage, error) {
	params := []interface{}{
		SendSignedOperationsParams{
			SignedOperations: signedOperations,
			AccountClusterId: accountClusterId,
		},
	}

	return c.SendJSONRPCRequest(c.OrbyURL, "orby_sendSignedOperations", params)
}
