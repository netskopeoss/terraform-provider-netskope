package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

// --- BeforeRequest (deviceTagRequestHook) tests ---

func buildDeviceTagGetRequest(path string) (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("GET", "https://example.com"+path, nil)
	req.Header.Set("Netskope-Api-Token", "test-token")
	ctx := BeforeRequestContext{
		HookContext: HookContext{OperationID: "getDeviceTag"},
	}
	return ctx, req
}

func buildDeviceTagListRequest() (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("GET", "https://example.com/api/v2/device/tags", nil)
	req.Header.Set("Netskope-Api-Token", "test-token")
	ctx := BeforeRequestContext{
		HookContext: HookContext{OperationID: "listDeviceTags"},
	}
	return ctx, req
}

// TestDeviceTag_GetRewritesToPost verifies that GET /device/tags/{id} is rewritten
// to POST /device/tags/gettags with the correct JSON body.
func TestDeviceTag_GetRewritesToPost(t *testing.T) {
	hook := &deviceTagRequestHook{}
	ctx, req := buildDeviceTagGetRequest("/api/v2/device/tags/42")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	if result.Method != http.MethodPost {
		t.Errorf("expected method POST, got %s", result.Method)
	}
	if !strings.HasSuffix(result.URL.Path, "/device/tags/gettags") {
		t.Errorf("expected path to end with /device/tags/gettags, got %s", result.URL.Path)
	}

	body, _ := io.ReadAll(result.Body)
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, body)
	}
	if idVal, ok := payload["id"]; !ok {
		t.Error("expected 'id' key in POST body")
	} else if idVal != float64(42) {
		t.Errorf("expected id=42, got %v", idVal)
	}
}

// TestDeviceTag_GetPreservesAuthHeader verifies that the API token is carried over
// to the rewritten POST request.
func TestDeviceTag_GetPreservesAuthHeader(t *testing.T) {
	hook := &deviceTagRequestHook{}
	ctx, req := buildDeviceTagGetRequest("/api/v2/device/tags/7")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}
	if result.Header.Get("Netskope-Api-Token") != "test-token" {
		t.Errorf("expected auth header to be preserved, got: %q", result.Header.Get("Netskope-Api-Token"))
	}
}

// TestDeviceTag_ListRewritesToPost verifies that GET /device/tags is rewritten to
// POST /device/tags/gettags with {"limit": 100}.
func TestDeviceTag_ListRewritesToPost(t *testing.T) {
	hook := &deviceTagRequestHook{}
	ctx, req := buildDeviceTagListRequest()

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	if result.Method != http.MethodPost {
		t.Errorf("expected method POST, got %s", result.Method)
	}

	body, _ := io.ReadAll(result.Body)
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("body is not valid JSON: %v\nbody: %s", err, body)
	}
	if limitVal, ok := payload["limit"]; !ok {
		t.Error("expected 'limit' key in list POST body")
	} else if limitVal != float64(100) {
		t.Errorf("expected limit=100, got %v", limitVal)
	}
}

// TestDeviceTag_BeforeRequest_Passthrough verifies that unrelated operations are not modified.
func TestDeviceTag_BeforeRequest_Passthrough(t *testing.T) {
	hook := &deviceTagRequestHook{}
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/device/tags", strings.NewReader(`{"name":"x"}`))
	ctx := BeforeRequestContext{HookContext: HookContext{OperationID: "createDeviceTag"}}

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Method != http.MethodPost {
		t.Errorf("expected passthrough to keep method POST, got %s", result.Method)
	}
	if !strings.HasSuffix(result.URL.Path, "/device/tags") {
		t.Errorf("expected path unchanged, got %s", result.URL.Path)
	}
}

// --- AfterSuccess (deviceTagAfterSuccessHook) tests ---

