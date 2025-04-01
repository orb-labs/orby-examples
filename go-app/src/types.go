// types.go defines data structures used throughout the application.

package main

import (
	"net/http"
)

// StandardizedTokenIdsResponse represents the response from orby_getStandardizedTokenIds
type StandardizedTokenIdsResponse struct {
	StandardizedTokenIds []string `json:"standardizedTokenIds"`
}

// OrbyInstanceResponse represents the response from orby_createInstance
type OrbyInstanceResponse struct {
	Success                bool   `json:"success"`
	OrbyInstancePrivateUrl string `json:"orbyInstancePrivateUrl"`
	OrbyInstancePublicUrl  string `json:"orbyInstancePublicUrl"`
}

// Define OrbyClient struct to interact with Orby Engine API
type OrbyClient struct {
	EngineAdminURL string
	OrbyURL        string
	HTTPClient     *http.Client
}

// TokenParams represents the parameters for a token in orby_getStandardizedTokenIds
type TokenParams struct {
	ChainId      string `json:"chainId"`
	TokenAddress string `json:"tokenAddress"`
}

// GetStandardizedTokenIdsParams represents the parameters for orby_getStandardizedTokenIds
type GetStandardizedTokenIdsParams struct {
	Tokens []TokenParams `json:"tokens"`
}

// SignedOperation represents a signed operation to be sent to orby_sendSignedOperations
type SignedOperation struct {
	Type      string `json:"type"`
	Signature string `json:"signature"`
	Data      string `json:"data"`
	ChainId   string `json:"chainId"`
	From      string `json:"from"`
}

// SendSignedOperationsParams represents the parameters for orby_sendSignedOperations
type SendSignedOperationsParams struct {
	SignedOperations []SignedOperation `json:"signedOperations"`
	AccountClusterId string            `json:"accountClusterId"`
}

type TokenSource struct {
	ChainID string `json:"chainId"`
	Address string `json:"address,omitempty"`
}

// InputSwapParam represents a token with its standardizedTokenId and amount
type InputSwapParam struct {
	StandardizedTokenId string        `json:"standardizedTokenId"`
	Amount              string        `json:"amount,omitempty"`
	TokenSources        []TokenSource `json:"tokenSources,omitempty"`
}

// SwapParam represents a token with its standardizedTokenId and amount
type OutputSwapParam struct {
	StandardizedTokenId string      `json:"standardizedTokenId"`
	Amount              string      `json:"amount,omitempty"`
	TokenDestination    TokenSource `json:"tokenDestination,omitempty"`
}

// GetOperationsToSwapParams represents the parameters for orby_getOperationsToSwap
type GetOperationsToSwapParams struct {
	AccountClusterId string          `json:"accountClusterId"`
	SwapType         string          `json:"swapType"`
	Input            InputSwapParam  `json:"input"`
	Output           OutputSwapParam `json:"output"`
}

// Asset represents a cryptocurrency or token asset
type Asset struct {
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}

// Currency represents a currency with its asset details and decimals
type Currency struct {
	Asset    Asset `json:"asset"`
	Decimals int   `json:"decimals"`
	IsNative bool  `json:"isNative"`
}

// AmountValue represents a monetary amount with value information
type AmountValue struct {
	Amount   string   `json:"amount"`
	Currency Currency `json:"currency"`
	Value    string   `json:"value"`
}

// Token represents a token with its details
type Token struct {
	Typename    string   `json:"__typename"`
	Address     string   `json:"address"`
	ChainId     string   `json:"chainId"`
	CoinGeckoId string   `json:"coinGeckoId,omitempty"`
	Currency    Currency `json:"currency"`
	IsNative    bool     `json:"isNative"`
}

// TokenAmount represents a token with its amount
type TokenAmount struct {
	Amount string `json:"amount"`
	Token  Token  `json:"token"`
	Value  string `json:"value"`
}

