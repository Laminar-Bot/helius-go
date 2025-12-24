# helius-go

[![Go Reference](https://pkg.go.dev/badge/github.com/Laminar-Bot/helius-go.svg)](https://pkg.go.dev/github.com/Laminar-Bot/helius-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/Laminar-Bot/helius-go)](https://goreportcard.com/report/github.com/Laminar-Bot/helius-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A comprehensive Go client for the [Helius](https://helius.dev) API, providing access to enhanced Solana RPC, webhooks, DAS (Digital Asset Standard) API, and more.

## Features

- 🚀 **Enhanced RPC** - Standard Solana RPC with Helius improvements
- 🪝 **Webhooks** - Create, manage, and validate webhook signatures
- 🎨 **DAS API** - Query NFTs, compressed NFTs, and token metadata
- 👥 **Token Holders** - Get holder information for any token
- ⚡ **Priority Fees** - Estimate and apply priority fees
- 🔄 **Transaction Parsing** - Parse enhanced transaction data

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
    client := helius.NewClient(helius.Config{
        APIKey: "your-api-key",
    })

    // Get SOL balance
    balance, err := client.GetBalance(context.Background(), "YourWalletAddress...")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Balance: %d lamports\n", balance)

    // Get token metadata via DAS
    asset, err := client.GetAsset(context.Background(), "TokenMintAddress...")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Token: %s (%s)\n", asset.Content.Metadata.Name, asset.Content.Metadata.Symbol)
}
```

## Webhook Signature Validation
```go
func handleWebhook(w http.ResponseWriter, r *http.Request) {
    body, _ := io.ReadAll(r.Body)
    signature := r.Header.Get("Authorization")

    if !helius.ValidateWebhookSignature(body, signature, "your-webhook-secret") {
        http.Error(w, "Invalid signature", http.StatusUnauthorized)
        return
    }

    // Process webhook...
}
```

## API Coverage

| Category | Endpoint | Status |
|----------|----------|--------|
| RPC | getBalance | ✅ |
| RPC | getTokenAccountsByOwner | ✅ |
| RPC | sendTransaction | ✅ |
| RPC | getTransaction | ✅ |
| DAS | getAsset | ✅ |
| DAS | getAssetsByOwner | ✅ |
| DAS | searchAssets | ✅ |
| Webhooks | createWebhook | ✅ |
| Webhooks | editWebhook | ✅ |
| Webhooks | deleteWebhook | ✅ |
| Enhanced | getTokenHolders | ✅ |
| Enhanced | getPriorityFeeEstimate | ✅ |

## Configuration
```go
client := helius.NewClient(helius.Config{
    APIKey:     "your-api-key",
    RPCURL:     "https://mainnet.helius-rpc.com", // optional, defaults to mainnet
    TimeoutSec: 30,                                // optional, defaults to 30
})
```

## Error Handling
```go
balance, err := client.GetBalance(ctx, address)
if err != nil {
    var apiErr *helius.APIError
    if errors.As(err, &apiErr) {
        fmt.Printf("API error: %d - %s\n", apiErr.Code, apiErr.Message)
    }
    return err
}
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) first.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Links

- [Helius Documentation](https://docs.helius.dev)
- [Helius Dashboard](https://dashboard.helius.dev)
- [Go Package Documentation](https://pkg.go.dev/github.com/Laminar-Bot/helius-go)
