package helius

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPriorityLevel(t *testing.T) {
	tests := []struct {
		level    PriorityLevel
		expected string
	}{
		{PriorityMin, "Min"},
		{PriorityLow, "Low"},
		{PriorityMedium, "Medium"},
		{PriorityHigh, "High"},
		{PriorityVeryHigh, "VeryHigh"},
		{PriorityUnsafeMax, "UnsafeMax"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.level) != tt.expected {
				t.Errorf("PriorityLevel = %s, want %s", tt.level, tt.expected)
			}
		})
	}
}

func TestGetPriorityFeeEstimate(t *testing.T) {
	t.Run("successful estimate", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/priority-fee" {
				t.Errorf("expected /priority-fee, got %s", r.URL.Path)
			}

			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)
			accounts := req["accountKeys"].([]interface{})
			if len(accounts) != 2 {
				t.Errorf("len(accountKeys) = %d, want 2", len(accounts))
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{
				PriorityFeeEstimate: 50000.0,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		estimate, err := client.GetPriorityFeeEstimate(context.Background(),
			[]string{"JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB", "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"},
			nil,
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimate returned error: %v", err)
		}
		if estimate.PriorityFeeEstimate != 50000.0 {
			t.Errorf("PriorityFeeEstimate = %f, want 50000", estimate.PriorityFeeEstimate)
		}
	})

	t.Run("empty account keys", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetPriorityFeeEstimate(context.Background(), []string{}, nil)
		if err == nil {
			t.Error("GetPriorityFeeEstimate should return error for empty accountKeys")
		}
	})

	t.Run("with priority level option", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["priorityLevel"] != "High" {
				t.Errorf("priorityLevel = %v, want High", opts["priorityLevel"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{
				PriorityFeeEstimate: 100000.0,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		estimate, err := client.GetPriorityFeeEstimate(context.Background(),
			[]string{"some-account"},
			&GetPriorityFeeOptions{
				PriorityLevel: PriorityHigh,
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimate returned error: %v", err)
		}
		if estimate.PriorityFeeEstimate != 100000.0 {
			t.Errorf("PriorityFeeEstimate = %f, want 100000", estimate.PriorityFeeEstimate)
		}
	})

	t.Run("with all priority levels", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["includeAllPriorityFeeLevels"] != true {
				t.Errorf("includeAllPriorityFeeLevels = %v, want true", opts["includeAllPriorityFeeLevels"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{
				PriorityFeeEstimate: 50000.0,
				PriorityFeeLevels: &PriorityFeeLevels{
					Min:       1000.0,
					Low:       10000.0,
					Medium:    50000.0,
					High:      100000.0,
					VeryHigh:  200000.0,
					UnsafeMax: 500000.0,
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		estimate, err := client.GetPriorityFeeEstimate(context.Background(),
			[]string{"some-account"},
			&GetPriorityFeeOptions{
				IncludeAllPriorityFeeLevels: true,
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimate returned error: %v", err)
		}
		if estimate.PriorityFeeLevels == nil {
			t.Fatal("PriorityFeeLevels should not be nil")
		}
		if estimate.PriorityFeeLevels.High != 100000.0 {
			t.Errorf("High = %f, want 100000", estimate.PriorityFeeLevels.High)
		}
	})

	t.Run("with lookback slots", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["lookbackSlots"] != float64(200) {
				t.Errorf("lookbackSlots = %v, want 200", opts["lookbackSlots"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{PriorityFeeEstimate: 50000.0})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.GetPriorityFeeEstimate(context.Background(),
			[]string{"some-account"},
			&GetPriorityFeeOptions{
				LookbackSlots: 200,
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimate returned error: %v", err)
		}
	})

	t.Run("with recommended", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["recommended"] != true {
				t.Errorf("recommended = %v, want true", opts["recommended"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{PriorityFeeEstimate: 75000.0})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.GetPriorityFeeEstimate(context.Background(),
			[]string{"some-account"},
			&GetPriorityFeeOptions{
				Recommended: true,
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimate returned error: %v", err)
		}
	})
}

func TestGetPriorityFeeEstimateForTransaction(t *testing.T) {
	t.Run("successful estimate", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			if req["transaction"] != "base64-encoded-transaction" {
				t.Errorf("transaction = %v, unexpected value", req["transaction"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{
				PriorityFeeEstimate: 60000.0,
			})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		estimate, err := client.GetPriorityFeeEstimateForTransaction(context.Background(),
			"base64-encoded-transaction",
			nil,
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimateForTransaction returned error: %v", err)
		}
		if estimate.PriorityFeeEstimate != 60000.0 {
			t.Errorf("PriorityFeeEstimate = %f, want 60000", estimate.PriorityFeeEstimate)
		}
	})

	t.Run("empty transaction", func(t *testing.T) {
		client, _ := NewClient("test-key")
		_, err := client.GetPriorityFeeEstimateForTransaction(context.Background(), "", nil)
		if err == nil {
			t.Error("GetPriorityFeeEstimateForTransaction should return error for empty transaction")
		}
	})

	t.Run("with encoding option", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["transactionEncoding"] != "base64" {
				t.Errorf("transactionEncoding = %v, want base64", opts["transactionEncoding"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{PriorityFeeEstimate: 50000.0})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.GetPriorityFeeEstimateForTransaction(context.Background(),
			"tx-data",
			&GetPriorityFeeOptions{
				TransactionEncoding: "base64",
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimateForTransaction returned error: %v", err)
		}
	})

	t.Run("with evaluate empty slot as zero", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req map[string]interface{}
			json.NewDecoder(r.Body).Decode(&req)

			opts := req["options"].(map[string]interface{})
			if opts["evaluateEmptySlotAsZero"] != true {
				t.Errorf("evaluateEmptySlotAsZero = %v, want true", opts["evaluateEmptySlotAsZero"])
			}

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(PriorityFeeEstimate{PriorityFeeEstimate: 50000.0})
		}))
		defer server.Close()

		client, _ := NewClient("test-key", WithAPIURL(server.URL))
		_, err := client.GetPriorityFeeEstimateForTransaction(context.Background(),
			"tx-data",
			&GetPriorityFeeOptions{
				EvaluateEmptySlotAsZero: true,
			},
		)

		if err != nil {
			t.Fatalf("GetPriorityFeeEstimateForTransaction returned error: %v", err)
		}
	})
}

func TestCalculatePriorityFee(t *testing.T) {
	tests := []struct {
		name             string
		computeUnits     int64
		microLamportsPerCU float64
		expected         int64
	}{
		{
			name:               "standard calculation",
			computeUnits:       200_000,
			microLamportsPerCU: 50_000,
			expected:           10_000, // (200000 * 50000) / 1000000 = 10000
		},
		{
			name:               "zero compute units",
			computeUnits:       0,
			microLamportsPerCU: 50_000,
			expected:           0,
		},
		{
			name:               "zero fee",
			computeUnits:       200_000,
			microLamportsPerCU: 0,
			expected:           0,
		},
		{
			name:               "large values",
			computeUnits:       1_400_000,
			microLamportsPerCU: 100_000,
			expected:           140_000, // (1400000 * 100000) / 1000000 = 140000
		},
		{
			name:               "small values",
			computeUnits:       100,
			microLamportsPerCU: 1000,
			expected:           0, // (100 * 1000) / 1000000 = 0.1 -> 0
		},
		{
			name:               "fractional result truncation",
			computeUnits:       150_000,
			microLamportsPerCU: 33_333,
			expected:           4_999, // (150000 * 33333) / 1000000 = 4999.95 -> 4999
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePriorityFee(tt.computeUnits, tt.microLamportsPerCU)
			if result != tt.expected {
				t.Errorf("CalculatePriorityFee(%d, %f) = %d, want %d",
					tt.computeUnits, tt.microLamportsPerCU, result, tt.expected)
			}
		})
	}
}

func TestPriorityFeeEstimateTypes(t *testing.T) {
	t.Run("priority fee estimate", func(t *testing.T) {
		est := PriorityFeeEstimate{
			PriorityFeeEstimate: 50000.0,
		}
		if est.PriorityFeeEstimate != 50000.0 {
			t.Errorf("PriorityFeeEstimate = %f, unexpected value", est.PriorityFeeEstimate)
		}
	})

	t.Run("priority fee levels", func(t *testing.T) {
		levels := PriorityFeeLevels{
			Min:       1000.0,
			Low:       10000.0,
			Medium:    50000.0,
			High:      100000.0,
			VeryHigh:  200000.0,
			UnsafeMax: 500000.0,
		}
		if levels.Medium != 50000.0 {
			t.Errorf("Medium = %f, want 50000", levels.Medium)
		}
		if levels.VeryHigh != 200000.0 {
			t.Errorf("VeryHigh = %f, want 200000", levels.VeryHigh)
		}
	})

	t.Run("get priority fee options", func(t *testing.T) {
		opts := GetPriorityFeeOptions{
			TransactionEncoding:        "base64",
			PriorityLevel:              PriorityMedium,
			IncludeAllPriorityFeeLevels: true,
			LookbackSlots:              150,
			IncludeVote:                true,
			Recommended:                true,
			EvaluateEmptySlotAsZero:    true,
		}
		if opts.PriorityLevel != PriorityMedium {
			t.Errorf("PriorityLevel = %s, want Medium", opts.PriorityLevel)
		}
		if opts.LookbackSlots != 150 {
			t.Errorf("LookbackSlots = %d, want 150", opts.LookbackSlots)
		}
	})
}
