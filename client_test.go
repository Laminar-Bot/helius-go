package helius

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("valid api key", func(t *testing.T) {
		client, err := NewClient("test-api-key")
		if err != nil {
			t.Fatalf("NewClient returned error: %v", err)
		}
		if client == nil {
			t.Fatal("NewClient returned nil client")
		}
		// Default should be mainnet
		if !strings.Contains(client.apiURL, "api.helius.xyz") {
			t.Errorf("apiURL should contain mainnet URL, got %s", client.apiURL)
		}
	})

	t.Run("empty api key", func(t *testing.T) {
		_, err := NewClient("")
		if err == nil {
			t.Error("NewClient should return error for empty API key")
		}
		apiErr, ok := IsAPIError(err)
		if !ok {
			t.Error("error should be APIError")
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
		}
	})
}

func TestNewClient_WithNetwork(t *testing.T) {
	t.Run("mainnet", func(t *testing.T) {
		client, err := NewClient("test-api-key", WithNetwork(Mainnet))
		if err != nil {
			t.Fatalf("NewClient returned error: %v", err)
		}
		if client.apiURL != DefaultMainnetAPIURL {
			t.Errorf("apiURL = %s, want %s", client.apiURL, DefaultMainnetAPIURL)
		}
		if client.rpcURL != DefaultMainnetRPCURL {
			t.Errorf("rpcURL = %s, want %s", client.rpcURL, DefaultMainnetRPCURL)
		}
	})

	t.Run("devnet", func(t *testing.T) {
		client, err := NewClient("test-api-key", WithNetwork(Devnet))
		if err != nil {
			t.Fatalf("NewClient returned error: %v", err)
		}
		if client.apiURL != DefaultDevnetAPIURL {
			t.Errorf("apiURL = %s, want %s", client.apiURL, DefaultDevnetAPIURL)
		}
		if client.rpcURL != DefaultDevnetRPCURL {
			t.Errorf("rpcURL = %s, want %s", client.rpcURL, DefaultDevnetRPCURL)
		}
	})
}

func TestNewClient_WithCustomURLs(t *testing.T) {
	customAPI := "https://custom-api.example.com"
	customRPC := "https://custom-rpc.example.com"

	client, err := NewClient("test-api-key",
		WithAPIURL(customAPI),
		WithRPCURL(customRPC),
	)
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.apiURL != customAPI {
		t.Errorf("apiURL = %s, want %s", client.apiURL, customAPI)
	}
	if client.rpcURL != customRPC {
		t.Errorf("rpcURL = %s, want %s", client.rpcURL, customRPC)
	}
}

func TestNewClient_WithTimeout(t *testing.T) {
	client, err := NewClient("test-api-key", WithTimeout(30*time.Second))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	// Can't directly check timeout, but verify client is created
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestNewClient_WithMaxRetries(t *testing.T) {
	client, err := NewClient("test-api-key", WithMaxRetries(5))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client == nil {
		t.Fatal("NewClient returned nil client")
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	customClient := &http.Client{Timeout: 60 * time.Second}
	client, err := NewClient("test-api-key", WithHTTPClient(customClient))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.httpClient != customClient {
		t.Error("httpClient should be the custom client")
	}
}

// mockLogger implements Logger for testing
type mockLogger struct {
	debugCalls int
	infoCalls  int
	warnCalls  int
	errorCalls int
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) { m.debugCalls++ }
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  { m.infoCalls++ }
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  { m.warnCalls++ }
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) { m.errorCalls++ }

func TestNewClient_WithLogger(t *testing.T) {
	logger := &mockLogger{}
	client, err := NewClient("test-api-key", WithLogger(logger))
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}
	if client.logger != logger {
		t.Error("logger should be the custom logger")
	}
}

func TestClient_RPCURL(t *testing.T) {
	client, err := NewClient("my-secret-key")
	if err != nil {
		t.Fatalf("NewClient returned error: %v", err)
	}

	rpcURL := client.RPCURL()
	expectedPrefix := DefaultMainnetRPCURL + "/?api-key="
	if !strings.HasPrefix(rpcURL, expectedPrefix) {
		t.Errorf("RPCURL should start with %s, got %s", expectedPrefix, rpcURL)
	}
	if !strings.Contains(rpcURL, "my-secret-key") {
		t.Error("RPCURL should contain the API key")
	}
}

