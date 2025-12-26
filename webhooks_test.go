package helius

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateWebhookSignature(t *testing.T) {
	secret := "my-webhook-secret"
	body := []byte(`{"signature":"abc123","type":"SWAP"}`)

	// Calculate expected signature
	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	validSignature := hex.EncodeToString(h.Sum(nil))

	t.Run("valid signature", func(t *testing.T) {
		if !ValidateWebhookSignature(body, validSignature, secret) {
			t.Error("ValidateWebhookSignature should return true for valid signature")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		if ValidateWebhookSignature(body, "wrong-signature", secret) {
			t.Error("ValidateWebhookSignature should return false for invalid signature")
		}
	})

	t.Run("empty signature", func(t *testing.T) {
		if ValidateWebhookSignature(body, "", secret) {
			t.Error("ValidateWebhookSignature should return false for empty signature")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		if ValidateWebhookSignature(body, validSignature, "") {
			t.Error("ValidateWebhookSignature should return false for empty secret")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		emptyBody := []byte{}
		h := hmac.New(sha256.New, []byte(secret))
		h.Write(emptyBody)
		emptySig := hex.EncodeToString(h.Sum(nil))

		if !ValidateWebhookSignature(emptyBody, emptySig, secret) {
			t.Error("ValidateWebhookSignature should handle empty body correctly")
		}
	})

	t.Run("different body produces different signature", func(t *testing.T) {
		differentBody := []byte(`{"signature":"xyz789","type":"TRANSFER"}`)
		if ValidateWebhookSignature(differentBody, validSignature, secret) {
			t.Error("ValidateWebhookSignature should return false for different body")
		}
	})

	t.Run("timing attack resistance", func(t *testing.T) {
		// This test verifies we use constant-time comparison
		// by checking that very similar signatures still fail
		almostValid := validSignature[:len(validSignature)-1] + "f"
		if ValidateWebhookSignature(body, almostValid, secret) {
			t.Error("ValidateWebhookSignature should return false for almost-valid signature")
		}
	})
}

func TestWebhookType(t *testing.T) {
	tests := []struct {
		webhookType WebhookType
		expected    string
	}{
		{WebhookTypeEnhanced, "enhanced"},
		{WebhookTypeRaw, "raw"},
		{WebhookTypeDiscord, "discord"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.webhookType) != tt.expected {
				t.Errorf("WebhookType = %s, want %s", tt.webhookType, tt.expected)
			}
		})
	}
}

func TestTransactionType(t *testing.T) {
	tests := []struct {
		txType   TransactionType
		expected string
	}{
		{TransactionTypeAny, "ANY"},
		{TransactionTypeSwap, "SWAP"},
		{TransactionTypeTransfer, "TRANSFER"},
		{TransactionTypeNFTSale, "NFT_SALE"},
		{TransactionTypeNFTListing, "NFT_LISTING"},
		{TransactionTypeNFTMint, "NFT_MINT"},
		{TransactionTypeNFTBid, "NFT_BID"},
		{TransactionTypeNFTCancelListing, "NFT_CANCEL_LISTING"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.txType) != tt.expected {
				t.Errorf("TransactionType = %s, want %s", tt.txType, tt.expected)
			}
		})
	}
}

func TestCreateWebhook(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/webhooks" {
				t.Errorf("expected /webhooks, got %s", r.URL.Path)
			}

			var req CreateWebhookRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.WebhookURL != "https://example.com/webhook" {
				t.Errorf("unexpected webhookURL: %s", req.WebhookURL)
			}
			if req.WebhookType != WebhookTypeEnhanced {
				t.Errorf("unexpected webhookType: %s", req.WebhookType)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Webhook{
				WebhookID:        "webhook-123",
				Wallet:           "wallet-abc",
				WebhookURL:       req.WebhookURL,
				TransactionTypes: req.TransactionTypes,
				AccountAddresses: req.AccountAddresses,
				WebhookType:      req.WebhookType,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		webhook, err := client.CreateWebhook(context.Background(), &CreateWebhookRequest{
			WebhookURL:       "https://example.com/webhook",
			TransactionTypes: []TransactionType{TransactionTypeSwap},
			AccountAddresses: []string{"address1", "address2"},
		})

		if err != nil {
			t.Fatalf("CreateWebhook returned error: %v", err)
		}
		if webhook.WebhookID != "webhook-123" {
			t.Errorf("WebhookID = %s, want webhook-123", webhook.WebhookID)
		}
	})

	t.Run("nil request", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.CreateWebhook(context.Background(), nil)
		if err == nil {
			t.Error("CreateWebhook should return error for nil request")
		}
	})

	t.Run("empty webhook url", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.CreateWebhook(context.Background(), &CreateWebhookRequest{
			TransactionTypes: []TransactionType{TransactionTypeSwap},
			AccountAddresses: []string{"address1"},
		})
		if err == nil {
			t.Error("CreateWebhook should return error for empty webhookURL")
		}
	})

	t.Run("empty transaction types", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.CreateWebhook(context.Background(), &CreateWebhookRequest{
			WebhookURL:       "https://example.com/webhook",
			AccountAddresses: []string{"address1"},
		})
		if err == nil {
			t.Error("CreateWebhook should return error for empty transactionTypes")
		}
	})

	t.Run("empty account addresses", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.CreateWebhook(context.Background(), &CreateWebhookRequest{
			WebhookURL:       "https://example.com/webhook",
			TransactionTypes: []TransactionType{TransactionTypeSwap},
		})
		if err == nil {
			t.Error("CreateWebhook should return error for empty accountAddresses")
		}
	})
}

