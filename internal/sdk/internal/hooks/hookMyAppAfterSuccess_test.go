package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// buildFakeAppJSONResponse constructs a fake *http.Response for single-app AfterSuccess hook tests.
func buildFakeAppJSONResponse(body string, operationID string) (AfterSuccessContext, *http.Response) {
	res := &http.Response{
		StatusCode: 200,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
	ctx := AfterSuccessContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, res
}

// TestAppAfterSuccess_PublisherIDFloat64ToString verifies that publisher_id delivered as
// a float64 JSON number (the standard Go json.Unmarshal representation for interface{})
// is converted to a string. This is the BUG-018 class risk: if the type switch breaks,
// publisher_id would remain a float64 and the SDK's string field would be empty on read.
func TestAppAfterSuccess_PublisherIDFloat64ToString(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 42,
			"app_name": "[my-app]",
			"service_publisher_assignments": [
				{"publisher_id": 7, "publisher_name": "pub-one"}
			],
			"protocols": [],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)

	var out struct {
		Data struct {
			Assignments []struct {
				PublisherID interface{} `json:"publisher_id"`
			} `json:"service_publisher_assignments"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(out.Data.Assignments) == 0 {
		t.Fatal("expected at least one publisher assignment")
	}
	id, ok := out.Data.Assignments[0].PublisherID.(string)
	if !ok {
		t.Fatalf("expected publisher_id to be string after hook, got %T: %v",
			out.Data.Assignments[0].PublisherID, out.Data.Assignments[0].PublisherID)
	}
	if id != "7" {
		t.Errorf("expected publisher_id = '7', got %q", id)
	}
}

// TestAppAfterSuccess_BracketStripping verifies that app_name has brackets stripped.
func TestAppAfterSuccess_BracketStripping(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[my-private-app]",
			"service_publisher_assignments": [],
			"protocols": [],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	if strings.Contains(string(raw), `"[my-private-app]"`) {
		t.Errorf("expected brackets stripped from app_name, got: %s", raw)
	}
	if !strings.Contains(string(raw), `"my-private-app"`) {
		t.Errorf("expected 'my-private-app' in result, got: %s", raw)
	}
}

// TestAppAfterSuccess_TransportCopiedToType verifies that the API inconsistency
// (GET returns "transport", schema expects "type") is corrected by the hook.
func TestAppAfterSuccess_TransportCopiedToType(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [],
			"protocols": [
				{"port": "443", "transport": "TCP", "type": ""}
			],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Protocols []struct {
				Type      string `json:"type"`
				Transport string `json:"transport"`
			} `json:"protocols"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if len(out.Data.Protocols) == 0 {
		t.Fatal("expected at least one protocol")
	}
	if out.Data.Protocols[0].Type != "TCP" {
		t.Errorf("expected type = 'TCP' (copied from transport), got %q", out.Data.Protocols[0].Type)
	}
}

// TestAppAfterSuccess_ProtocolsSortedByTypeAndPort verifies deterministic protocol ordering.
func TestAppAfterSuccess_ProtocolsSortedByTypeAndPort(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [],
			"protocols": [
				{"port": "8080", "transport": "TCP", "type": "TCP"},
				{"port": "443",  "transport": "TCP", "type": "TCP"},
				{"port": "53",   "transport": "UDP", "type": "UDP"}
			],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Protocols []struct {
				Type string `json:"type"`
				Port string `json:"port"`
			} `json:"protocols"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Expected order: TCP/443, TCP/8080, UDP/53
	want := []struct{ typ, port string }{
		{"TCP", "443"},
		{"TCP", "8080"},
		{"UDP", "53"},
	}
	for i, w := range want {
		if i >= len(out.Data.Protocols) {
			t.Fatalf("expected %d protocols, got %d", len(want), len(out.Data.Protocols))
		}
		p := out.Data.Protocols[i]
		if p.Type != w.typ || p.Port != w.port {
			t.Errorf("protocol[%d]: want %s/%s, got %s/%s", i, w.typ, w.port, p.Type, p.Port)
		}
	}
}

// TestAppAfterSuccess_PublishersSortedByID verifies publishers are sorted by ID numerically.
func TestAppAfterSuccess_PublishersSortedByID(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [
				{"publisher_id": 20, "publisher_name": "pub-b"},
				{"publisher_id": 5,  "publisher_name": "pub-a"},
				{"publisher_id": 100,"publisher_name": "pub-c"}
			],
			"protocols": [],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Assignments []struct {
				PublisherID interface{} `json:"publisher_id"`
			} `json:"service_publisher_assignments"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	wantOrder := []string{"5", "20", "100"}
	for i, want := range wantOrder {
		if i >= len(out.Data.Assignments) {
			t.Fatalf("expected %d assignments", len(wantOrder))
		}
		got := out.Data.Assignments[i].PublisherID.(string)
		if got != want {
			t.Errorf("assignment[%d]: want publisher_id %q, got %q", i, want, got)
		}
	}
}

// TestAppAfterSuccess_LabelIdsExtracted verifies that label_ids is populated from the
// labels array returned by the API.
func TestAppAfterSuccess_LabelIdsExtracted(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [],
			"protocols": [],
			"tags": [],
			"labels": [
				{"label_id": "label-b", "permission": "read"},
				{"label_id": "label-a", "permission": "write"}
			]
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			LabelIds []string `json:"label_ids"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	// Should be sorted alphabetically
	if len(out.Data.LabelIds) != 2 {
		t.Fatalf("expected 2 label_ids, got %d", len(out.Data.LabelIds))
	}
	if out.Data.LabelIds[0] != "label-a" || out.Data.LabelIds[1] != "label-b" {
		t.Errorf("expected sorted label_ids [label-a, label-b], got %v", out.Data.LabelIds)
	}
}

// TestAppAfterSuccess_PublisherNameWhitespaceTrimmed verifies leading/trailing
// whitespace is stripped from publisher_name (the API sometimes returns leading spaces).
func TestAppAfterSuccess_PublisherNameWhitespaceTrimmed(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [
				{"publisher_id": 1, "publisher_name": "  pub-with-spaces  "}
			],
			"protocols": [],
			"tags": []
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "getNPAPrivateApp")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Assignments []struct {
				PublisherName string `json:"publisher_name"`
			} `json:"service_publisher_assignments"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if out.Data.Assignments[0].PublisherName != "pub-with-spaces" {
		t.Errorf("expected trimmed publisher_name, got %q", out.Data.Assignments[0].PublisherName)
	}
}

// TestAppAfterSuccess_CreateAndUpdateOperationsHandled verifies that createNPAPrivateApps
// and updateNPAPrivateApp also go through the hook (not just getNPAPrivateApp).
func TestAppAfterSuccess_CreateAndUpdateOperationsHandled(t *testing.T) {
	hook := &myAppResponse{}

	body := `{
		"data": {
			"app_id": 1,
			"app_name": "[app]",
			"service_publisher_assignments": [],
			"protocols": [],
			"tags": []
		},
		"status": "success"
	}`

	for _, opID := range []string{"createNPAPrivateApps", "updateNPAPrivateApp"} {
		ctx, res := buildFakeAppJSONResponse(body, opID)
		result, err := hook.AfterSuccess(ctx, res)
		if err != nil {
			t.Errorf("AfterSuccess(%s) failed: %v", opID, err)
			continue
		}
		raw, _ := io.ReadAll(result.Body)
		if !strings.Contains(string(raw), `"app"`) {
			t.Errorf("AfterSuccess(%s): expected bracket-stripped name in result", opID)
		}
	}
}

// TestAppAfterSuccess_NonMatchingOperationPassthrough verifies non-app operations are
// passed through unchanged.
func TestAppAfterSuccess_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &myAppResponse{}

	body := `{"data": {"app_id": 1}, "status": "success"}`
	ctx, res := buildFakeAppJSONResponse(body, "deleteNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed for non-matching operation: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response for passthrough")
	}
}
