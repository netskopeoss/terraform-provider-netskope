package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestCCIDataAfterSuccess_flattensCategories(t *testing.T) {
	hook := &cciDataAfterSuccess{}

	body := `{
		"data": {
			"category": [
				{"category_id": 600, "category_name": "Generative AI"},
				{"category_id": 5001, "category_name": "Uncategorized"}
			]
		},
		"status": "Success",
		"status_code": 200
	}`

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/services/cci/data", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{OperationID: "getCCICategoryList"},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, _ := io.ReadAll(result.Body)

	var flat struct {
		Categories []struct {
			CategoryID   int    `json:"category_id"`
			CategoryName string `json:"category_name"`
		} `json:"categories"`
	}
	if err := json.Unmarshal(out, &flat); err != nil {
		t.Fatalf("failed to unmarshal rewritten body: %v", err)
	}

	if len(flat.Categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(flat.Categories))
	}
	if flat.Categories[0].CategoryID != 600 {
		t.Errorf("expected category_id 600, got %d", flat.Categories[0].CategoryID)
	}
	if flat.Categories[0].CategoryName != "Generative AI" {
		t.Errorf("expected category_name 'Generative AI', got %q", flat.Categories[0].CategoryName)
	}
	if flat.Categories[1].CategoryID != 5001 {
		t.Errorf("expected category_id 5001, got %d", flat.Categories[1].CategoryID)
	}
}

func TestCCIDataAfterSuccess_noCategoryField(t *testing.T) {
	hook := &cciDataAfterSuccess{}

	// data exists but has no "category" key
	body := `{"data": {}, "status": "Success", "status_code": 200}`

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/services/cci/data", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{OperationID: "getCCICategoryList"},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out, _ := io.ReadAll(result.Body)

	var flat struct {
		Categories []json.RawMessage `json:"categories"`
	}
	if err := json.Unmarshal(out, &flat); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(flat.Categories) != 0 {
		t.Errorf("expected empty categories, got %d", len(flat.Categories))
	}
}

func TestCCIDataAfterSuccess_noopForUnrelatedOps(t *testing.T) {
	hook := &cciDataAfterSuccess{}

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/something", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{OperationID: "someOtherOperation"},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != res {
		t.Error("expected response to pass through unchanged for unrelated operation")
	}
}

func TestCCIDataAfterSuccess_invalidJSON(t *testing.T) {
	hook := &cciDataAfterSuccess{}

	body := `not valid json`

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/services/cci/data", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{OperationID: "getCCICategoryList"},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should pass through the original body unchanged
	out, _ := io.ReadAll(result.Body)
	if string(out) != body {
		t.Errorf("expected original body to pass through, got %q", string(out))
	}
}
