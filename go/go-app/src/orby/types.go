// types.go defines data structures used throughout the application.
package orby

import (
	"math/big"
)

// ************************************** Common **************************************

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
type CurrencyAmount struct {
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
type State struct {
	FungibleTokenAmounts     []TokenAmount `json:"fungibleTokenAmounts"`
	NonFungibleTokenAmounts  []TokenAmount `json:"nonFungibleTokenAmounts"`
	SemiFungibleTokenAmounts []TokenAmount `json:"semiFungibleTokenAmounts"`
}

// Operation represents a blockchain operation/transaction
type Operation struct {
	ChainId                            string          `json:"chainId"`
	Data                               string          `json:"data"`
	EstimatedNetworkFees               *CurrencyAmount `json:"estimatedNetworkFees,omitempty"`
	EstimatedNetworkFeesInFiatCurrency *CurrencyAmount `json:"estimatedNetworkFeesInFiatCurrency,omitempty"`
	EstimatedTimeInMs                  int             `json:"estimatedTimeInMs"`
	Format                             string          `json:"format"`
	From                               string          `json:"from"`
	GasLimit                           string          `json:"gasLimit,omitempty"`
	InputState                         State           `json:"inputState"`
	MaxFeePerGas                       string          `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas               string          `json:"maxPriorityFeePerGas,omitempty"`
	Nonce                              string          `json:"nonce,omitempty"`
	OutputState                        State           `json:"outputState"`
	To                                 string          `json:"to"`
	TxRpcUrl                           string          `json:"txRpcUrl"`
	Type                               string          `json:"type"`
}

// Intent represents a swap intent with operations
type Intent struct {
	EstimatedProtocolFeesInFiatCurrency CurrencyAmount `json:"estimatedProtocolFeesInFiatCurrency"`
	InputState                          State          `json:"inputState"`
	IntentOperations                    []Operation    `json:"intentOperations"`
	OutputState                         State          `json:"outputState"`
}

// ErrorResponse represents the response if an error occurs
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// OperationSet represents the complete response from any orby_getOperation calls
type OperationSet struct {
	AggregateEstimatedTimeInMs          int            `json:"aggregateEstimatedTimeInMs"`
	AggregateNetworkFeeInFiatCurrency   CurrencyAmount `json:"aggregateNetworkFeeInFiatCurrency"`
	AggregateOperationFeeInFiatCurrency CurrencyAmount `json:"aggregateOperationFeeInFiatCurrency"`
	InputState                          State          `json:"inputState"`
	OutputState                         State          `json:"outputState"`
	Intents                             []Intent       `json:"intents"`
	PrimaryOperationPreconditions       State          `json:"primaryOperationPreconditions"`
	PrimaryOperation                    Operation      `json:"operation"`
	Status                              string         `json:"status"`
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

// StandardizedBalance represents balances for a standardized token id
type StandardizedBalance struct {
	Typename              *string         `json:"__typename,omitempty"`
	StandardizedTokenId   string          `json:"standardizedTokenId"`
	TokenBalances         []TokenAmount   `json:"tokenBalances"`
	TokenBalancesOnChains []TokenAmount   `json:"tokenBalancesOnChains"`
	Total                 CurrencyAmount  `json:"total"`
	TotalInFiatCurrency   *CurrencyAmount `json:"totalInFiatCurrency,omitempty"`
	TotalValueInFiat      *CurrencyAmount `json:"totalValueInFiat,omitempty"`
}

// ************************************** Function Params/Responses **************************************

// GetOperationsToExecuteTransactionParams represents the parameters for orby_getOperationsToExecuteTransaction
type GetOperationsToExecuteTransactionParams struct {
	AccountClusterId string `json:"accountClusterId"`
	Data             string `json:"data"`
	To               string `json:"to"`
}

// GetOperationsToSignTypedDataParams represents the parameters for orby_getOperationsToSignTypedData
type GetOperationsToSignTypedDataParams struct {
	AccountClusterId string `json:"accountClusterId"`
	Data             string `json:"data"`
}

// GetFungibleTokenPortfolioParams represents the parameters for orby_getFungibleTokenPortfolio
type GetFungibleTokenPortfolioParams struct {
	AccountClusterId string `json:"accountClusterId"`
}

// GetFungibleTokenPortfolioResponse represents the response for orby_getFungibleTokenPortfolio
type GetFungibleTokenPortfolioResponse struct {
	FungibleTokenBalances []StandardizedBalance `json:"fungibleTokenBalances"`
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

// StandardizedTokenIdsResponse represents the response from orby_getStandardizedTokenIds
type StandardizedTokenIdsResponse struct {
	StandardizedTokenIds []string `json:"standardizedTokenIds"`
}

// TokenSource represents a token
type TokenSource struct {
	ChainID string `json:"chainId"`
	Address string `json:"address,omitempty"`
}

// InputSwapParam represents the input tokens for orby_getOperationsToSwap
type InputSwapParam struct {
	StandardizedTokenId string        `json:"standardizedTokenId"`
	Amount              string        `json:"amount,omitempty"`
	TokenSources        []TokenSource `json:"tokenSources,omitempty"`
}

// OutputSwapParam represents the output tokens for orby_getOperationsToSwap
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

// ************************************** Setup **************************************

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

// OrbyInstanceResponse represents the response from orby_createInstance
type OrbyInstanceResponse struct {
	Success                bool   `json:"success"`
	OrbyInstancePrivateUrl string `json:"orbyInstancePrivateUrl"`
	OrbyInstancePublicUrl  string `json:"orbyInstancePublicUrl"`
}

// ************************************** Permit2 **************************************

// TokenPermissions matches { token: address, amount: uint256 }
type TokenPermissions struct {
	Token  string   `json:"token"`
	Amount *big.Int `json:"amount"`
}

// PermitTransferFrom matches the primaryType with nested types
type PermitTransferFrom struct {
	Permitted TokenPermissions `json:"permitted"`
	Spender   string           `json:"spender"`
	Nonce     *big.Int         `json:"nonce"`
	Deadline  *big.Int         `json:"deadline"`
}

// EIP712Domain matches { name: string, chainId: uint256, verifyingContract: address }
type EIP712Domain struct {
	Name              string   `json:"name"`
	ChainId           *big.Int `json:"chainId"`
	VerifyingContract string   `json:"verifyingContract"`
}

// Permit2Object is the root structure
type Permit2Object struct {
	Types       map[string][]EIP712Type `json:"types"`
	Domain      EIP712Domain            `json:"domain"`
	PrimaryType string                  `json:"primaryType"`
	Message     PermitTransferFrom      `json:"message"`
}

// EIP712Type defines a typed field (name and type)
type EIP712Type struct {
	Name string `json:"name"`
	Type string `json:"type"`
}
