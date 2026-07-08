package hooks

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/internal/config"
)

// fakeHTTPClient wraps a function as a config.HTTPClient for testing.
type fakeHTTPClient func(*http.Request) (*http.Response, error)

func (f fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return f(req)
}

// fakePublisherGetClient returns a client that responds with a publisher GET body
// containing the given name and connected_apps value.
func fakePublisherGetClient(t *testing.T, name string, connectedApps interface{}) config.HTTPClient {
	t.Helper()
	return fakeHTTPClient(func(req *http.Request) (*http.Response, error) {
		body := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"name":           name,
				"connected_apps": connectedApps,
			},
		}
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal test response: %v", err)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(string(b))),
			Header:     make(http.Header),
		}, nil
	})
}

// deletePublisherCtx builds a BeforeRequestContext for a publisher DELETE.
func deletePublisherCtx(operationID string, client config.HTTPClient) BeforeRequestContext {
	return BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
			Context:     context.Background(),
			SDKConfiguration: config.SDKConfiguration{
				Client: client,
			},
		},
	}
}

// deletePublisherReq builds a DELETE request for a given publisher ID.
func deletePublisherReq(t *testing.T, publisherID int) *http.Request {
	t.Helper()
	url := fmt.Sprintf("https://example.goskope.com/api/v2/infrastructure/publishers/%d", publisherID)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, url, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Netskope-Api-Token", "test-token")
	return req
}

// ============================================================================
// BeforeRequest tests
// ============================================================================

// TestPublisherDeleteGuard_NonMatchingOp verifies that non-delete operations
// pass through without any HTTP check being performed.
func TestPublisherDeleteGuard_NonMatchingOp(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)

	// Client would panic if called — proves it is never invoked.
	panicClient := fakeHTTPClient(func(_ *http.Request) (*http.Response, error) {
		t.Fatal("GET should not be issued for non-delete operations")
		return nil, nil
	})
	ctx := deletePublisherCtx("getNPAPublisherById", panicClient)

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != req {
		t.Error("expected original request to be returned unchanged")
	}
}

// TestPublisherDeleteGuard_NoConnectedApps verifies that a publisher with an
// empty connected_apps array passes through without error.
func TestPublisherDeleteGuard_NoConnectedApps(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	ctx := deletePublisherCtx("deleteNPAPublishers", fakePublisherGetClient(t, "my-publisher", []interface{}{}))

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error for publisher with no apps, got %v", err)
	}
	if result == nil {
		t.Error("expected original request to be returned")
	}
}

// TestPublisherDeleteGuard_ConnectedAppsAsStrings verifies that the hook blocks
// a delete and names the connected apps when connected_apps is an array of strings
// (the format returned by the list endpoint).
func TestPublisherDeleteGuard_ConnectedAppsAsStrings(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	ctx := deletePublisherCtx("deleteNPAPublishers", fakePublisherGetClient(t, "my-publisher", []interface{}{
		"[AppA]", "[AppB]",
	}))

	_, err := hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error for publisher with connected apps (string form)")
	}
	for _, want := range []string{"my-publisher", "[AppA]", "[AppB]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error to contain %q, got: %v", want, err)
		}
	}
}

// TestPublisherDeleteGuard_ConnectedAppsAsObjects verifies that the hook blocks
// a delete and names the connected apps when connected_apps is an array of objects
// (the format returned by the get-by-ID endpoint, introduced in BUG-017).
func TestPublisherDeleteGuard_ConnectedAppsAsObjects(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	ctx := deletePublisherCtx("deleteNPAPublishers", fakePublisherGetClient(t, "my-publisher", []interface{}{
		map[string]interface{}{"name": "[AppA]", "access_method": "client", "host": "172.20.36.40", "last_connected": nil},
		map[string]interface{}{"name": "[AppB]", "access_method": "client", "host": "10.0.0.2", "last_connected": nil},
	}))

	_, err := hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error for publisher with connected apps (object form)")
	}
	for _, want := range []string{"my-publisher", "[AppA]", "[AppB]"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("expected error to contain %q, got: %v", want, err)
		}
	}
}

// TestPublisherDeleteGuard_AppCountInError verifies that the error message
// includes the correct count of connected apps.
func TestPublisherDeleteGuard_AppCountInError(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	ctx := deletePublisherCtx("deleteNPAPublishers", fakePublisherGetClient(t, "my-publisher", []interface{}{
		"[App1]", "[App2]", "[App3]",
	}))

	_, err := hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "3 app(s)") {
		t.Errorf("expected error to report 3 apps, got: %v", err)
	}
}

