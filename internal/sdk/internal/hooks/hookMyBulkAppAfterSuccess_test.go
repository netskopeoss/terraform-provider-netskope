package hooks

import (
	"encoding/json"
	"io"
	"strings"
	"testing"
)

// TestBulkAppAfterSuccess_PublisherIDFloat64ToString verifies that publisher_id delivered
// as a float64 JSON number is converted to a string in list responses.
// This is the BUG-018 class risk in the bulk app hook.
func TestBulkAppAfterSuccess_PublisherIDFloat64ToString(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{
		"data": {
			"private_apps": [
				{
					"app_id": 1,
					"app_name": "[app]",
					"service_publisher_assignments": [
						{"publisher_id": 7, "publisher_name": "pub-one"}
					],
					"protocols": [],
					"tags": []
				}
			]
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "listNPAPrivateApps")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Apps []struct {
				Assignments []struct {
					PublisherID interface{} `json:"publisher_id"`
				} `json:"service_publisher_assignments"`
			} `json:"private_apps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	if len(out.Data.Apps) == 0 || len(out.Data.Apps[0].Assignments) == 0 {
		t.Fatal("expected at least one app with one publisher assignment")
	}
	id, ok := out.Data.Apps[0].Assignments[0].PublisherID.(string)
	if !ok {
		t.Fatalf("expected publisher_id to be string after hook, got %T: %v",
			out.Data.Apps[0].Assignments[0].PublisherID,
			out.Data.Apps[0].Assignments[0].PublisherID)
	}
	if id != "7" {
		t.Errorf("expected publisher_id = '7', got %q", id)
	}
}

// TestBulkAppAfterSuccess_BracketStripping verifies app_name has brackets stripped
// across all apps in the list.
func TestBulkAppAfterSuccess_BracketStripping(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{
		"data": {
			"private_apps": [
				{"app_id": 1, "app_name": "[app-one]", "service_publisher_assignments": [], "protocols": [], "tags": []},
				{"app_id": 2, "app_name": "[app-two]", "service_publisher_assignments": [], "protocols": [], "tags": []}
			]
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "listNPAPrivateApps")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	if strings.Contains(string(raw), `"[app-one]"`) || strings.Contains(string(raw), `"[app-two]"`) {
		t.Errorf("expected brackets stripped from all app names, got: %s", raw)
	}
	if !strings.Contains(string(raw), `"app-one"`) || !strings.Contains(string(raw), `"app-two"`) {
		t.Errorf("expected clean app names in result, got: %s", raw)
	}
}

// TestBulkAppAfterSuccess_AppIDCopiedToID verifies BUG-012 fix: app_id is copied to id
// so the data source can read private_app_id.
func TestBulkAppAfterSuccess_AppIDCopiedToID(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{
		"data": {
			"private_apps": [
				{"app_id": 42, "id": 0, "app_name": "[app]", "service_publisher_assignments": [], "protocols": [], "tags": []}
			]
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "listNPAPrivateApps")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Apps []struct {
				ID    int `json:"id"`
				AppID int `json:"app_id"`
			} `json:"private_apps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if out.Data.Apps[0].ID != 42 {
		t.Errorf("expected id = 42 (copied from app_id), got %d", out.Data.Apps[0].ID)
	}
}

// TestBulkAppAfterSuccess_TransportCopiedToType verifies the transport→type copy
// is applied to each app's protocols in the list.
func TestBulkAppAfterSuccess_TransportCopiedToType(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{
		"data": {
			"private_apps": [
				{
					"app_id": 1,
					"app_name": "[app]",
					"service_publisher_assignments": [],
					"protocols": [{"port": "443", "transport": "TCP", "type": ""}],
					"tags": []
				}
			]
		},
		"status": "success"
	}`

	ctx, res := buildFakeAppJSONResponse(body, "listNPAPrivateApps")
	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out struct {
		Data struct {
			Apps []struct {
				Protocols []struct {
					Type string `json:"type"`
				} `json:"protocols"`
			} `json:"private_apps"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if out.Data.Apps[0].Protocols[0].Type != "TCP" {
		t.Errorf("expected type = 'TCP' copied from transport, got %q", out.Data.Apps[0].Protocols[0].Type)
	}
}

// TestBulkAppAfterSuccess_EmptyList verifies that an empty apps list doesn't crash.
func TestBulkAppAfterSuccess_EmptyList(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{"data": {"private_apps": []}, "status": "success"}`
	ctx, res := buildFakeAppJSONResponse(body, "listNPAPrivateApps")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed on empty list: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestBulkAppAfterSuccess_NonMatchingOperationPassthrough verifies non-list operations
// are passed through unchanged.
func TestBulkAppAfterSuccess_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &myBulkAppResponse{}

	body := `{"data": {"private_apps": []}, "status": "success"}`
	ctx, res := buildFakeAppJSONResponse(body, "deleteNPAPrivateApp")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed for non-matching operation: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response for passthrough")
	}
}