// TokenState represents a state containing different token types and amounts
type TokenState struct {
	FungibleTokenAmounts     []TokenAmount `json:"fungibleTokenAmounts"`
	NonFungibleTokenAmounts  []TokenAmount `json:"nonFungibleTokenAmounts"`
	SemiFungibleTokenAmounts []TokenAmount `json:"semiFungibleTokenAmounts"`
}

// Operation represents a blockchain operation/transaction
type Operation struct {
	ChainId                            string       `json:"chainId"`
	Data                               string       `json:"data"`
	EstimatedNetworkFees               *AmountValue `json:"estimatedNetworkFees,omitempty"`
	EstimatedNetworkFeesInFiatCurrency *AmountValue `json:"estimatedNetworkFeesInFiatCurrency,omitempty"`
	EstimatedTimeInMs                  int          `json:"estimatedTimeInMs"`
	Format                             string       `json:"format"`
	From                               string       `json:"from"`
	GasLimit                           string       `json:"gasLimit,omitempty"`
	InputState                         TokenState   `json:"inputState"`
	MaxFeePerGas                       string       `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas               string       `json:"maxPriorityFeePerGas,omitempty"`
	Nonce                              string       `json:"nonce,omitempty"`
	OutputState                        TokenState   `json:"outputState"`
	To                                 string       `json:"to"`
	TxRpcUrl                           string       `json:"txRpcUrl"`
	Type                               string       `json:"type"`
}

// Intent represents a swap intent with operations
type Intent struct {
	EstimatedProtocolFeesInFiatCurrency AmountValue `json:"estimatedProtocolFeesInFiatCurrency"`
	InputState                          TokenState  `json:"inputState"`
	IntentOperations                    []Operation `json:"intentOperations"`
	OutputState                         TokenState  `json:"outputState"`
}

// GetOperationsToSwapResponse represents the complete response from the orby_getOperationsToSwap call
type GetOperationsToSwapResponse struct {
	AggregateEstimatedTimeInMs          int         `json:"aggregateEstimatedTimeInMs"`
	AggregateNetworkFeeInFiatCurrency   AmountValue `json:"aggregateNetworkFeeInFiatCurrency"`
	AggregateOperationFeeInFiatCurrency AmountValue `json:"aggregateOperationFeeInFiatCurrency"`
	InputState                          TokenState  `json:"inputState"`
	Intents                             []Intent    `json:"intents"`
	PrimaryOperationPreconditions       TokenState  `json:"primaryOperationPreconditions"`
	Status                              string      `json:"status"`
}

// AccountParams represents the parameters for an account in the account cluster
type AccountParams struct {
	VMType      string `json:"vmType"`
	Address     string `json:"address"`
	AccountType string `json:"accountType"`
}

// CreateAccountClusterParams represents the parameters for orby_createAccountCluster
type CreateAccountClusterParams struct {
	Accounts []AccountParams `json:"accounts"`
}

// AccountClusterAccount represents an account in the account cluster response
type AccountClusterAccount struct {
	AccountType string `json:"accountType"`
	Address     string `json:"address"`
	ChainId     string `json:"chainId"`
	VMType      string `json:"vmType"`
}

// AccountClusterResponse represents the response from orby_createAccountCluster
type AccountClusterResponse struct {
	AccountClusterId string                  `json:"accountClusterId"`
	Accounts         []AccountClusterAccount `json:"accounts"`
	Id               string                  `json:"id"`
}

// GetVirtualNodeRpcUrlParams represents the parameters for orby_getVirtualNodeRpcUrl
type GetVirtualNodeRpcUrlParams struct {
	AccountClusterId         string `json:"accountClusterId"`
	ChainId                  string `json:"chainId"`
	EntrypointAccountAddress string `json:"entrypointAccountAddress"`
}

// VirtualNodeRpcUrlResponse represents the response from orby_getVirtualNodeRpcUrl
type VirtualNodeRpcUrlResponse struct {
	VirtualNodeRpcUrl string `json:"virtualNodeRpcUrl"`
}
