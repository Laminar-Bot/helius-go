package helius

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
)

// PriorityLevel represents the priority level for fee estimation.
type PriorityLevel string

const (
	// PriorityMin is the minimum priority level (lowest fees).
	PriorityMin PriorityLevel = "Min"
	// PriorityLow is a low priority level.
	PriorityLow PriorityLevel = "Low"
	// PriorityMedium is a medium priority level (recommended).
	PriorityMedium PriorityLevel = "Medium"
	// PriorityHigh is a high priority level.
	PriorityHigh PriorityLevel = "High"
	// PriorityVeryHigh is a very high priority level.
	PriorityVeryHigh PriorityLevel = "VeryHigh"
	// PriorityUnsafeMax is the maximum priority (highest fees, use with caution).
	PriorityUnsafeMax PriorityLevel = "UnsafeMax"
)

// PriorityFeeEstimate contains the estimated priority fees.
type PriorityFeeEstimate struct {
	// PriorityFeeEstimate is the recommended fee in microlamports per compute unit.
	PriorityFeeEstimate float64 `json:"priorityFeeEstimate"`

	// PriorityFeeLevels contains fees for different priority levels.
	PriorityFeeLevels *PriorityFeeLevels `json:"priorityFeeLevels,omitempty"`
}

// PriorityFeeLevels contains fees for each priority level.
type PriorityFeeLevels struct {
	Min       float64 `json:"min"`
	Low       float64 `json:"low"`
	Medium    float64 `json:"medium"`
	High      float64 `json:"high"`
	VeryHigh  float64 `json:"veryHigh"`
	UnsafeMax float64 `json:"unsafeMax"`
}

// GetPriorityFeeOptions configures the priority fee estimation request.
type GetPriorityFeeOptions struct {
	// TransactionEncoding is the encoding of the transaction (base58 or base64).
	TransactionEncoding string `json:"transactionEncoding,omitempty"`

	// PriorityLevel is the desired priority level for the estimate.
	PriorityLevel PriorityLevel `json:"priorityLevel,omitempty"`

	// IncludeAllPriorityFeeLevels returns estimates for all priority levels.
	IncludeAllPriorityFeeLevels bool `json:"includeAllPriorityFeeLevels,omitempty"`

	// LookbackSlots is the number of slots to look back for fee data (default: 150).
	LookbackSlots int `json:"lookbackSlots,omitempty"`

	// IncludeVote includes vote transactions in the fee analysis.
	IncludeVote bool `json:"includeVote,omitempty"`

	// Recommended uses Helius's recommended fee algorithm.
	Recommended bool `json:"recommended,omitempty"`

	// EvaluateEmptySlotAsZero treats empty slots as zero fee instead of skipping.
	EvaluateEmptySlotAsZero bool `json:"evaluateEmptySlotAsZero,omitempty"`
}

// GetPriorityFeeEstimate gets the estimated priority fee for a transaction.
//
// You can either provide account addresses that the transaction will access,
// or provide a serialized transaction.
//
// Example with accounts:
//
//	estimate, err := client.GetPriorityFeeEstimate(ctx, []string{
//	    "JUP4Fb2cqiRUcaTHdrPC8h2gNsA2ETXiPDD33WcGuJB",
//	}, nil)
//
// Example with transaction:
//
//	estimate, err := client.GetPriorityFeeEstimateForTransaction(ctx, txBase64, &helius.GetPriorityFeeOptions{
//	    PriorityLevel: helius.PriorityMedium,
//	})
func (c *Client) GetPriorityFeeEstimate(ctx context.Context, accountKeys []string, opts *GetPriorityFeeOptions) (*PriorityFeeEstimate, error) {
	if len(accountKeys) == 0 {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "at least one account key is required",
			Path:       "/priority-fee",
		}
	}

	reqBody := map[string]interface{}{
		"accountKeys": accountKeys,
	}

	if opts != nil {
		if opts.PriorityLevel != "" {
			reqBody["options"] = map[string]interface{}{
				"priorityLevel": opts.PriorityLevel,
			}
		}
		if opts.IncludeAllPriorityFeeLevels {
			if reqBody["options"] == nil {
				reqBody["options"] = map[string]interface{}{}
			}
			reqBody["options"].(map[string]interface{})["includeAllPriorityFeeLevels"] = true
		}
		if opts.LookbackSlots > 0 {
			if reqBody["options"] == nil {
				reqBody["options"] = map[string]interface{}{}
			}
			reqBody["options"].(map[string]interface{})["lookbackSlots"] = opts.LookbackSlots
		}
		if opts.Recommended {
			if reqBody["options"] == nil {
				reqBody["options"] = map[string]interface{}{}
			}
			reqBody["options"].(map[string]interface{})["recommended"] = true
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, err := c.doRequest(ctx, "POST", "/priority-fee", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var estimate PriorityFeeEstimate
	if err := json.Unmarshal(body, &estimate); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("got priority fee estimate",
		"fee", estimate.PriorityFeeEstimate,
		"accounts", len(accountKeys),
	)

	return &estimate, nil
}

// GetPriorityFeeEstimateForTransaction gets the estimated priority fee for a serialized transaction.
func (c *Client) GetPriorityFeeEstimateForTransaction(ctx context.Context, transaction string, opts *GetPriorityFeeOptions) (*PriorityFeeEstimate, error) {
	if transaction == "" {
		return nil, &APIError{
			StatusCode: 400,
			Message:    "transaction is required",
			Path:       "/priority-fee",
		}
	}

	reqBody := map[string]interface{}{
		"transaction": transaction,
	}

	if opts != nil {
		options := map[string]interface{}{}

		if opts.TransactionEncoding != "" {
			options["transactionEncoding"] = opts.TransactionEncoding
		}
		if opts.PriorityLevel != "" {
			options["priorityLevel"] = opts.PriorityLevel
		}
		if opts.IncludeAllPriorityFeeLevels {
			options["includeAllPriorityFeeLevels"] = true
		}
		if opts.LookbackSlots > 0 {
			options["lookbackSlots"] = opts.LookbackSlots
		}
		if opts.Recommended {
			options["recommended"] = true
		}
		if opts.EvaluateEmptySlotAsZero {
			options["evaluateEmptySlotAsZero"] = true
		}

		if len(options) > 0 {
			reqBody["options"] = options
		}
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	body, err := c.doRequest(ctx, "POST", "/priority-fee", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}

	var estimate PriorityFeeEstimate
	if err := json.Unmarshal(body, &estimate); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	c.logger.Debug("got priority fee estimate for transaction",
		"fee", estimate.PriorityFeeEstimate,
	)

	return &estimate, nil
}

// CalculatePriorityFee calculates the total priority fee in lamports for a transaction.
//
// Formula: priority_fee = (compute_units * micro_lamports_per_cu) / 1_000_000
//
// Example:
//
//	fee := helius.CalculatePriorityFee(200_000, 50_000) // 10,000 lamports
func CalculatePriorityFee(computeUnits int64, microLamportsPerCU float64) int64 {
	return int64(float64(computeUnits) * microLamportsPerCU / 1_000_000)
}
