package hooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// errorStatusResponse handles API responses that return HTTP 200 OK
// but contain "status": "error" in the body (Issues #3, #5 from KNOWN_API_ISSUES.md)
type errorStatusResponse struct{}

// APIErrorResponse represents the common error response structure from Netskope API
type APIErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	_ afterSuccessHook = (*errorStatusResponse)(nil)
)

// ErrResourceNotFound indicates a resource was not found (soft 404)
var ErrResourceNotFound = errors.New("resource not found")

func (e *errorStatusResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	// Only process 200 OK responses
	if res.StatusCode != http.StatusOK {
		return res, nil
	}

	// Read the response body
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read response body: %w", err)
	}

	// Check for error status in response
	var apiResponse APIErrorResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		// If we can't parse it, restore body and continue
		res.Body = io.NopCloser(bytes.NewReader(body))
		return res, nil
	}

	// If status is "error", convert to actual error
	if apiResponse.Status == "error" {
		// Determine error type based on message content
		msg := apiResponse.Message
		if msg == "" {
			msg = "API returned error status"
		}

		// Check for "not found" patterns to return specific error
		if isNotFoundError(msg) {
			return nil, fmt.Errorf("%w: %s", ErrResourceNotFound, msg)
		}

		return nil, fmt.Errorf("API error: %s", msg)
	}

	// Restore body for downstream processing
	res.Body = io.NopCloser(bytes.NewReader(body))
	return res, nil
}

// isNotFoundError checks if the error message indicates a missing resource
func isNotFoundError(msg string) bool {
	notFoundPatterns := []string{
		"not found",
		"No private app with id",
		"No publisher upgrade profile with id",
		"does not exist",
	}

	for _, pattern := range notFoundPatterns {
		if containsIgnoreCase(msg, pattern) {
			return true
		}
	}
	return false
}

func containsIgnoreCase(s, substr string) bool {
	return bytes.Contains(
		bytes.ToLower([]byte(s)),
		bytes.ToLower([]byte(substr)),
	)
}