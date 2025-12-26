package helius

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// WebhookType represents the type of webhook.
type WebhookType string

const (
	// WebhookTypeEnhanced provides parsed transaction data with human-readable info.
	WebhookTypeEnhanced WebhookType = "enhanced"
	// WebhookTypeRaw provides raw transaction data.
	WebhookTypeRaw WebhookType = "raw"
	// WebhookTypeDiscord sends formatted messages to Discord.
	WebhookTypeDiscord WebhookType = "discord"
)

// TransactionType represents the type of transactions to monitor.
type TransactionType string

const (
	// TransactionTypeAny matches all transaction types.
	TransactionTypeAny TransactionType = "ANY"
	// TransactionTypeSwap matches DEX swap transactions.
	TransactionTypeSwap TransactionType = "SWAP"
	// TransactionTypeTransfer matches token transfer transactions.
	TransactionTypeTransfer TransactionType = "TRANSFER"
	// TransactionTypeNFTSale matches NFT sale transactions.
	TransactionTypeNFTSale TransactionType = "NFT_SALE"
	// TransactionTypeNFTListing matches NFT listing transactions.
	TransactionTypeNFTListing TransactionType = "NFT_LISTING"
	// TransactionTypeNFTMint matches NFT mint transactions.
	TransactionTypeNFTMint TransactionType = "NFT_MINT"
	// TransactionTypeNFTBid matches NFT bid transactions.
	TransactionTypeNFTBid TransactionType = "NFT_BID"
	// TransactionTypeNFTCancelListing matches NFT cancel listing transactions.
	TransactionTypeNFTCancelListing TransactionType = "NFT_CANCEL_LISTING"
)

// Webhook represents a Helius webhook configuration.
type Webhook struct {
	// WebhookID is the unique identifier for the webhook.
	WebhookID string `json:"webhookID"`

	// Wallet is the wallet address that created the webhook.
	Wallet string `json:"wallet"`

	// WebhookURL is the URL that receives webhook events.
	WebhookURL string `json:"webhookURL"`

	// TransactionTypes lists the transaction types being monitored.
	TransactionTypes []TransactionType `json:"transactionTypes"`

	// AccountAddresses lists the addresses being monitored.
	AccountAddresses []string `json:"accountAddresses"`

	// WebhookType is the format of webhook data.
	WebhookType WebhookType `json:"webhookType"`

	// AuthHeader is an optional authorization header sent with webhooks.
	AuthHeader string `json:"authHeader,omitempty"`
}

// CreateWebhookRequest configures a new webhook.
type CreateWebhookRequest struct {
	// WebhookURL is the URL that will receive webhook events (required).
	WebhookURL string `json:"webhookURL"`

	// TransactionTypes lists which transaction types to monitor (required).
	TransactionTypes []TransactionType `json:"transactionTypes"`

	// AccountAddresses lists the addresses to monitor (required, max 10,000).
	AccountAddresses []string `json:"accountAddresses"`

	// WebhookType is the format of webhook data (default: enhanced).
	WebhookType WebhookType `json:"webhookType,omitempty"`

	// AuthHeader is an optional authorization header to include in webhooks.
	AuthHeader string `json:"authHeader,omitempty"`
}

// CreateWebhook creates a new webhook for monitoring transactions.
func (c *Client) CreateWebhook(ctx context.Context, req *CreateWebhookRequest) (*Webhook, error) {
	if req == nil {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "request is required",
			Path:       "/webhooks",
		}
	}
	if req.WebhookURL == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "webhookURL is required",
			Path:       "/webhooks",
		}
	}
	if len(req.TransactionTypes) == 0 {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "at least one transactionType is required",
			Path:       "/webhooks",
		}
	}
	if len(req.AccountAddresses) == 0 {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "at least one accountAddress is required",
			Path:       "/webhooks",
		}
	}

	// Default to enhanced webhooks
	if req.WebhookType == "" {
		req.WebhookType = WebhookTypeEnhanced
	}

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, err := c.doRequest(ctx, "POST", "/webhooks", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var webhook Webhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Info("created webhook",
		"webhookID", webhook.WebhookID,
		"url", webhook.WebhookURL,
		"addresses", len(webhook.AccountAddresses),
	)

	return &webhook, nil
}

// GetWebhook fetches a webhook by its ID.
func (c *Client) GetWebhook(ctx context.Context, webhookID string) (*Webhook, error) {
	if webhookID == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "webhookID is required",
			Path:       "/webhooks",
		}
	}

	path := fmt.Sprintf("/webhooks/%s", webhookID)
	body, err := c.doGet(ctx, path)
	if err != nil {
		return nil, err
	}

	var webhook Webhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &webhook, nil
}

// ListWebhooks lists all webhooks for the account.
func (c *Client) ListWebhooks(ctx context.Context) ([]Webhook, error) {
	body, err := c.doGet(ctx, "/webhooks")
	if err != nil {
		return nil, err
	}

	var webhooks []Webhook
	if err := json.Unmarshal(body, &webhooks); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("listed webhooks", "count", len(webhooks))

	return webhooks, nil
}

// UpdateWebhookRequest configures webhook updates.
type UpdateWebhookRequest struct {
	// WebhookURL updates the webhook endpoint.
	WebhookURL string `json:"webhookURL,omitempty"`

	// TransactionTypes updates which transaction types to monitor.
	TransactionTypes []TransactionType `json:"transactionTypes,omitempty"`

	// AccountAddresses updates the addresses to monitor.
	AccountAddresses []string `json:"accountAddresses,omitempty"`

	// WebhookType updates the format of webhook data.
	WebhookType WebhookType `json:"webhookType,omitempty"`

	// AuthHeader updates the authorization header.
	AuthHeader string `json:"authHeader,omitempty"`
}

