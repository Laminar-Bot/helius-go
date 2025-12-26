# helius-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Laminar-Bot/helius-go.svg)](https://pkg.go.dev/github.com/Laminar-Bot/helius-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Laminar-Bot/helius-go)](https://goreportcard.com/report/github.com/Laminar-Bot/helius-go)
[![CI](https://github.com/Laminar-Bot/helius-go/actions/workflows/ci.yml/badge.svg)](https://github.com/Laminar-Bot/helius-go/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Go client for [Helius](https://helius.dev) proprietary APIs on Solana.

**Note:** For standard Solana RPC operations (getBalance, sendTransaction, etc.), use [solana-go](https://github.com/gagliardetto/solana-go) directly with a Helius RPC endpoint. This package only wraps Helius-proprietary APIs.

## Features

- 🎨 **DAS API** - Query NFTs, compressed NFTs, and digital assets
- 🪝 **Webhooks** - Create, manage, and validate webhook signatures
- 👥 **Token Holders** - Get holder information and concentration stats
- ⚡ **Priority Fees** - Estimate optimal transaction fees
- 🔄 **Network Support** - Mainnet and Devnet
- 🔁 **Automatic Retries** - Built-in retry logic for 429/5xx errors

## Installation

```bash
go get github.com/Laminar-Bot/helius-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/Laminar-Bot/helius-go"
)

func main() {
    // Create client
    client, err := helius.NewClient("your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // Get asset metadata via DAS API
    asset, err := client.GetAsset(context.Background(), "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Asset: %s (Interface: %s)\n", asset.ID, asset.Interface)
}
```

## Configuration

```go
// Default mainnet client
client, _ := helius.NewClient("your-api-key")

// Devnet client
client, _ := helius.NewClient("your-api-key",
    helius.WithNetwork(helius.Devnet),
)

// Custom configuration
client, _ := helius.NewClient("your-api-key",
    helius.WithNetwork(helius.Mainnet),
    helius.WithTimeout(30*time.Second),
    helius.WithMaxRetries(5),
    helius.WithLogger(myLogger),
)

// Get RPC URL for use with solana-go
rpcURL := client.RPCURL()
// Returns: https://mainnet.helius-rpc.com/?api-key=your-api-key
```

## DAS API (Digital Asset Standard)

```go
// Get a single asset
asset, err := client.GetAsset(ctx, "mint-address")

// Get all assets owned by a wallet
assets, err := client.GetAssetsByOwner(ctx, "owner-wallet", &helius.AssetsByOwnerOptions{
    Limit:             100,
    ShowFungible:      true,
    ShowNativeBalance: true,
})

// Search for assets
results, err := client.SearchAssets(ctx, &helius.SearchAssetsOptions{
    OwnerAddress: "wallet-address",
    GroupKey:     "collection",
    GroupValue:   "collection-mint",
})

// Batch fetch multiple assets
assets, err := client.GetAssetBatch(ctx, []string{"mint1", "mint2", "mint3"})
```

## Webhooks

```go
// Create a webhook
webhook, err := client.CreateWebhook(ctx, &helius.CreateWebhookRequest{
    WebhookURL:       "https://your-server.com/webhook",
    TransactionTypes: []helius.TransactionType{helius.TransactionTypeSwap},
    AccountAddresses: []string{"wallet-to-monitor"},
})

// List all webhooks
webhooks, err := client.ListWebhooks(ctx)

// Update a webhook
updated, err := client.UpdateWebhook(ctx, webhookID, &helius.UpdateWebhookRequest{
    AccountAddresses: []string{"new-wallet"},
})

// Delete a webhook
err := client.DeleteWebhook(ctx, webhookID)
```

### Webhook Signature Validation

**Critical for production**: Always validate webhook signatures to prevent spoofing.

```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("X-Helius-Signature")

    if !helius.ValidateWebhookSignature(body, signature, "your-webhook-secret") {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Parse the webhook event
    events, err := helius.ParseWebhookEvents(body)
    if err != nil {
        http.Error(w, "Invalid payload", http.StatusBadRequest)
        return
    }

    for _, event := range events {
        fmt.Printf("Transaction: %s (Type: %s)\n", event.Signature, event.Type)
    }

    w.WriteHeader(http.StatusOK)
}
```

## Priority Fees

```go
// Estimate priority fee for accounts
estimate, err := client.GetPriorityFeeEstimate(ctx,
    []string{"JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB"},
    &helius.GetPriorityFeeOptions{
        PriorityLevel: helius.PriorityMedium,
    },
)
fmt.Printf("Recommended fee: %.0f microlamports/CU\n", estimate.PriorityFeeEstimate)

// Get all priority levels
estimate, err := client.GetPriorityFeeEstimate(ctx, accounts,
    &helius.GetPriorityFeeOptions{
        IncludeAllPriorityFeeLevels: true,
    },
)
fmt.Printf("Low: %.0f, Medium: %.0f, High: %.0f\n",
    estimate.PriorityFeeLevels.Low,
    estimate.PriorityFeeLevels.Medium,
    estimate.PriorityFeeLevels.High,
)

// Calculate total fee in lamports
fee := helius.CalculatePriorityFee(200_000, estimate.PriorityFeeEstimate)
fmt.Printf("Total priority fee: %d lamports\n", fee)
```

## Token Holders

```go
// Get paginated token holders
page, err := client.GetTokenHolders(ctx, "token-mint", &helius.GetTokenHoldersOptions{
    Limit: 100,
})
fmt.Printf("Total holders: %d\n", page.Total)

// Get ALL token holders (handles pagination automatically)
// Warning: Can be slow for tokens with many holders
holders, err := client.GetAllTokenHolders(ctx, "token-mint")

// Calculate top holder concentration (rug pull detection)
stats := helius.CalculateTopHolderStats(holders, 10)
fmt.Printf("Top 10 holders own %.2f%% of supply\n", stats.TopHoldersPercent)
```

## Error Handling

```go
asset, err := client.GetAsset(ctx, "invalid-mint")
if err != nil {
    apiErr, ok := helius.IsAPIError(err)
    if ok {
        switch {
        case apiErr.IsNotFound():
            fmt.Println("Asset not found")
        case apiErr.IsRateLimited():
            fmt.Println("Rate limited, retry later")
        case apiErr.IsUnauthorized():
            fmt.Println("Invalid API key")
        default:
            fmt.Printf("API error %d: %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
    return err
}
```

## API Coverage

| Category | Method | Status |
|----------|--------|--------|
| DAS | GetAsset | ✅ |
| DAS | GetAssetsByOwner | ✅ |
| DAS | SearchAssets | ✅ |
| DAS | GetAssetBatch | ✅ |
| Webhooks | CreateWebhook | ✅ |
| Webhooks | GetWebhook | ✅ |
| Webhooks | ListWebhooks | ✅ |
| Webhooks | UpdateWebhook | ✅ |
| Webhooks | DeleteWebhook | ✅ |
| Webhooks | ValidateWebhookSignature | ✅ |
| Priority Fees | GetPriorityFeeEstimate | ✅ |
| Priority Fees | GetPriorityFeeEstimateForTransaction | ✅ |
| Token Holders | GetTokenHolders | ✅ |
| Token Holders | GetAllTokenHolders | ✅ |

## Using with solana-go

This package is designed to work alongside [solana-go](https://github.com/gagliardetto/solana-go):

```go
import (
    "github.com/gagliardetto/solana-go/rpc"
    "github.com/Laminar-Bot/helius-go"
)

// Create Helius client for proprietary APIs
heliusClient, _ := helius.NewClient("your-api-key")

// Create Solana RPC client with Helius endpoint
solanaClient := rpc.New(heliusClient.RPCURL())

// Use solana-go for standard RPC
balance, _ := solanaClient.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)

// Use helius-go for proprietary APIs
asset, _ := heliusClient.GetAsset(ctx, mintAddress)
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [Helius Documentation](https://docs.helius.dev)
- [Helius Dashboard](https://dashboard.helius.dev)
- [Go Package Documentation](https://pkg.go.dev/github.com/Laminar-Bot/helius-go)
