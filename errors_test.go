package helius

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "basic error",
			err: &APIError{
				StatusCode: 400,
				Message:    "invalid request",
				Path:       "/assets",
			},
			expected: "helius api error: /assets returned status 400: invalid request",
		},
		{
			name: "rate limited error",
			err: &APIError{
				StatusCode: 429,
				Message:    "rate limit exceeded",
				Path:       "/webhooks",
			},
			expected: "helius api error: /webhooks returned status 429: rate limit exceeded",
		},
		{
			name: "server error",
			err: &APIError{
				StatusCode: 500,
				Message:    "internal server error",
				Path:       "/priority-fee",
			},
			expected: "helius api error: /priority-fee returned status 500: internal server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsNotFound(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusNotFound, true},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsNotFound(); got != tt.expected {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsRateLimited(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusTooManyRequests, true},
		{http.StatusOK, false},
		{http.StatusNotFound, false},
		{http.StatusServiceUnavailable, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsRateLimited(); got != tt.expected {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsServerError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		{http.StatusServiceUnavailable, true},
		{http.StatusGatewayTimeout, true},
		{500, true},
		{599, true},
		{499, false},
		{600, false},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsServerError(); got != tt.expected {
				t.Errorf("IsServerError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsClientError(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusBadRequest, true},
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, true},
		{http.StatusNotFound, true},
		{http.StatusTooManyRequests, true},
		{400, true},
		{499, true},
		{399, false},
		{500, false},
		{http.StatusOK, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsClientError(); got != tt.expected {
				t.Errorf("IsClientError() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsUnauthorized(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusUnauthorized, true},
		{http.StatusForbidden, false},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsUnauthorized(); got != tt.expected {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestAPIError_IsForbidden(t *testing.T) {
	tests := []struct {
		statusCode int
		expected   bool
	}{
		{http.StatusForbidden, true},
		{http.StatusUnauthorized, false},
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.statusCode), func(t *testing.T) {
			err := &APIError{StatusCode: tt.statusCode}
			if got := err.IsForbidden(); got != tt.expected {
				t.Errorf("IsForbidden() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsAPIError(t *testing.T) {
	t.Run("direct api error", func(t *testing.T) {
		err := &APIError{StatusCode: 400, Message: "bad request", Path: "/test"}
		apiErr, ok := IsAPIError(err)
		if !ok {
			t.Error("IsAPIError should return true for APIError")
		}
		if apiErr.StatusCode != 400 {
			t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
		}
	})

	t.Run("wrapped api error", func(t *testing.T) {
		original := &APIError{StatusCode: 404, Message: "not found", Path: "/assets"}
		wrapped := fmt.Errorf("operation failed: %w", original)

		apiErr, ok := IsAPIError(wrapped)
		if !ok {
			t.Error("IsAPIError should return true for wrapped APIError")
		}
		if apiErr.StatusCode != 404 {
			t.Errorf("StatusCode = %d, want 404", apiErr.StatusCode)
		}
	})

	t.Run("non api error", func(t *testing.T) {
		err := errors.New("some other error")
		apiErr, ok := IsAPIError(err)
		if ok {
			t.Error("IsAPIError should return false for non-APIError")
		}
		if apiErr != nil {
			t.Error("apiErr should be nil for non-APIError")
		}
	})

	t.Run("nil error", func(t *testing.T) {
		apiErr, ok := IsAPIError(nil)
		if ok {
			t.Error("IsAPIError should return false for nil error")
		}
		if apiErr != nil {
			t.Error("apiErr should be nil for nil error")
		}
	})
}

func TestAPIError_ImplementsError(t *testing.T) {
	// Compile-time check that APIError implements error interface
	var _ error = (*APIError)(nil)
}
