package hooks

import (
	"bytes"
	"io"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func noSleep(time.Duration) {}

func testRetryClient(mock HTTPClient) *ruleCreateRetryClient {
	return &ruleCreateRetryClient{wrapped: mock, sleepFn: noSleep}
}

// mockHTTPClient records calls and returns configurable responses.
type mockHTTPClient struct {
	responses []*http.Response
	callCount atomic.Int32
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	idx := int(m.callCount.Add(1)) - 1
	if idx >= len(m.responses) {
		idx = len(m.responses) - 1
	}
	// Consume request body to simulate real HTTP call
	io.ReadAll(req.Body)
	return m.responses[idx], nil
}

func jsonResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
}

func ruleCreateRequest() *http.Request {
	req, _ := http.NewRequest("POST", "https://tenant.goskope.com/api/v2/policy/npa/rules",
		bytes.NewReader([]byte(`{"rule_name":"test-rule"}`)))
	return req
}

// TestRetryClient_SuccessOnFirstAttempt verifies no retry when the first request succeeds.
func TestRetryClient_SuccessOnFirstAttempt(t *testing.T) {
	mock := &mockHTTPClient{
		responses: []*http.Response{
			jsonResponse(`{"status":"success","data":{"rule_id":1}}`),
		},
	}
	client := testRetryClient(mock)

	resp, err := client.Do(ruleCreateRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if mock.callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount.Load())
	}
}

// TestRetryClient_RetriesOnDoesntExist verifies the client retries when the API
// returns "doesn't exist" and eventually succeeds.
func TestRetryClient_RetriesOnDoesntExist(t *testing.T) {
	mock := &mockHTTPClient{
		responses: []*http.Response{
			jsonResponse(`{"status":"error","message":"Private app [my-app] doesn't exist"}`),
			jsonResponse(`{"status":"error","message":"Private app [my-app] doesn't exist"}`),
			jsonResponse(`{"status":"success","data":{"rule_id":1}}`),
		},
	}
	client := testRetryClient(mock)

	// Override backoff for testing — we test backoff calculation separately
	origBase := ruleRetryBaseDelay
	defer func() { /* restore not needed — const */ }()
	_ = origBase

	resp, err := client.Do(ruleCreateRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("success")) {
		t.Errorf("expected success response, got: %s", body)
	}
	if mock.callCount.Load() != 3 {
		t.Errorf("expected 3 calls (2 retries), got %d", mock.callCount.Load())
	}
}

// TestRetryClient_ReturnsLastResponseOnMaxRetries verifies that after exhausting
// retries, the last error response is returned (not swallowed).
func TestRetryClient_ReturnsLastResponseOnMaxRetries(t *testing.T) {
	errorResp := `{"status":"error","message":"Private app [my-app] doesn't exist"}`
	responses := make([]*http.Response, ruleRetryMaxAttempts)
	for i := range responses {
		responses[i] = jsonResponse(errorResp)
	}

	mock := &mockHTTPClient{responses: responses}
	client := testRetryClient(mock)

	resp, err := client.Do(ruleCreateRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("doesn't exist")) {
		t.Errorf("expected error response to be returned, got: %s", body)
	}
	if int(mock.callCount.Load()) != ruleRetryMaxAttempts {
		t.Errorf("expected %d calls, got %d", ruleRetryMaxAttempts, mock.callCount.Load())
	}
}

// TestRetryClient_DoesNotRetryNonCreateRequests verifies that GET, PUT, DELETE
// requests are passed through without retry logic.
func TestRetryClient_DoesNotRetryNonCreateRequests(t *testing.T) {
	methods := []string{"GET", "PUT", "DELETE"}
	for _, method := range methods {
		mock := &mockHTTPClient{
			responses: []*http.Response{
				jsonResponse(`{"status":"error","message":"doesn't exist"}`),
			},
		}
		client := testRetryClient(mock)

		req, _ := http.NewRequest(method, "https://tenant.goskope.com/api/v2/policy/npa/rules",
			bytes.NewReader([]byte(`{}`)))
		_, err := client.Do(req)
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", method, err)
		}
		if mock.callCount.Load() != 1 {
			t.Errorf("%s: expected 1 call (no retry), got %d", method, mock.callCount.Load())
		}
	}
}

// TestRetryClient_DoesNotRetryNonRulesEndpoint verifies that POST requests
// to other endpoints are not retried.
func TestRetryClient_DoesNotRetryNonRulesEndpoint(t *testing.T) {
	mock := &mockHTTPClient{
		responses: []*http.Response{
			jsonResponse(`{"status":"error","message":"doesn't exist"}`),
		},
	}
	client := testRetryClient(mock)

	req, _ := http.NewRequest("POST", "https://tenant.goskope.com/api/v2/steering/apps/private",
		bytes.NewReader([]byte(`{}`)))
	_, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.callCount.Load() != 1 {
		t.Errorf("expected 1 call (no retry), got %d", mock.callCount.Load())
	}
}

