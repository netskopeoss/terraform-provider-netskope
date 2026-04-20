package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakeDCTagsArrayResponse(t *testing.T, tags []map[string]interface{}) *http.Response {
	t.Helper()
	b, err := json.Marshal(tags)
	if err != nil {
		t.Fatalf("failed to marshal test response: %v", err)
	}
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(string(b))),
		Header:     make(http.Header),
	}
}

func TestDeviceClassificationTags_WrapsArrayResponse(t *testing.T) {
	hook := &deviceClassificationTagsResponse{}
	ctx := hookCtxForOp("listDeviceClassificationTags")

	tags := []map[string]interface{}{
		{"id": 100, "name": "Managed", "description": "Managed devices", "priority": 1},
		{"id": 200, "name": "Unmanaged", "description": "Unmanaged devices", "priority": 2},
	}
	res := buildFakeDCTagsArrayResponse(t, tags)

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err != nil {
		t.Fatalf("result is not a JSON object: %v", err)
	}

	tagsRaw, ok := wrapped["tags"]
	if !ok {
		t.Fatal("result missing 'tags' key")
	}

	var resultTags []map[string]interface{}
	if err := json.Unmarshal(tagsRaw, &resultTags); err != nil {
		t.Fatalf("failed to unmarshal tags array: %v", err)
	}

	if len(resultTags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(resultTags))
	}
	if resultTags[0]["name"] != "Managed" {
		t.Errorf("expected first tag name 'Managed', got %v", resultTags[0]["name"])
	}
	if resultTags[1]["name"] != "Unmanaged" {
		t.Errorf("expected second tag name 'Unmanaged', got %v", resultTags[1]["name"])
	}
}

func TestDeviceClassificationTags_EmptyArray(t *testing.T) {
	hook := &deviceClassificationTagsResponse{}
	ctx := hookCtxForOp("listDeviceClassificationTags")

	res := buildFakeDCTagsArrayResponse(t, []map[string]interface{}{})

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}

	var wrapped map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapped); err != nil {
		t.Fatalf("result is not a JSON object: %v", err)
	}

	var resultTags []interface{}
	if err := json.Unmarshal(wrapped["tags"], &resultTags); err != nil {
		t.Fatalf("failed to unmarshal tags: %v", err)
	}
	if len(resultTags) != 0 {
		t.Errorf("expected 0 tags, got %d", len(resultTags))
	}
}

func TestDeviceClassificationTags_SkipsOtherOperations(t *testing.T) {
	hook := &deviceClassificationTagsResponse{}
	ctx := hookCtxForOp("listNPARules")

	// Pass a non-array body — should be returned unchanged since operationID doesn't match
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader(`{"data":[],"status":"success"}`)),
		Header:     make(http.Header),
	}

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}

	if string(body) != `{"data":[],"status":"success"}` {
		t.Errorf("expected body unchanged, got %s", string(body))
	}
}
