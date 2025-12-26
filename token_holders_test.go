package helius

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTokenHolders(t *testing.T) {
	t.Run("successful get", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/token-holders" {
				t.Errorf("expected /token-holders, got %s", r.URL.Path)
			}

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			if req["mint"] != "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v" {
				t.Errorf("unexpected mint: %v", req["mint"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TokenHoldersPage{
				Total: 50000,
				Limit: 1000,
				TokenHolders: []TokenHolder{
					{Owner: "holder-1", Balance: 1000000000, Decimals: 6},
					{Owner: "holder-2", Balance: 500000000, Decimals: 6},
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetTokenHolders(context.Background(),
			"EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			nil,
		)

		if err != nil {
			t.Fatalf("GetTokenHolders returned error: %v", err)
		}
		if page.Total != 50000 {
			t.Errorf("Total = %d, want 50000", page.Total)
		}
		if len(page.TokenHolders) != 2 {
			t.Errorf("len(TokenHolders) = %d, want 2", len(page.TokenHolders))
		}
	})

	t.Run("empty mint", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetTokenHolders(context.Background(), "", nil)
		if err == nil {
			t.Error("GetTokenHolders should return error for empty mint")
		}
	})

	t.Run("with pagination options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["cursor"] != "next-cursor" {
				t.Errorf("cursor = %v, want next-cursor", req["cursor"])
			}
			if req["limit"] != float64(500) {
				t.Errorf("limit = %v, want 500", req["limit"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TokenHoldersPage{
				Total:        50000,
				Limit:        500,
				Cursor:       "another-cursor",
				TokenHolders: []TokenHolder{},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		page, err := client.GetTokenHolders(context.Background(), "some-mint", &GetTokenHoldersOptions{
			Cursor: "next-cursor",
			Limit:  500,
		})

		if err != nil {
			t.Fatalf("GetTokenHolders returned error: %v", err)
		}
		if page.Cursor != "another-cursor" {
			t.Errorf("Cursor = %s, want another-cursor", page.Cursor)
		}
	})
}

func TestGetAllTokenHolders(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TokenHoldersPage{
				Total:  3,
				Limit:  10000,
				Cursor: "", // Empty cursor means no more pages
				TokenHolders: []TokenHolder{
					{Owner: "holder-1", Balance: 100},
					{Owner: "holder-2", Balance: 50},
					{Owner: "holder-3", Balance: 25},
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		holders, err := client.GetAllTokenHolders(context.Background(), "some-mint")

		if err != nil {
			t.Fatalf("GetAllTokenHolders returned error: %v", err)
		}
		if len(holders) != 3 {
			t.Errorf("len(holders) = %d, want 3", len(holders))
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			w.WriteHeader(http.StatusOK)
			if callCount == 1 {
				json.NewEncoder(w).Encode(TokenHoldersPage{
					Total:  5,
					Limit:  2,
					Cursor: "page-2",
					TokenHolders: []TokenHolder{
						{Owner: "holder-1", Balance: 100},
						{Owner: "holder-2", Balance: 50},
					},
				})
			} else if callCount == 2 {
				if req["cursor"] != "page-2" {
					t.Errorf("expected cursor page-2, got %v", req["cursor"])
				}
				json.NewEncoder(w).Encode(TokenHoldersPage{
					Total:  5,
					Limit:  2,
					Cursor: "page-3",
					TokenHolders: []TokenHolder{
						{Owner: "holder-3", Balance: 25},
						{Owner: "holder-4", Balance: 10},
					},
				})
			} else {
				if req["cursor"] != "page-3" {
					t.Errorf("expected cursor page-3, got %v", req["cursor"])
				}
				json.NewEncoder(w).Encode(TokenHoldersPage{
					Total:  5,
					Limit:  2,
					Cursor: "", // No more pages
					TokenHolders: []TokenHolder{
						{Owner: "holder-5", Balance: 5},
					},
				})
			}
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		holders, err := client.GetAllTokenHolders(context.Background(), "some-mint")

		if err != nil {
			t.Fatalf("GetAllTokenHolders returned error: %v", err)
		}
		if len(holders) != 5 {
			t.Errorf("len(holders) = %d, want 5", len(holders))
		}
		if callCount != 3 {
			t.Errorf("callCount = %d, want 3", callCount)
		}
	})

	t.Run("empty result", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(TokenHoldersPage{
				Total:        0,
				Limit:        10000,
				TokenHolders: []TokenHolder{},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		holders, err := client.GetAllTokenHolders(context.Background(), "some-mint")

		if err != nil {
			t.Fatalf("GetAllTokenHolders returned error: %v", err)
		}
		if len(holders) != 0 {
			t.Errorf("len(holders) = %d, want 0", len(holders))
		}
	})
}

func TestCalculateTopHolderStats(t *testing.T) {
	t.Run("normal case", func(t *testing.T) {
		holders := []TokenHolder{
			{Owner: "whale-1", Balance: 1000},
			{Owner: "whale-2", Balance: 500},
			{Owner: "user-3", Balance: 200},
			{Owner: "user-4", Balance: 150},
			{Owner: "user-5", Balance: 100},
			{Owner: "user-6", Balance: 50},
		}

		stats := CalculateTopHolderStats(holders, 3)

		if stats.TotalHolders != 6 {
			t.Errorf("TotalHolders = %d, want 6", stats.TotalHolders)
		}
		if stats.TotalSupply != 2000 {
			t.Errorf("TotalSupply = %d, want 2000", stats.TotalSupply)
		}
		if len(stats.TopHolders) != 3 {
			t.Errorf("len(TopHolders) = %d, want 3", len(stats.TopHolders))
		}
		if stats.TopHoldersBalance != 1700 { // 1000 + 500 + 200
			t.Errorf("TopHoldersBalance = %d, want 1700", stats.TopHoldersBalance)
		}
		expectedPercent := 85.0 // (1700 / 2000) * 100
		if stats.TopHoldersPercent != expectedPercent {
			t.Errorf("TopHoldersPercent = %f, want %f", stats.TopHoldersPercent, expectedPercent)
		}
	})

	t.Run("top n greater than holders", func(t *testing.T) {
		holders := []TokenHolder{
			{Owner: "holder-1", Balance: 100},
			{Owner: "holder-2", Balance: 50},
		}

		stats := CalculateTopHolderStats(holders, 10)

		if stats.TotalHolders != 2 {
			t.Errorf("TotalHolders = %d, want 2", stats.TotalHolders)
		}
		if len(stats.TopHolders) != 2 {
			t.Errorf("len(TopHolders) = %d, want 2", len(stats.TopHolders))
		}
		if stats.TopHoldersBalance != 150 {
			t.Errorf("TopHoldersBalance = %d, want 150", stats.TopHoldersBalance)
		}
		if stats.TopHoldersPercent != 100.0 {
			t.Errorf("TopHoldersPercent = %f, want 100", stats.TopHoldersPercent)
		}
	})

	t.Run("empty holders", func(t *testing.T) {
		holders := []TokenHolder{}

		stats := CalculateTopHolderStats(holders, 10)

		if stats.TotalHolders != 0 {
			t.Errorf("TotalHolders = %d, want 0", stats.TotalHolders)
		}
		if stats.TotalSupply != 0 {
			t.Errorf("TotalSupply = %d, want 0", stats.TotalSupply)
		}
		if stats.TopHoldersPercent != 0 {
			t.Errorf("TopHoldersPercent = %f, want 0", stats.TopHoldersPercent)
		}
	})

	t.Run("single holder", func(t *testing.T) {
		holders := []TokenHolder{
			{Owner: "only-holder", Balance: 1000000},
		}

		stats := CalculateTopHolderStats(holders, 5)

		if stats.TotalHolders != 1 {
			t.Errorf("TotalHolders = %d, want 1", stats.TotalHolders)
		}
		if len(stats.TopHolders) != 1 {
			t.Errorf("len(TopHolders) = %d, want 1", len(stats.TopHolders))
		}
		if stats.TopHoldersPercent != 100.0 {
			t.Errorf("TopHoldersPercent = %f, want 100", stats.TopHoldersPercent)
		}
	})

	t.Run("top 1 holder only", func(t *testing.T) {
		holders := []TokenHolder{
			{Owner: "whale", Balance: 900},
			{Owner: "fish-1", Balance: 50},
			{Owner: "fish-2", Balance: 50},
		}

		stats := CalculateTopHolderStats(holders, 1)

		if len(stats.TopHolders) != 1 {
			t.Errorf("len(TopHolders) = %d, want 1", len(stats.TopHolders))
		}
		if stats.TopHoldersBalance != 900 {
			t.Errorf("TopHoldersBalance = %d, want 900", stats.TopHoldersBalance)
		}
		if stats.TopHoldersPercent != 90.0 {
			t.Errorf("TopHoldersPercent = %f, want 90", stats.TopHoldersPercent)
		}
	})

	t.Run("zero supply edge case", func(t *testing.T) {
		// All holders have 0 balance
		holders := []TokenHolder{
			{Owner: "holder-1", Balance: 0},
			{Owner: "holder-2", Balance: 0},
		}

		stats := CalculateTopHolderStats(holders, 5)

		if stats.TotalSupply != 0 {
			t.Errorf("TotalSupply = %d, want 0", stats.TotalSupply)
		}
		if stats.TopHoldersPercent != 0 {
			t.Errorf("TopHoldersPercent = %f, want 0", stats.TopHoldersPercent)
		}
	})
}

func TestTokenHolderTypes(t *testing.T) {
	t.Run("token holder", func(t *testing.T) {
		holder := TokenHolder{
			Owner:        "owner-wallet",
			TokenAccount: "token-account",
			Balance:      1000000000,
			Decimals:     6,
		}
		if holder.Owner != "owner-wallet" {
			t.Errorf("Owner = %s, unexpected value", holder.Owner)
		}
		if holder.Balance != 1000000000 {
			t.Errorf("Balance = %d, unexpected value", holder.Balance)
		}
	})

	t.Run("token holders page", func(t *testing.T) {
		page := TokenHoldersPage{
			Total:  100,
			Limit:  10,
			Cursor: "next-page",
			TokenHolders: []TokenHolder{
				{Owner: "holder-1"},
			},
		}
		if page.Total != 100 {
			t.Errorf("Total = %d, want 100", page.Total)
		}
		if page.Cursor != "next-page" {
			t.Errorf("Cursor = %s, want next-page", page.Cursor)
		}
	})

	t.Run("top holder stats", func(t *testing.T) {
		stats := TopHolderStats{
			TotalHolders:      1000,
			TopHolders:        []TokenHolder{{Owner: "whale"}},
			TopHoldersBalance: 5000000,
			TopHoldersPercent: 45.5,
			TotalSupply:       11000000,
		}
		if stats.TotalHolders != 1000 {
			t.Errorf("TotalHolders = %d, want 1000", stats.TotalHolders)
		}
		if stats.TopHoldersPercent != 45.5 {
			t.Errorf("TopHoldersPercent = %f, want 45.5", stats.TopHoldersPercent)
		}
	})

	t.Run("get token holders options", func(t *testing.T) {
		opts := GetTokenHoldersOptions{
			Cursor: "cursor-value",
			Limit:  5000,
		}
		if opts.Cursor != "cursor-value" {
			t.Errorf("Cursor = %s, want cursor-value", opts.Cursor)
		}
		if opts.Limit != 5000 {
			t.Errorf("Limit = %d, want 5000", opts.Limit)
		}
	})
}