// TestRetryClient_DoesNotRetryNonExistenceErrors verifies that other API errors
// (not "doesn't exist") are not retried.
func TestRetryClient_DoesNotRetryNonExistenceErrors(t *testing.T) {
	errors := []string{
		`{"status":"error","message":"Invalid field value"}`,
		`{"status":"error","message":"Duplicate entry '2' for key 'PRIMARY'"}`,
		`{"status":"error","message":"Permission denied"}`,
	}
	for _, errBody := range errors {
		mock := &mockHTTPClient{
			responses: []*http.Response{jsonResponse(errBody)},
		}
		client := testRetryClient(mock)

		_, err := client.Do(ruleCreateRequest())
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", errBody, err)
		}
		if mock.callCount.Load() != 1 {
			t.Errorf("expected 1 call (no retry) for %s, got %d", errBody, mock.callCount.Load())
		}
	}
}

// TestRetryClient_DoesNotRetrySuccessResponse verifies success responses pass through.
func TestRetryClient_DoesNotRetrySuccessResponse(t *testing.T) {
	mock := &mockHTTPClient{
		responses: []*http.Response{
			jsonResponse(`{"status":"success","data":{"rule_id":5}}`),
		},
	}
	client := testRetryClient(mock)

	resp, err := client.Do(ruleCreateRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	if !bytes.Contains(body, []byte("success")) {
		t.Errorf("expected success response, got: %s", body)
	}
	if mock.callCount.Load() != 1 {
		t.Errorf("expected 1 call, got %d", mock.callCount.Load())
	}
}

// TestRetryClient_PreservesRequestBody verifies the request body is correctly
// restored for each retry attempt.
func TestRetryClient_PreservesRequestBody(t *testing.T) {
	originalBody := `{"rule_name":"my-important-rule","rule_data":{"privateApps":["[app-1]"]}}`
	var capturedBodies []string

	mock := &mockHTTPClient{
		responses: []*http.Response{
			jsonResponse(`{"status":"error","message":"Private app [app-1] doesn't exist"}`),
			jsonResponse(`{"status":"success","data":{"rule_id":1}}`),
		},
	}

	// Wrap mock to capture request bodies
	capturing := &bodyCapturingClient{wrapped: mock, bodies: &capturedBodies}
	client := testRetryClient(capturing)

	req, _ := http.NewRequest("POST", "https://tenant.goskope.com/api/v2/policy/npa/rules",
		bytes.NewReader([]byte(originalBody)))
	_, err := client.Do(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(capturedBodies) != 2 {
		t.Fatalf("expected 2 captured bodies, got %d", len(capturedBodies))
	}
	for i, body := range capturedBodies {
		if body != originalBody {
			t.Errorf("attempt %d: request body was not preserved, got: %s", i+1, body)
		}
	}
}

// bodyCapturingClient wraps an HTTP client and records the request body of each call.
type bodyCapturingClient struct {
	wrapped *mockHTTPClient
	bodies  *[]string
}

func (c *bodyCapturingClient) Do(req *http.Request) (*http.Response, error) {
	body, _ := io.ReadAll(req.Body)
	*c.bodies = append(*c.bodies, string(body))
	req.Body = io.NopCloser(bytes.NewReader(body))
	return c.wrapped.Do(req)
}

// TestRetryBackoff verifies the exponential backoff calculation.
func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 16 * time.Second},
		{5, 30 * time.Second},  // capped
		{6, 30 * time.Second},  // capped
		{10, 30 * time.Second}, // capped
	}

	for _, tt := range tests {
		got := ruleRetryBackoff(tt.attempt)
		if got != tt.expected {
			t.Errorf("ruleRetryBackoff(%d) = %s, want %s", tt.attempt, got, tt.expected)
		}
	}
}

// TestIsRetryableRuleError verifies the error detection logic.
func TestIsRetryableRuleError(t *testing.T) {
	tests := []struct {
		body     string
		expected bool
	}{
		{`{"status":"error","message":"Private app [my-app] doesn't exist"}`, true},
		{`{"status":"error","message":"Private app [my-app] does not exist"}`, true},
		{`{"status":"error","message":"PRIVATE APP [X] DOESN'T EXIST"}`, true},
		{`{"status":"success","data":{}}`, false},
		{`{"status":"error","message":"Duplicate entry"}`, false},
		{`{"status":"error","message":"Invalid field"}`, false},
		{`not json`, false},
		{`{}`, false},
	}

	for _, tt := range tests {
		got := isRetryableRuleError([]byte(tt.body))
		if got != tt.expected {
			t.Errorf("isRetryableRuleError(%s) = %v, want %v", tt.body, got, tt.expected)
		}
	}
}

// TestSDKInitHook_WrapsClient verifies the SDKInit hook wraps the HTTP client.
func TestSDKInitHook_WrapsClient(t *testing.T) {
	hook := &ruleCreateRetryHook{}
	mock := &mockHTTPClient{}

	baseURL, client := hook.SDKInit("https://example.com", mock)
	if baseURL != "https://example.com" {
		t.Errorf("expected baseURL unchanged, got %s", baseURL)
	}

	retryClient, ok := client.(*ruleCreateRetryClient)
	if !ok {
		t.Fatal("expected client to be wrapped in ruleCreateRetryClient")
	}
	if retryClient.wrapped != mock {
		t.Error("expected wrapped client to be the original mock")
	}
}
