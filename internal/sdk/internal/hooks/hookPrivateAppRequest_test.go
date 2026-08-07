package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakePrivateAppRequest(body string, operationID string) (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("PUT", "https://example.com/api/v2/privateapps/1", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, req
}

// TestPrivateAppRequest_NonClientlessRemovesAppOption verifies that app_option is
// removed from the payload when clientless_access is false. The API rejects
// app_option for regular (non-clientless) apps with an explicit error.
func TestPrivateAppRequest_NonClientlessRemovesAppOption(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": false, "app_option": {"key": "value"}}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["app_option"]; exists {
		t.Errorf("expected app_option removed for non-clientless app, still present in: %s", raw)
	}
}

// TestPrivateAppRequest_ClientlessKeepsNonEmptyAppOption verifies that app_option
// is preserved for clientless apps when it has content.
func TestPrivateAppRequest_ClientlessKeepsNonEmptyAppOption(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": true, "app_option": {"option_key": "value"}}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["app_option"]; !exists {
		t.Errorf("expected app_option preserved for clientless app, missing from: %s", raw)
	}
}

// TestPrivateAppRequest_EmptyPathsStripped verifies that empty paths array is removed.
// The API has a bug where empty arrays for paths/bypass_uris are stored as bytes
// rather than JSON, causing SQL serialization errors.
func TestPrivateAppRequest_EmptyPathsStripped(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": false, "paths": [], "bypass_uris": []}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["paths"]; exists {
		t.Errorf("expected empty paths removed, still present in: %s", raw)
	}
	if _, exists := out["bypass_uris"]; exists {
		t.Errorf("expected empty bypass_uris removed, still present in: %s", raw)
	}
}

// TestPrivateAppRequest_NonEmptyPathsPreserved verifies that non-empty paths are kept.
func TestPrivateAppRequest_NonEmptyPathsPreserved(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": true, "paths": ["/api/v1"]}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["paths"]; !exists {
		t.Errorf("expected non-empty paths preserved, missing from: %s", raw)
	}
}

// TestPrivateAppRequest_EmptyAppOptionOnClientlessStripped verifies that even for
// clientless apps, an empty app_option map is stripped to prevent serialization errors.
func TestPrivateAppRequest_EmptyAppOptionOnClientlessStripped(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": true, "app_option": {}}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["app_option"]; exists {
		t.Errorf("expected empty app_option stripped for clientless app, still present in: %s", raw)
	}
}

// TestPrivateAppRequest_NullUribypassHeaderStripped verifies null uribypass_header_value
// is removed from the payload.
func TestPrivateAppRequest_NullUribypassHeaderStripped(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "clientless_access": false, "uribypass_header_value": null}`
	ctx, req := buildFakePrivateAppRequest(body, "updateNPAPrivateApp")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["uribypass_header_value"]; exists {
		t.Errorf("expected null uribypass_header_value removed, still present in: %s", raw)
	}
}

// TestPrivateAppRequest_NonMatchingOperationPassthrough verifies that operations
// other than updateNPAPrivateApp are passed through unchanged.
func TestPrivateAppRequest_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &privateAppRequestHook{}

	body := `{"app_name": "my-app", "paths": [], "app_option": {"key": "val"}}`
	ctx, req := buildFakePrivateAppRequest(body, "createNPAPrivateApps")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed for non-matching operation: %v", err)
	}

	// Body should be unchanged — empty paths and app_option NOT stripped
	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["paths"]; !exists {
		t.Errorf("expected paths unchanged for non-matching operation, missing from: %s", raw)
	}
}
