package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestNPAPrivateAppTagHook_BeforeRequest_wrapsTagName(t *testing.T) {
	hook := &npaPrivateAppTagHook{}
	ctx := BeforeRequestContext{HookContext: HookContext{OperationID: "createNPAPrivateAppTag"}}

	body := []byte(`{"tag_name":"production"}`)
	req, _ := http.NewRequest("POST", "https://tenant.goskope.com/api/v2/steering/apps/private/tags", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	out, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := io.ReadAll(out.Body)
	var wrapped map[string]interface{}
	if err := json.Unmarshal(got, &wrapped); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	tags, ok := wrapped["tags"].([]interface{})
	if !ok || len(tags) != 1 {
		t.Fatalf("expected wrapped[tags] to be a single-element array, got: %v", wrapped)
	}
	item, ok := tags[0].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tags[0] to be an object, got: %T", tags[0])
	}
	if item["tag_name"] != "production" {
		t.Errorf("expected tag_name=production, got: %v", item["tag_name"])
	}
}

func TestNPAPrivateAppTagHook_BeforeRequest_skipsOtherOperations(t *testing.T) {
	hook := &npaPrivateAppTagHook{}
	ctx := BeforeRequestContext{HookContext: HookContext{OperationID: "updateNPAPrivateAppTag"}}

	body := []byte(`{"tag_name":"production"}`)
	req, _ := http.NewRequest("PUT", "https://tenant.goskope.com/api/v2/steering/apps/private/tags/42", bytes.NewReader(body))

	out, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be the same request, unmodified
	got, _ := io.ReadAll(out.Body)
	if string(got) != string(body) {
		t.Errorf("expected body unchanged, got: %s", string(got))
	}
}

func TestNPAPrivateAppTagHook_AfterSuccess_unwrapsTagsArray(t *testing.T) {
	hook := &npaPrivateAppTagHook{}
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "createNPAPrivateAppTag"}}

	apiBody := []byte(`{"data":{"id":null,"tags":[{"tag_id":42,"tag_name":"production"}],"tld_domain":""},"status":"success"}`)
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(apiBody)),
	}

	out, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := io.ReadAll(out.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(got, &result); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected result[data] to be an object, got: %T", result["data"])
	}
	if int(data["tag_id"].(float64)) != 42 {
		t.Errorf("expected tag_id=42, got: %v", data["tag_id"])
	}
	if data["tag_name"] != "production" {
		t.Errorf("expected tag_name=production, got: %v", data["tag_name"])
	}
	if _, hasTld := data["tld_domain"]; hasTld {
		t.Error("tld_domain should not appear in normalized response")
	}
	if result["status"] != "success" {
		t.Errorf("expected status=success, got: %v", result["status"])
	}
}

func TestNPAPrivateAppTagHook_AfterSuccess_emptyTagsReturnsError(t *testing.T) {
	hook := &npaPrivateAppTagHook{}
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "createNPAPrivateAppTag"}}

	apiBody := []byte(`{"data":{"id":null,"tags":[]},"status":"success"}`)
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(apiBody)),
	}

	_, err := hook.AfterSuccess(ctx, res)
	if err == nil {
		t.Fatal("expected error for empty tags array, got nil")
	}
}

func TestNPAPrivateAppTagHook_AfterSuccess_skipsOtherOperations(t *testing.T) {
	hook := &npaPrivateAppTagHook{}
	ctx := AfterSuccessContext{HookContext: HookContext{OperationID: "getNPAPrivateAppTag"}}

	original := []byte(`{"data":{"tag_id":42,"tag_name":"production"},"status":"success"}`)
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(bytes.NewReader(original)),
	}

	out, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, _ := io.ReadAll(out.Body)
	if string(got) != string(original) {
		t.Errorf("expected body unchanged for non-create operations, got: %s", string(got))
	}
}
