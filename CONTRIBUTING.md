# Contributing to helius-go

Thank you for your interest in contributing to helius-go! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/helius-go.git`
3. Create a branch: `git checkout -b feature/your-feature`
4. Make your changes
5. Run tests: `go test -cover ./...`
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.21 or later
- golangci-lint (for linting)

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...

# Run specific test
go test -run TestValidateWebhookSignature ./...
```

### Linting

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run
```

## Code Guidelines

### API Client Patterns

1. **All API methods take `context.Context` as the first parameter**

```go
// Good
func (c *Client) GetAsset(ctx context.Context, id string) (*Asset, error)

// Bad
func (c *Client) GetAsset(id string) (*Asset, error)
```

2. **Return custom `*APIError` for HTTP errors**

```go
if resp.StatusCode >= 400 {
    return nil, &APIError{
        StatusCode: resp.StatusCode,
        Message:    string(body),
        Path:       path,
    }
}
```

3. **Use functional options for configuration**

```go
client, err := helius.NewClient(apiKey,
    helius.WithNetwork(helius.Devnet),
    helius.WithTimeout(30*time.Second),
)
```

### Testing Standards

1. **Minimum 80% code coverage required**
2. **Use `httptest.Server` for mocking HTTP responses**
3. **Table-driven tests for multiple cases**
4. **Test both success and error paths**

Example test structure:

```go
func TestGetAsset(t *testing.T) {
    t.Run("successful get", func(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            json.NewEncoder(w).Encode(Asset{ID: "test-id"})
        }))
        defer server.Close()

        client, _ := NewClient("test-key", WithAPIURL(server.URL))
        asset, err := client.GetAsset(context.Background(), "test-id")

        if err != nil {
            t.Fatalf("unexpected error: %v", err)
        }
        if asset.ID != "test-id" {
            t.Errorf("ID = %s, want test-id", asset.ID)
        }
    })

    t.Run("empty id returns error", func(t *testing.T) {
        client, _ := NewClient("test-key")
        _, err := client.GetAsset(context.Background(), "")
        if err == nil {
            t.Error("expected error for empty id")
        }
    })
}
```

### Documentation

- All exported types and functions must have doc comments
- Include code examples in doc comments where helpful
- Use `// Example:` blocks for usage examples

```go
// GetAsset fetches a single asset by its ID (mint address).
//
// Example:
//
//     asset, err := client.GetAsset(ctx, "mint-address")
//     if err != nil {
//         log.Fatal(err)
//     }
//     fmt.Println(asset.ID)
func (c *Client) GetAsset(ctx context.Context, id string) (*Asset, error)
```

## Pull Request Process

1. Update documentation if needed
2. Ensure all tests pass
3. Ensure code coverage is at least 80%
4. Update README.md if adding new features
5. Wait for code review

## Security

### Webhook Signature Validation

When contributing to webhook-related code:

- **Always use constant-time comparison** for signature validation
- Never log secrets or API keys
- Test edge cases (empty strings, nil values)

### Sensitive Data

- Never commit API keys or secrets
- Use environment variables for testing with real APIs
- Mark test functions that require real API access with build tags

## Questions?

Open an issue for questions or discussion about potential contributions.