// UpdateWebhook updates an existing webhook.
func (c *Client) UpdateWebhook(ctx context.Context, webhookID string, req *UpdateWebhookRequest) (*Webhook, error) {
	if webhookID == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "webhookID is required",
			Path:       "/webhooks",
		}
	}
	if req == nil {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "request is required",
			Path:       "/webhooks",
		}
	}

	path := fmt.Sprintf("/webhooks/%s", webhookID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, err := c.doRequest(ctx, "PUT", path, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var webhook Webhook
	if err := json.Unmarshal(body, &webhook); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Info("updated webhook", "webhookID", webhookID)

	return &webhook, nil
}

// DeleteWebhook deletes a webhook.
func (c *Client) DeleteWebhook(ctx context.Context, webhookID string) error {
	if webhookID == "" {
		return &APIError{
			StatusCode: 400,
			Message:    "webhookID is required",
			Path:       "/webhooks",
		}
	}

	path := fmt.Sprintf("/webhooks/%s", webhookID)

	_, err := c.doRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return err
	}

	c.logger.Info("deleted webhook", "webhookID", webhookID)

	return nil
}

// ValidateWebhookSignature validates the HMAC signature of a webhook payload.
//
// This should be called for every incoming webhook to verify authenticity.
// The signature is typically passed in the X-Helius-Signature header.
//
// Example:
//
//	func handleWebhook(w http.ResponseWriter, r *http.Request) {
//	    body, _ := io.ReadAll(r.Body)
//	    signature := r.Header.Get("X-Helius-Signature")
//
//	    if !helius.ValidateWebhookSignature(body, signature, webhookSecret) {
//	        http.Error(w, "invalid signature", http.StatusUnauthorized)
//	        return
//	    }
//	    // Process webhook...
//	}
func ValidateWebhookSignature(body []byte, signature string, secret string) bool {
	if signature == "" || secret == "" {
		return false
	}

	h := hmac.New(sha256.New, []byte(secret))
	h.Write(body)
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// WebhookEvent represents an incoming webhook event.
type WebhookEvent struct {
	// AccountData contains the account data changes.
	AccountData []AccountData `json:"accountData,omitempty"`

	// Description is a human-readable description (enhanced webhooks only).
	Description string `json:"description,omitempty"`

	// Events contains parsed event information.
	Events interface{} `json:"events,omitempty"`

	// Fee is the transaction fee in lamports.
	Fee int64 `json:"fee,omitempty"`

	// FeePayer is the address that paid the transaction fee.
	FeePayer string `json:"feePayer,omitempty"`

	// Instructions contains parsed instructions.
	Instructions []interface{} `json:"instructions,omitempty"`

	// NativeTransfers contains SOL transfer information.
	NativeTransfers []NativeTransfer `json:"nativeTransfers,omitempty"`

	// Signature is the transaction signature.
	Signature string `json:"signature"`

	// Slot is the slot the transaction was processed in.
	Slot int64 `json:"slot"`

	// Source is the source of the transaction (e.g., "JUPITER").
	Source string `json:"source,omitempty"`

	// Timestamp is the Unix timestamp of the transaction.
	Timestamp int64 `json:"timestamp,omitempty"`

	// TokenTransfers contains token transfer information.
	TokenTransfers []TokenTransfer `json:"tokenTransfers,omitempty"`

	// Type is the transaction type (e.g., "SWAP").
	Type string `json:"type,omitempty"`
}

// AccountData represents account data changes.
type AccountData struct {
	Account             string `json:"account"`
	NativeBalanceChange int64  `json:"nativeBalanceChange,omitempty"`
	TokenBalanceChanges []TokenBalanceChange `json:"tokenBalanceChanges,omitempty"`
}

// TokenBalanceChange represents a token balance change.
type TokenBalanceChange struct {
	Mint            string `json:"mint"`
	RawTokenAmount  RawTokenAmount `json:"rawTokenAmount"`
	TokenAccount    string `json:"tokenAccount"`
	UserAccount     string `json:"userAccount"`
}

// RawTokenAmount represents a raw token amount.
type RawTokenAmount struct {
	Decimals    int    `json:"decimals"`
	TokenAmount string `json:"tokenAmount"`
}

// NativeTransfer represents a SOL transfer.
type NativeTransfer struct {
	Amount      int64  `json:"amount"`
	FromUserAccount string `json:"fromUserAccount"`
	ToUserAccount   string `json:"toUserAccount"`
}

// TokenTransfer represents a token transfer.
type TokenTransfer struct {
	FromTokenAccount string  `json:"fromTokenAccount"`
	FromUserAccount  string  `json:"fromUserAccount"`
	Mint             string  `json:"mint"`
	ToTokenAccount   string  `json:"toTokenAccount"`
	ToUserAccount    string  `json:"toUserAccount"`
	TokenAmount      float64 `json:"tokenAmount"`
	TokenStandard    string  `json:"tokenStandard,omitempty"`
}

// ParseWebhookEvent parses a webhook payload into a WebhookEvent.
func ParseWebhookEvent(body []byte) (*WebhookEvent, error) {
	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("parse webhook event: %w", err)
	}
	return &event, nil
}

// ParseWebhookEvents parses a webhook payload containing multiple events.
func ParseWebhookEvents(body []byte) ([]WebhookEvent, error) {
	var events []WebhookEvent
	if err := json.Unmarshal(body, &events); err != nil {
		// Try single event
		var event WebhookEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return nil, fmt.Errorf("parse webhook events: %w", err)
		}
		return []WebhookEvent{event}, nil
	}
	return events, nil
}