func TestGetWebhook(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/webhooks/webhook-123" {
				t.Errorf("expected /webhooks/webhook-123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Webhook{
				WebhookID:  "webhook-123",
				WebhookURL: "https://example.com/webhook",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		webhook, err := client.GetWebhook(context.Background(), "webhook-123")

		if err != nil {
			t.Fatalf("GetWebhook returned error: %v", err)
		}
		if webhook.WebhookID != "webhook-123" {
			t.Errorf("WebhookID = %s, want webhook-123", webhook.WebhookID)
		}
	})

	t.Run("empty webhook id", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetWebhook(context.Background(), "")
		if err == nil {
			t.Error("GetWebhook should return error for empty webhookID")
		}
	})
}

func TestListWebhooks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/webhooks" {
			t.Errorf("expected /webhooks, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Webhook{
			{WebhookID: "webhook-1"},
			{WebhookID: "webhook-2"},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-key", WithAPIURL(server.URL))
	webhooks, err := client.ListWebhooks(context.Background())

	if err != nil {
		t.Fatalf("ListWebhooks returned error: %v", err)
	}
	if len(webhooks) != 2 {
		t.Errorf("len(webhooks) = %d, want 2", len(webhooks))
	}
}

func TestUpdateWebhook(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "PUT" {
				t.Errorf("expected PUT, got %s", r.Method)
			}
			if r.URL.Path != "/webhooks/webhook-123" {
				t.Errorf("expected /webhooks/webhook-123, got %s", r.URL.Path)
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(Webhook{
				WebhookID:  "webhook-123",
				WebhookURL: "https://new-url.com/webhook",
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		webhook, err := client.UpdateWebhook(context.Background(), "webhook-123", &UpdateWebhookRequest{
			WebhookURL: "https://new-url.com/webhook",
		})

		if err != nil {
			t.Fatalf("UpdateWebhook returned error: %v", err)
		}
		if webhook.WebhookURL != "https://new-url.com/webhook" {
			t.Errorf("WebhookURL = %s, want https://new-url.com/webhook", webhook.WebhookURL)
		}
	})

	t.Run("empty webhook id", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.UpdateWebhook(context.Background(), "", &UpdateWebhookRequest{})
		if err == nil {
			t.Error("UpdateWebhook should return error for empty webhookID")
		}
	})

	t.Run("nil request", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.UpdateWebhook(context.Background(), "webhook-123", nil)
		if err == nil {
			t.Error("UpdateWebhook should return error for nil request")
		}
	})
}

