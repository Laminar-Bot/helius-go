package helius

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// Network represents a Solana network.
type Network string

const (
	// Mainnet is the Solana mainnet.
	Mainnet Network = "mainnet"
	// Devnet is the Solana devnet for testing.
	Devnet Network = "devnet"
)

const (
	// DefaultMainnetAPIURL is the default Helius API URL for mainnet.
	DefaultMainnetAPIURL = "https://api.helius.xyz/v0"
	// DefaultDevnetAPIURL is the default Helius API URL for devnet.
	DefaultDevnetAPIURL = "https://api-devnet.helius.xyz/v0"

	// DefaultMainnetRPCURL is the default Helius RPC URL for mainnet.
	DefaultMainnetRPCURL = "https://mainnet.helius-rpc.com"
	// DefaultDevnetRPCURL is the default Helius RPC URL for devnet.
	DefaultDevnetRPCURL = "https://devnet.helius-rpc.com"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 10 * time.Second
	// DefaultMaxRetries is the default maximum number of retries.
	DefaultMaxRetries = 3
	// DefaultRetryWaitMin is the minimum wait time between retries.
	DefaultRetryWaitMin = 500 * time.Millisecond
	// DefaultRetryWaitMax is the maximum wait time between retries.
	DefaultRetryWaitMax = 5 * time.Second
)

// Logger interface for optional logging.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// noopLogger discards all log messages.
type noopLogger struct{}

func (noopLogger) Debug(_ string, _ ...interface{}) {}
func (noopLogger) Info(_ string, _ ...interface{})  {}
func (noopLogger) Warn(_ string, _ ...interface{})  {}
func (noopLogger) Error(_ string, _ ...interface{}) {}

// config holds client configuration.
type config struct {
	network      Network
	apiURL       string
	rpcURL       string
	timeout      time.Duration
	maxRetries   int
	retryWaitMin time.Duration
	retryWaitMax time.Duration
	httpClient   *http.Client
	logger       Logger
}

// Option configures the client.
type Option func(*config)

// WithNetwork sets the Solana network (Mainnet or Devnet).
func WithNetwork(network Network) Option {
	return func(c *config) {
		c.network = network
	}
}

// WithAPIURL sets a custom API base URL.
func WithAPIURL(url string) Option {
	return func(c *config) {
		c.apiURL = url
	}
}

// WithRPCURL sets a custom RPC base URL.
func WithRPCURL(url string) Option {
	return func(c *config) {
		c.rpcURL = url
	}
}

// WithTimeout sets the HTTP request timeout.
func WithTimeout(d time.Duration) Option {
	return func(c *config) {
		c.timeout = d
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries(n int) Option {
	return func(c *config) {
		c.maxRetries = n
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *config) {
		c.httpClient = client
	}
}

// WithLogger sets a custom logger.
func WithLogger(l Logger) Option {
	return func(c *config) {
		c.logger = l
	}
}

// Client is the Helius API client.
type Client struct {
	apiKey     string
	apiURL     string
	rpcURL     string
	httpClient *http.Client
	logger     Logger
}

// NewClient creates a new Helius API client.
func NewClient(apiKey string, opts ...Option) (*Client, error) {
	if apiKey == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "API key is required",
			Path:       "client",
		}
	}

	cfg := &config{
		network:      Mainnet,
		timeout:      DefaultTimeout,
		maxRetries:   DefaultMaxRetries,
		retryWaitMin: DefaultRetryWaitMin,
		retryWaitMax: DefaultRetryWaitMax,
		logger:       noopLogger{},
	}

	for _, opt := range opts {
		opt(cfg)
	}

	// Set URLs based on network if not explicitly provided
	if cfg.apiURL == "" {
		if cfg.network == Devnet {
			cfg.apiURL = DefaultDevnetAPIURL
		} else {
			cfg.apiURL = DefaultMainnetAPIURL
		}
	}
	if cfg.rpcURL == "" {
		if cfg.network == Devnet {
			cfg.rpcURL = DefaultDevnetRPCURL
		} else {
			cfg.rpcURL = DefaultMainnetRPCURL
		}
	}

	var httpClient *http.Client
	if cfg.httpClient != nil {
		httpClient = cfg.httpClient
	} else {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = cfg.maxRetries
		retryClient.RetryWaitMin = cfg.retryWaitMin
		retryClient.RetryWaitMax = cfg.retryWaitMax
		retryClient.Logger = nil // Disable default logging

		retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
			if ctx.Err() != nil {
				return false, ctx.Err()
			}
			if err != nil {
				return true, err
			}
			if resp.StatusCode == http.StatusTooManyRequests {
				return true, nil
			}
			if resp.StatusCode >= 500 {
				return true, nil
			}
			return false, nil
		}

		httpClient = retryClient.StandardClient()
		httpClient.Timeout = cfg.timeout
	}

	return &Client{
		apiKey:     apiKey,
		apiURL:     cfg.apiURL,
		rpcURL:     cfg.rpcURL,
		httpClient: httpClient,
		logger:     cfg.logger,
	}, nil
}

// RPCURL returns the RPC URL with API key for use with solana-go.
func (c *Client) RPCURL() string {
	return fmt.Sprintf("%s/?api-key=%s", c.rpcURL, c.apiKey)
}

// doRequest performs an HTTP request and returns the response body.
func (c *Client) doRequest(ctx context.Context, method, path string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s?api-key=%s", c.apiURL, path, c.apiKey)

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logger.Debug("making request", "method", method, "path", path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("api error", "status", resp.StatusCode, "path", path, "body", string(respBody))
		return nil, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
			Path:       path,
		}
	}

	return respBody, nil
}

// doGet performs an HTTP GET request.
func (c *Client) doGet(ctx context.Context, path string) ([]byte, error) {
	return c.doRequest(ctx, http.MethodGet, path, nil)
}

// doPost performs an HTTP POST request with JSON body.
func (c *Client) doPost(ctx context.Context, path string, body interface{}) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}
	return c.doRequest(ctx, http.MethodPost, path, io.NopCloser(io.Reader(jsonReaderFrom(jsonBody))))
}

// jsonReaderFrom creates a reader from JSON bytes.
func jsonReaderFrom(data []byte) io.Reader {
	return &jsonReader{data: data}
}

type jsonReader struct {
	data []byte
	pos  int
}

func (r *jsonReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
