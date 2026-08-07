package hooks

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakeErrorStatusResponse(body string, statusCode int, operationID string) (AfterSuccessContext, *http.Response) {
	res := &http.Response{
		StatusCode: statusCode,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader([]byte(body))),
	}
	ctx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, res
}

// TestErrorStatus_200WithErrorBodyReturnsError verifies that a 200 OK response
// containing "status": "error" is converted into an actual Go error.
// Without this hook, API failures are silently swallowed — the resource
// appears to succeed and the state is corrupted.
func TestErrorStatus_200WithErrorBodyReturnsError(t *testing.T) {
	hook := &errorStatusResponse{}

	body := `{"status": "error", "message": "Invalid configuration"}`
	ctx, res := buildFakeErrorStatusResponse(body, 200, "updateNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err == nil {
		t.Fatal("expected error for status=error body, got nil")
	}
	if result != nil {
		t.Errorf("expected nil response when error returned, got non-nil")
	}
	if !strings.Contains(err.Error(), "Invalid configuration") {
		t.Errorf("expected error message to contain 'Invalid configuration', got: %v", err)
	}
}

// TestErrorStatus_200WithSuccessBodyPassesThrough verifies that a normal 200 OK
// success response is passed through unchanged.
func TestErrorStatus_200WithSuccessBodyPassesThrough(t *testing.T) {
	hook := &errorStatusResponse{}

	body := `{"status": "success", "data": {"id": 1}}`
	ctx, res := buildFakeErrorStatusResponse(body, 200, "getNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("expected nil error for success body, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response for success body")
	}

	// Verify the body is still readable downstream
	raw, _ := io.ReadAll(result.Body)
	if !strings.Contains(string(raw), `"success"`) {
		t.Errorf("expected success body preserved, got: %s", raw)
	}
}

// TestErrorStatus_Non200SkipsProcessing verifies that non-200 responses (e.g. 404, 500)
// are returned as-is without any body inspection.
func TestErrorStatus_Non200SkipsProcessing(t *testing.T) {
	hook := &errorStatusResponse{}

	body := `{"error": "not found"}`
	ctx, res := buildFakeErrorStatusResponse(body, 404, "getNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("expected nil error for non-200 status, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response for non-200 passthrough")
	}
}

// TestErrorStatus_NotFoundPatternReturnsErrResourceNotFound verifies that "not found"
// messages in error bodies produce the sentinel ErrResourceNotFound, enabling callers
// to implement soft-delete / import logic that distinguishes 404 from other errors.
func TestErrorStatus_NotFoundPatternReturnsErrResourceNotFound(t *testing.T) {
	hook := &errorStatusResponse{}

	notFoundMessages := []string{
		`{"status": "error", "message": "No private app with id 999 not found"}`,
		`{"status": "error", "message": "Resource does not exist"}`,
		`{"status": "error", "message": "No publisher upgrade profile with id 42"}`,
	}

	for _, body := range notFoundMessages {
		ctx, res := buildFakeErrorStatusResponse(body, 200, "getNPAPrivateApp")
		_, err := hook.AfterSuccess(ctx, res)
		if err == nil {
			t.Errorf("expected error for body %q, got nil", body)
			continue
		}
		// Should wrap ErrResourceNotFound
		if !strings.Contains(err.Error(), ErrResourceNotFound.Error()) {
			t.Errorf("expected ErrResourceNotFound wrapped, got: %v (body: %s)", err, body)
		}
	}
}

// TestErrorStatus_EmptyMessageFallback verifies that an empty message in the error body
// uses a default message rather than returning an empty error string.
func TestErrorStatus_EmptyMessageFallback(t *testing.T) {
	hook := &errorStatusResponse{}

	body := `{"status": "error", "message": ""}`
	ctx, res := buildFakeErrorStatusResponse(body, 200, "createNPAPrivateApps")

	_, err := hook.AfterSuccess(ctx, res)
	if err == nil {
		t.Fatal("expected error for status=error body with empty message")
	}
	if err.Error() == "API error: " {
		t.Errorf("expected non-empty error message fallback, got: %v", err)
	}
}

// TestErrorStatus_InvalidJSONBodyPassesThrough verifies that if the body cannot be
// parsed as JSON, the hook restores the body and passes through rather than crashing.
func TestErrorStatus_InvalidJSONBodyPassesThrough(t *testing.T) {
	hook := &errorStatusResponse{}

	body := `not valid json`
	ctx, res := buildFakeErrorStatusResponse(body, 200, "getNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("expected nil error for unparseable body, got: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response for unparseable body passthrough")
	}
}
