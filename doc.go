// Package helius provides a Go client for Helius APIs on Solana.
//
// Helius provides enhanced Solana infrastructure including:
//   - DAS API (Digital Asset Standard) for NFT and token metadata
//   - Webhook management for real-time transaction monitoring
//   - Priority fee estimation for transaction optimization
//   - Token holder analysis
//
// Note: For standard Solana RPC calls (getBalance, sendTransaction, etc.),
// use github.com/gagliardetto/solana-go directly with a Helius RPC endpoint.
// This package only wraps Helius-proprietary APIs.
//
// # Quick Start
//
//	client, err := helius.NewClient("your-api-key")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get assets owned by a wallet
//	assets, err := client.GetAssetsByOwner(ctx, "wallet-address", nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d assets\n", assets.Total)
//
// # Configuration
//
//	client, err := helius.NewClient("your-api-key",
//	    helius.WithTimeout(30 * time.Second),
//	    helius.WithNetwork(helius.Devnet),
//	)
//
// # Webhook Signature Validation
//
// When receiving webhooks, always validate the signature:
//
//	isValid := helius.ValidateWebhookSignature(requestBody, signatureHeader, webhookSecret)
//	if !isValid {
//	    http.Error(w, "invalid signature", http.StatusUnauthorized)
//	    return
//	}
//
// # Networks
//
// Helius supports multiple Solana networks:
//
//	helius.Mainnet  // Production (default)
//	helius.Devnet   // Development/testing
//
// # Error Handling
//
// All API errors are returned as *APIError with helpful methods:
//
//	asset, err := client.GetAsset(ctx, "invalid-id")
//	if err != nil {
//	    if apiErr, ok := helius.IsAPIError(err); ok {
//	        if apiErr.IsNotFound() {
//	            // Asset doesn't exist
//	        }
//	        if apiErr.IsRateLimited() {
//	            // Slow down requests
//	        }
//	    }
//	}
package helius
