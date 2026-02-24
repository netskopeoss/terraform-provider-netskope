package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// ruleCreateRetryClient wraps an HTTP client to retry rule creation requests
// that fail because a referenced resource (e.g., private app) hasn't propagated
// through the backend yet.
//
// The API returns HTTP 200 with {"status": "error", "message": "... doesn't exist"}
// when a rule references a private app that was just created but hasn't propagated
// to the rules service. This wrapper detects that pattern and retries with
// exponential backoff.
//
// Implemented as an HTTP client wrapper (via SDKInit hook) so retries happen
// transparently before any AfterSuccess hooks process the response.
//
// See docs/bugs/BUG-009-rule-after-app-eventual-consistency.md
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/65
type ruleCreateRetryClient struct {
	wrapped HTTPClient
	// sleepFn is the function used to sleep between retries.
	// Defaults to time.Sleep; overridden in tests.
	sleepFn func(time.Duration)
}

// ruleCreateRetryHook installs the retry client wrapper during SDK initialization.
type ruleCreateRetryHook struct{}

var _ sdkInitHook = (*ruleCreateRetryHook)(nil)

const (
	ruleRetryMaxAttempts  = 6
	ruleRetryBaseDelay    = 2 * time.Second
	ruleRetryMaxDelay     = 30 * time.Second
	ruleRetryRulesPath    = "/policy/npa/rules"
)

func (h *ruleCreateRetryHook) SDKInit(baseURL string, client HTTPClient) (string, HTTPClient) {
	return baseURL, &ruleCreateRetryClient{wrapped: client, sleepFn: time.Sleep}
}

func (c *ruleCreateRetryClient) Do(req *http.Request) (*http.Response, error) {
	if !isRuleCreateRequest(req) {
		return c.wrapped.Do(req)
	}

	// Buffer the request body for retries
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	req.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var resp *http.Response
	for attempt := 0; attempt < ruleRetryMaxAttempts; attempt++ {
		if attempt > 0 {
			delay := ruleRetryBackoff(attempt)
			log.Printf("Rule creation: referenced resource not yet available, retrying in %s (attempt %d/%d)",
				delay, attempt+1, ruleRetryMaxAttempts)
			c.sleepFn(delay)
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err = c.wrapped.Do(req)
		if err != nil {
			return resp, err
		}

		// Read response to check for retryable error
		respBody, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return resp, readErr
		}
		resp.Body = io.NopCloser(bytes.NewReader(respBody))

		if !isRetryableRuleError(respBody) {
			return resp, nil
		}
	}

	// Return the last response — errorStatusResponse will convert it to a Go error
	return resp, nil
}

// isRuleCreateRequest returns true for POST requests to the rules endpoint.
func isRuleCreateRequest(req *http.Request) bool {
	return req.Method == http.MethodPost &&
		strings.HasSuffix(req.URL.Path, ruleRetryRulesPath)
}

// isRetryableRuleError checks if the response body indicates a resource
// that doesn't exist yet (eventual consistency).
func isRetryableRuleError(body []byte) bool {
	var apiResp APIErrorResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return false
	}
	if apiResp.Status != "error" {
		return false
	}
	msg := strings.ToLower(apiResp.Message)
	return strings.Contains(msg, "doesn't exist") ||
		strings.Contains(msg, "does not exist")
}

// ruleRetryBackoff returns the delay for the given attempt using exponential backoff.
// Attempt 1: 2s, 2: 4s, 3: 8s, 4: 16s, 5: 30s (capped)
func ruleRetryBackoff(attempt int) time.Duration {
	delay := ruleRetryBaseDelay
	for i := 1; i < attempt; i++ {
		delay *= 2
	}
	if delay > ruleRetryMaxDelay {
		delay = ruleRetryMaxDelay
	}
	return delay
}
