package helius

import (
	"errors"
	"fmt"
	"net/http"
)

// APIError represents an error returned by the Helius API.
type APIError struct {
	// StatusCode is the HTTP status code.
	StatusCode int

	// Message is the error message from the API.
	Message string

	// Path is the API endpoint that returned the error.
	Path string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	return fmt.Sprintf("helius api error: %s returned status %d: %s", e.Path, e.StatusCode, e.Message)
}

// IsNotFound returns true if the error is a 404 Not Found.
func (e *APIError) IsNotFound() bool {
	return e.StatusCode == http.StatusNotFound
}

// IsRateLimited returns true if the error is a 429 Too Many Requests.
func (e *APIError) IsRateLimited() bool {
	return e.StatusCode == http.StatusTooManyRequests
}

// IsServerError returns true if the error is a 5xx server error.
func (e *APIError) IsServerError() bool {
	return e.StatusCode >= 500 && e.StatusCode < 600
}

// IsClientError returns true if the error is a 4xx client error.
func (e *APIError) IsClientError() bool {
	return e.StatusCode >= 400 && e.StatusCode < 500
}

// IsUnauthorized returns true if the error is a 401 Unauthorized.
func (e *APIError) IsUnauthorized() bool {
	return e.StatusCode == http.StatusUnauthorized
}

// IsForbidden returns true if the error is a 403 Forbidden.
func (e *APIError) IsForbidden() bool {
	return e.StatusCode == http.StatusForbidden
}

// IsAPIError checks if an error is an APIError and returns it.
// This works with wrapped errors using errors.As.
func IsAPIError(err error) (*APIError, bool) {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr, true
	}
	return nil, false
}