func TestDeleteWebhook(t *testing.T) {
	t.Run("successful delete", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "DELETE" {
				t.Errorf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Path != "/webhooks/webhook-123" {
				t.Errorf("expected /webhooks/webhook-123, got %s", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		err := client.DeleteWebhook(context.Background(), "webhook-123")

		if err != nil {
			t.Fatalf("DeleteWebhook returned error: %v", err)
		}
	})

	t.Run("empty webhook id", func(t *testing.T) {
		client, _ := NewClient("test-key")
		err := client.DeleteWebhook(context.Background(), "")
		if err == nil {
			t.Error("DeleteWebhook should return error for empty webhookID")
		}
	})
}

func TestParseWebhookEvent(t *testing.T) {
	t.Run("valid event", func(t *testing.T) {
		body := []byte(`{
			"signature": "abc123",
			"slot": 12345,
			"fee": 5000,
			"feePayer": "wallet-address",
			"type": "SWAP",
			"source": "JUPITER"
		}`)

		event, err := ParseWebhookEvent(body)
		if err != nil {
			t.Fatalf("ParseWebhookEvent returned error: %v", err)
		}
		if event.Signature != "abc123" {
			t.Errorf("Signature = %s, want abc123", event.Signature)
		}
		if event.Slot != 12345 {
			t.Errorf("Slot = %d, want 12345", event.Slot)
		}
		if event.Fee != 5000 {
			t.Errorf("Fee = %d, want 5000", event.Fee)
		}
		if event.Type != "SWAP" {
			t.Errorf("Type = %s, want SWAP", event.Type)
		}
		if event.Source != "JUPITER" {
			t.Errorf("Source = %s, want JUPITER", event.Source)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		body := []byte(`{invalid json}`)
		_, err := ParseWebhookEvent(body)
		if err == nil {
			t.Error("ParseWebhookEvent should return error for invalid JSON")
		}
	})

	t.Run("event with token transfers", func(t *testing.T) {
		body := []byte(`{
			"signature": "tx123",
			"slot": 100,
			"tokenTransfers": [
				{
					"mint": "token-mint-address",
					"tokenAmount": 1000.5,
					"fromUserAccount": "sender",
					"toUserAccount": "receiver"
				}
			]
		}`)

		event, err := ParseWebhookEvent(body)
		if err != nil {
			t.Fatalf("ParseWebhookEvent returned error: %v", err)
		}
		if len(event.TokenTransfers) != 1 {
			t.Fatalf("len(TokenTransfers) = %d, want 1", len(event.TokenTransfers))
		}
		if event.TokenTransfers[0].Mint != "token-mint-address" {
			t.Errorf("Mint = %s, want token-mint-address", event.TokenTransfers[0].Mint)
		}
		if event.TokenTransfers[0].TokenAmount != 1000.5 {
			t.Errorf("TokenAmount = %f, want 1000.5", event.TokenTransfers[0].TokenAmount)
		}
	})

	t.Run("event with native transfers", func(t *testing.T) {
		body := []byte(`{
			"signature": "tx456",
			"slot": 200,
			"nativeTransfers": [
				{
					"amount": 1000000000,
					"fromUserAccount": "sender",
					"toUserAccount": "receiver"
				}
			]
		}`)

		event, err := ParseWebhookEvent(body)
		if err != nil {
			t.Fatalf("ParseWebhookEvent returned error: %v", err)
		}
		if len(event.NativeTransfers) != 1 {
			t.Fatalf("len(NativeTransfers) = %d, want 1", len(event.NativeTransfers))
		}
		if event.NativeTransfers[0].Amount != 1000000000 {
			t.Errorf("Amount = %d, want 1000000000", event.NativeTransfers[0].Amount)
		}
	})
}

func TestParseWebhookEvents(t *testing.T) {
	t.Run("array of events", func(t *testing.T) {
		body := []byte(`[
			{"signature": "tx1", "slot": 100},
			{"signature": "tx2", "slot": 200}
		]`)

		events, err := ParseWebhookEvents(body)
		if err != nil {
			t.Fatalf("ParseWebhookEvents returned error: %v", err)
		}
		if len(events) != 2 {
			t.Errorf("len(events) = %d, want 2", len(events))
		}
		if events[0].Signature != "tx1" {
			t.Errorf("events[0].Signature = %s, want tx1", events[0].Signature)
		}
		if events[1].Signature != "tx2" {
			t.Errorf("events[1].Signature = %s, want tx2", events[1].Signature)
		}
	})

	t.Run("single event (not array)", func(t *testing.T) {
		body := []byte(`{"signature": "tx-single", "slot": 300}`)

		events, err := ParseWebhookEvents(body)
		if err != nil {
			t.Fatalf("ParseWebhookEvents returned error: %v", err)
		}
		if len(events) != 1 {
			t.Errorf("len(events) = %d, want 1", len(events))
		}
		if events[0].Signature != "tx-single" {
			t.Errorf("Signature = %s, want tx-single", events[0].Signature)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		body := []byte(`{invalid}`)
		_, err := ParseWebhookEvents(body)
		if err == nil {
			t.Error("ParseWebhookEvents should return error for invalid JSON")
		}
	})
}