func buildDeviceTagSuccessResponse(statusCode int, body string) (*http.Response, *http.Request) {
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/device/tags", nil)
	res := &http.Response{
		StatusCode: statusCode,
		Status:     fmt.Sprintf("%d OK", statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}
	return res, req
}

// TestDeviceTag_CreateUnwrapsEnvelope verifies that the create response envelope
// {success, data: {id, name, description}} is unwrapped to the flat item.
func TestDeviceTag_CreateUnwrapsEnvelope(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"success":true,"data":{"id":1,"name":"prod-tag","description":"production"}}`
	res, _ := buildDeviceTagSuccessResponse(201, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "createDeviceTag"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var item map[string]interface{}
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, raw)
	}
	if item["id"] != float64(1) {
		t.Errorf("expected id=1, got %v", item["id"])
	}
	if item["name"] != "prod-tag" {
		t.Errorf("expected name=prod-tag, got %v", item["name"])
	}
	if _, ok := item["success"]; ok {
		t.Error("expected 'success' envelope key to be removed")
	}
}

// TestDeviceTag_UpdateUnwrapsEnvelope verifies that the update response envelope is unwrapped.
func TestDeviceTag_UpdateUnwrapsEnvelope(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"success":true,"data":{"id":5,"name":"updated-tag","description":"new desc"}}`
	res, _ := buildDeviceTagSuccessResponse(200, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "updateDeviceTag"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var item map[string]interface{}
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, raw)
	}
	if item["name"] != "updated-tag" {
		t.Errorf("expected name=updated-tag, got %v", item["name"])
	}
}

// TestDeviceTag_GetUnwrapsSingleFromPaginated verifies that the gettags paginated response
// is unwrapped to the single tag item.
func TestDeviceTag_GetUnwrapsSingleFromPaginated(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"success":true,"data":{"data":[{"id":3,"name":"my-tag","description":"desc"}],"total_count":1,"offset":0,"limit":1}}`
	res, _ := buildDeviceTagSuccessResponse(200, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "getDeviceTag"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var item map[string]interface{}
	if err := json.Unmarshal(raw, &item); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, raw)
	}
	if item["id"] != float64(3) {
		t.Errorf("expected id=3, got %v", item["id"])
	}
	if item["name"] != "my-tag" {
		t.Errorf("expected name=my-tag, got %v", item["name"])
	}
}

// TestDeviceTag_GetNotFoundSynthesises404 verifies that when gettags returns empty results
// the hook synthesises a 404 response so Terraform removes the resource from state.
func TestDeviceTag_GetNotFoundSynthesises404(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"success":true,"data":{"data":[],"total_count":0,"offset":0,"limit":1}}`
	res, _ := buildDeviceTagSuccessResponse(200, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "getDeviceTag"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed unexpectedly: %v", err)
	}

	if result.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", result.StatusCode)
	}
}

// TestDeviceTag_ListUnwrapsToTagsWrapper verifies that the paginated list response
// is converted to {tags: [{...}]} for the SDK's list schema.
func TestDeviceTag_ListUnwrapsToTagsWrapper(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"success":true,"data":{"data":[{"id":1,"name":"tag-a"},{"id":2,"name":"tag-b"}],"total_count":2,"offset":0,"limit":100}}`
	res, _ := buildDeviceTagSuccessResponse(200, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "listDeviceTags"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var wrapped map[string]interface{}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		t.Fatalf("response is not valid JSON: %v\nbody: %s", err, raw)
	}
	tags, ok := wrapped["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected 'tags' array in response, got: %s", raw)
	}
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

// TestDeviceTag_AfterSuccess_Passthrough verifies that unrelated operations are not modified.
func TestDeviceTag_AfterSuccess_Passthrough(t *testing.T) {
	hook := &deviceTagAfterSuccessHook{}
	body := `{"something":"else"}`
	res, _ := buildDeviceTagSuccessResponse(200, body)
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "deleteDeviceTag"}}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	if string(raw) != body {
		t.Errorf("expected body unchanged for passthrough, got: %s", raw)
	}
}