// TestPublisherDeleteGuard_GetFails verifies that a network error on the
// pre-check GET does not block the delete — the hook passes through.
func TestPublisherDeleteGuard_GetFails(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	client := fakeHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("network error")
	})
	ctx := deletePublisherCtx("deleteNPAPublishers", client)

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected hook to pass through on GET failure, got %v", err)
	}
	if result == nil {
		t.Error("expected original request to pass through")
	}
}

// TestPublisherDeleteGuard_GetNon200 verifies that a non-200 from the pre-check
// GET does not block the delete.
func TestPublisherDeleteGuard_GetNon200(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	client := fakeHTTPClient(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusNotFound,
			Body:       io.NopCloser(strings.NewReader(`{"status":"not found"}`)),
			Header:     make(http.Header),
		}, nil
	})
	ctx := deletePublisherCtx("deleteNPAPublishers", client)

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected hook to pass through on non-200, got %v", err)
	}
	if result == nil {
		t.Error("expected original request to pass through")
	}
}

// TestPublisherDeleteGuard_NilClient verifies that a nil HTTP client (e.g. in
// unit tests that construct minimal SDK configs) does not panic.
func TestPublisherDeleteGuard_NilClient(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	ctx := deletePublisherCtx("deleteNPAPublishers", nil)

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("expected no error with nil client, got %v", err)
	}
	if result == nil {
		t.Error("expected original request to pass through")
	}
}

// TestPublisherDeleteGuard_AuthHeaderCopied verifies that the Netskope-Api-Token
// header from the DELETE request is forwarded to the pre-check GET.
func TestPublisherDeleteGuard_AuthHeaderCopied(t *testing.T) {
	hook := &publisherDeleteGuardHook{}
	req := deletePublisherReq(t, 10950)
	req.Header.Set("Netskope-Api-Token", "secret-token")

	var capturedToken string
	client := fakeHTTPClient(func(r *http.Request) (*http.Response, error) {
		capturedToken = r.Header.Get("Netskope-Api-Token")
		body := `{"status":"success","data":{"name":"pub","connected_apps":[]}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})
	ctx := deletePublisherCtx("deleteNPAPublishers", client)

	_, _ = hook.BeforeRequest(ctx, req)
	if capturedToken != "secret-token" {
		t.Errorf("expected Netskope-Api-Token to be forwarded, got %q", capturedToken)
	}
}

// ============================================================================
// extractConnectedAppNames tests
// ============================================================================

func TestExtractConnectedAppNames_Strings(t *testing.T) {
	elems := jsonRawMessages(t, []interface{}{"[AppA]", "[AppB]"})
	names := extractConnectedAppNames(elems)
	if len(names) != 2 || names[0] != "[AppA]" || names[1] != "[AppB]" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestExtractConnectedAppNames_Objects(t *testing.T) {
	elems := jsonRawMessages(t, []interface{}{
		map[string]interface{}{"name": "[AppA]", "access_method": "client"},
		map[string]interface{}{"name": "[AppB]", "access_method": "client"},
	})
	names := extractConnectedAppNames(elems)
	if len(names) != 2 || names[0] != "[AppA]" || names[1] != "[AppB]" {
		t.Errorf("unexpected names: %v", names)
	}
}

func TestExtractConnectedAppNames_Empty(t *testing.T) {
	names := extractConnectedAppNames(nil)
	if len(names) != 0 {
		t.Errorf("expected empty slice, got %v", names)
	}
}

func TestExtractConnectedAppNames_ObjectMissingName(t *testing.T) {
	elems := jsonRawMessages(t, []interface{}{
		map[string]interface{}{"access_method": "client"}, // no name field
	})
	names := extractConnectedAppNames(elems)
	if len(names) != 0 {
		t.Errorf("expected no names from object without name field, got %v", names)
	}
}

// jsonRawMessages marshals a slice of values into []json.RawMessage for testing.
func jsonRawMessages(t *testing.T, vals []interface{}) []json.RawMessage {
	t.Helper()
	out := make([]json.RawMessage, len(vals))
	for i, v := range vals {
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("failed to marshal test value: %v", err)
		}
		out[i] = json.RawMessage(b)
	}
	return out
}
