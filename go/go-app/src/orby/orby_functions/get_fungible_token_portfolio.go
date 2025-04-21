package orbyfunctions

import (
	"encoding/json"
	"fmt"
	"log"

	"go-app/src/orby"
)

type GetFungibleTokenPortfolio struct {
	VirtualNodeProvider orby.OrbyClient
	AccountClusterId    string
}

func NewGetFungibleTokenPortfolio(client orby.OrbyClient, accountClusterId string) *GetFungibleTokenPortfolio {
	return &GetFungibleTokenPortfolio{
		VirtualNodeProvider: client,
		AccountClusterId:    accountClusterId,
	}
}

func (g *GetFungibleTokenPortfolio) Run() error {
	// 1. Call operation
	fmt.Println("\n[INFO] calling GetFungibleTokenPortfolio...")
	result, err := g.VirtualNodeProvider.GetFungibleTokenPortfolio(
		g.AccountClusterId)
	if err != nil {
		log.Printf("[ERROR] Error getting fungible token portfolio: %v", err)
	}

	// 2. Parse the response into our structured type
	var response orby.GetFungibleTokenPortfolioResponse
	if err := json.Unmarshal(result, &response); err != nil {
		log.Printf("[ERROR] Error parsing orby_GetFungibleTokenPortfolio response: %v", err)
		// Try to display raw response
		var rawResponse any
		if json.Unmarshal(result, &rawResponse) == nil {
			fmt.Printf("          Raw response: %v\n", rawResponse)
		}
		return err
	}

	// 3. Print result
	fmt.Printf("\n[INFO] Fungible Token Portfolio Response:\n")
	for _, balance := range response.FungibleTokenBalances {
		fmt.Println("\nStandardized Token ID:", balance.StandardizedTokenId)
		fmt.Println("Total:", balance.Total.Amount)

		// Loop through nested token balances
		for i, tokenBalance := range balance.TokenBalances {
			fmt.Printf("\n  Token Balance %d:\n", i)
			fmt.Println("    Token Address:", tokenBalance.Token.Address)
			fmt.Println("    Amount:", tokenBalance.Amount)
		}
		for i, tokenBalance := range balance.TokenBalancesOnChains {
			fmt.Printf("\n  Token Balance On Chain %d:\n", i)
			fmt.Println("    Token Address:", tokenBalance.Token.Address)
			fmt.Println("    Amount:", tokenBalance.Amount)
		}
	}

	return nil
}