func TestClient_doRequest(t *testing.T) {
	t.Run("successful request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify API key is in query params
			if !strings.Contains(r.URL.RawQuery, "api-key=test-key") {
				t.Errorf("request should contain api-key, got query: %s", r.URL.RawQuery)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		body, err := client.doRequest(context.Background(), "GET", "/test", nil)
		if err != nil {
			t.Fatalf("doRequest returned error: %v", err)
		}
		if !strings.Contains(string(body), "success") {
			t.Errorf("body should contain success, got: %s", string(body))
		}
	})

	t.Run("error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid request"))
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.doRequest(context.Background(), "GET", "/test", nil)
		if err == nil {
			t.Fatal("doRequest should return error for 4xx response")
		}

		apiErr, ok := IsAPIError(err)
		if !ok {
			t.Fatal("error should be APIError")
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
		}
	})

	t.Run("post with body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
			}
			body, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(body), "test-data") {
				t.Errorf("body should contain test-data, got: %s", string(body))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		body, err := client.doRequest(context.Background(), "POST", "/test", strings.NewReader(`{"data":"test-data"}`))
		if err != nil {
			t.Fatalf("doRequest returned error: %v", err)
		}
		if !strings.Contains(string(body), "ok") {
			t.Errorf("body should contain ok, got: %s", string(body))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()

		client, _ := NewClient("test-key", WithAPIURL(server.URL), WithMaxRetries(0))
		_, err := client.doRequest(ctx, "GET", "/test", nil)
		if err == nil {
			t.Fatal("doRequest should return error for cancelled context")
		}
	})
}

func TestClient_doGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":"test"}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-key", WithAPIURL(server.URL))
	body, err := client.doGet(context.Background(), "/test")
	if err != nil {
		t.Fatalf("doGet returned error: %v", err)
	}
	if !strings.Contains(string(body), "test") {
		t.Errorf("body should contain test, got: %s", string(body))
	}
}

func TestClient_doPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var reqBody map[string]interface{}
		json.NewDecoder(r.Body).Decode(&reqBody)
		if reqBody["key"] != "value" {
			t.Errorf("body should contain key=value, got: %v", reqBody)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"result":"ok"}`))
	}))
	defer server.Close()

	client, _ := NewClient("test-key", WithAPIURL(server.URL))
	body, err := client.doPost(context.Background(), "/test", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("doPost returned error: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body should contain ok, got: %s", string(body))
	}
}

func TestNetworkConstants(t *testing.T) {
	if Mainnet != "mainnet" {
		t.Errorf("Mainnet = %s, want mainnet", Mainnet)
	}
	if Devnet != "devnet" {
		t.Errorf("Devnet = %s, want devnet", Devnet)
	}
}

func TestDefaultConstants(t *testing.T) {
	if DefaultMainnetAPIURL != "https://api.helius.xyz/v0" {
		t.Errorf("DefaultMainnetAPIURL = %s, unexpected value", DefaultMainnetAPIURL)
	}
	if DefaultDevnetAPIURL != "https://api-devnet.helius.xyz/v0" {
		t.Errorf("DefaultDevnetAPIURL = %s, unexpected value", DefaultDevnetAPIURL)
	}
	if DefaultMainnetRPCURL != "https://mainnet.helius-rpc.com" {
		t.Errorf("DefaultMainnetRPCURL = %s, unexpected value", DefaultMainnetRPCURL)
	}
	if DefaultDevnetRPCURL != "https://devnet.helius-rpc.com" {
		t.Errorf("DefaultDevnetRPCURL = %s, unexpected value", DefaultDevnetRPCURL)
	}
	if DefaultTimeout != 10*time.Second {
		t.Errorf("DefaultTimeout = %v, want 10s", DefaultTimeout)
	}
	if DefaultMaxRetries != 3 {
		t.Errorf("DefaultMaxRetries = %d, want 3", DefaultMaxRetries)
	}
}
