package helius

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// TokenHolder represents a holder of a token.
type TokenHolder struct {
	// Owner is the wallet address of the token holder.
	Owner string `json:"owner"`

	// TokenAccount is the associated token account address.
	TokenAccount string `json:"tokenAccount"`

	// Balance is the raw token balance (without decimals).
	Balance int64 `json:"balance"`

	// Decimals is the token's decimal places.
	Decimals int `json:"decimals"`
}

// TokenHoldersPage represents a paginated response of token holders.
type TokenHoldersPage struct {
	// Total is the total number of holders.
	Total int `json:"total"`

	// Limit is the page size.
	Limit int `json:"limit"`

	// Cursor is the pagination cursor for the next page.
	Cursor string `json:"cursor,omitempty"`

	// TokenHolders is the list of token holders.
	TokenHolders []TokenHolder `json:"token_holders"`
}

// GetTokenHoldersOptions configures the token holders request.
type GetTokenHoldersOptions struct {
	// Cursor for pagination (from previous response).
	Cursor string `json:"cursor,omitempty"`

	// Limit is the maximum number of holders to return (default: 1000, max: 10000).
	Limit int `json:"limit,omitempty"`
}

// GetTokenHolders fetches the holders of a token.
//
// Example:
//
//	holders, err := client.GetTokenHolders(ctx, "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Total USDC holders: %d\n", holders.Total)
//	for _, h := range holders.TokenHolders {
//	    fmt.Printf("  %s: %d\n", h.Owner, h.Balance)
//	}
func (c *Client) GetTokenHolders(ctx context.Context, mint string, opts *GetTokenHoldersOptions) (*TokenHoldersPage, error) {
	if mint == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "mint address is required",
			Path:       "/token-holders",
		}
	}

	reqBody := map[string]interface{}{
		"mint": mint,
	}

	if opts != nil {
		if opts.Cursor != "" {
			reqBody["cursor"] = opts.Cursor
		}
		if opts.Limit > 0 {
			reqBody["limit"] = opts.Limit
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, err := c.doRequest(ctx, "POST", "/token-holders", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var page TokenHoldersPage
	if err := json.Unmarshal(body, &page); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("fetched token holders",
		"mint", mint,
		"total", page.Total,
		"returned", len(page.TokenHolders),
	)

	return &page, nil
}

// GetAllTokenHolders fetches all holders of a token, handling pagination automatically.
//
// Warning: This can be slow and memory-intensive for tokens with many holders.
// Consider using GetTokenHolders with pagination for large tokens.
func (c *Client) GetAllTokenHolders(ctx context.Context, mint string) ([]TokenHolder, error) {
	var allHolders []TokenHolder
	var cursor string

	for {
		opts := &GetTokenHoldersOptions{
			Cursor: cursor,
			Limit:  10000, // Max per page
		}

		page, err := c.GetTokenHolders(ctx, mint, opts)
		if err != nil {
			return nil, err
		}

		allHolders = append(allHolders, page.TokenHolders...)

		if page.Cursor == "" || len(page.TokenHolders) == 0 {
			break
		}

		cursor = page.Cursor
	}

	c.logger.Info("fetched all token holders",
		"mint", mint,
		"total", len(allHolders),
	)

	return allHolders, nil
}

// TopHolderStats calculates statistics about top token holders.
type TopHolderStats struct {
	// TotalHolders is the total number of holders.
	TotalHolders int

	// TopHolders is the list of top holders.
	TopHolders []TokenHolder

	// TopHoldersBalance is the combined balance of top holders.
	TopHoldersBalance int64

	// TopHoldersPercent is the percentage of supply held by top holders.
	TopHoldersPercent float64

	// TotalSupply is the total token supply held by all queried holders.
	TotalSupply int64
}

// CalculateTopHolderStats calculates concentration statistics for token holders.
//
// Example:
//
//	holders, _ := client.GetTokenHolders(ctx, mint, &helius.GetTokenHoldersOptions{Limit: 100})
//	stats := helius.CalculateTopHolderStats(holders.TokenHolders, 10)
//	fmt.Printf("Top 10 hold %.2f%%\n", stats.TopHoldersPercent)
func CalculateTopHolderStats(holders []TokenHolder, topN int) *TopHolderStats {
	if len(holders) == 0 {
		return &TopHolderStats{}
	}

	// Calculate total supply
	var totalSupply int64
	for _, h := range holders {
		totalSupply += h.Balance
	}

	// Get top N holders (assuming sorted by balance descending)
	topCount := topN
	if topCount > len(holders) {
		topCount = len(holders)
	}

	var topBalance int64
	topHolders := make([]TokenHolder, topCount)
	// Copy top holders - bounds are guaranteed by topCount check above
	copy(topHolders, holders[:topCount])
	for _, h := range topHolders {
		topBalance += h.Balance
	}

	var topPercent float64
	if totalSupply > 0 {
		topPercent = float64(topBalance) / float64(totalSupply) * 100
	}

	return &TopHolderStats{
		TotalHolders:      len(holders),
		TopHolders:        topHolders,
		TopHoldersBalance: topBalance,
		TopHoldersPercent: topPercent,
		TotalSupply:       totalSupply,
	}
}
