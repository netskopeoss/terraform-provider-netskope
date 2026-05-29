package hooks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUrllistHandleList(t *testing.T) {
	hook := &urllistAfterSuccess{}

	// Simulate the raw array response from GET /policy/urllist
	rawArray := `[{"id":1,"name":"test-list","data":{"urls":["www.example.com"],"type":"exact"},"pending":0}]`

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/policy/urllist", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(rawArray)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "listUrllists",
		},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(result.Body)

	var wrapped struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		t.Fatalf("failed to unmarshal wrapped response: %v", err)
	}

	if len(wrapped.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(wrapped.Items))
	}
}

func TestUrllistHandleListEmpty(t *testing.T) {
	hook := &urllistAfterSuccess{}

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/policy/urllist", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString("[]")),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "listUrllists",
		},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(result.Body)

	var wrapped struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(body, &wrapped); err != nil {
		t.Fatalf("failed to unmarshal wrapped response: %v", err)
	}

	if len(wrapped.Items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(wrapped.Items))
	}
}

func TestUrllistHandleCreate(t *testing.T) {
	// Set up a mock server that handles deploy + GET
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/policy/urllist/deploy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/api/v2/policy/urllist/42", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":42,"name":"test-list","data":{"urls":["www.example.com"],"type":"exact","json_version":2},"modify_type":"Created","modify_by":"test","modify_time":"2026-01-01T00:00:00.000Z","pending":0}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	hook := &urllistAfterSuccess{}

	// Simulate create response (array with single item)
	createResp := `[{"id":42,"name":"test-list","data":{"urls":["www.example.com"],"type":"exact","json_version":2},"modify_type":"Created","modify_by":"test","modify_time":"2026-01-01T00:00:00.000Z","pending":1}]`

	req, _ := http.NewRequest("POST", server.URL+"/api/v2/policy/urllist", nil)
	req.Header.Set("Netskope-Api-Token", "test-token")
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(createResp)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "createUrllist",
		},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(result.Body)

	var item struct {
		ID      int64  `json:"id"`
		Name    string `json:"name"`
		Pending int64  `json:"pending"`
		Data    struct {
			JSONVersion *int64 `json:"json_version"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if item.ID != 42 {
		t.Errorf("expected ID 42, got %d", item.ID)
	}
	if item.Name != "test-list" {
		t.Errorf("expected name 'test-list', got %q", item.Name)
	}
	if item.Pending != 0 {
		t.Errorf("expected pending 0 (deployed), got %d", item.Pending)
	}
	if item.Data.JSONVersion != nil {
		t.Errorf("expected json_version to be stripped, but it was present")
	}
}

func TestUrllistHandleUpdate(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/policy/urllist/deploy", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	})
	mux.HandleFunc("/api/v2/policy/urllist/7", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":7,"name":"updated","data":{"urls":["www.new.com"],"type":"regex"},"pending":0}`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	hook := &urllistAfterSuccess{}

	updateResp := `{"id":7,"name":"updated","data":{"urls":["www.new.com"],"type":"regex","json_version":2},"pending":1}`

	req, _ := http.NewRequest("PUT", server.URL+"/api/v2/policy/urllist/7", nil)
	req.Header.Set("Netskope-Api-Token", "test-token")
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(updateResp)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "updateUrllist",
		},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body, _ := io.ReadAll(result.Body)

	var item struct {
		ID      int64 `json:"id"`
		Pending int64 `json:"pending"`
	}
	if err := json.Unmarshal(body, &item); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if item.ID != 7 {
		t.Errorf("expected ID 7, got %d", item.ID)
	}
	if item.Pending != 0 {
		t.Errorf("expected pending 0, got %d", item.Pending)
	}
}

func TestUrllistHandleDelete(t *testing.T) {
	deployed := false
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/policy/urllist/deploy", func(w http.ResponseWriter, r *http.Request) {
		deployed = true
		w.WriteHeader(200)
		w.Write([]byte(`[]`))
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	hook := &urllistAfterSuccess{}

	deleteResp := `{"id":3,"name":"to-delete","modify_type":"Deleted","pending":1}`

	req, _ := http.NewRequest("DELETE", server.URL+"/api/v2/policy/urllist/3", nil)
	req.Header.Set("Netskope-Api-Token", "test-token")
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(deleteResp)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "deleteUrllist",
		},
	}

	_, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deployed {
		t.Error("expected deploy to be called")
	}
}

func TestUrllistNoopForUnrelatedOps(t *testing.T) {
	hook := &urllistAfterSuccess{}

	req, _ := http.NewRequest("GET", "https://example.com/api/v2/something", nil)
	res := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewBufferString(`{"foo":"bar"}`)),
		Header:     make(http.Header),
		Request:    req,
	}

	hookCtx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: "someOtherOperation",
		},
	}

	result, err := hook.AfterSuccess(hookCtx, res)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should pass through unchanged
	if result != res {
		t.Error("expected response to pass through unchanged")
	}
}
