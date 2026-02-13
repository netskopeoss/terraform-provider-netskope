package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

// buildFakePolicyRequest creates a fake HTTP request with the given JSON body
// and operation ID, suitable for testing the BeforeRequest hook.
func buildFakePolicyRequest(t *testing.T, body string, operationID string) (BeforeRequestContext, *http.Request) {
	t.Helper()
	req, err := http.NewRequest("POST", "https://example.com/api/v2/steering/apps/private/rules", io.NopCloser(strings.NewReader(body)))
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.ContentLength = int64(len(body))
	ctx := BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, req
}

// readRequestBody reads and returns the body of an HTTP request as a string.
func readRequestBody(t *testing.T, req *http.Request) string {
	t.Helper()
	body, err := io.ReadAll(req.Body)
	if err != nil {
		t.Fatalf("failed to read request body: %v", err)
	}
	return string(body)
}

// TestBeforeRequest_RuleOrderWithNumericRuleId verifies that the BeforeRequest
// hook can unmarshal a request body where rule_order.rule_id is a JSON number.
// This is the exact scenario from BUG-003: the SDK serializes rule_id as int64,
// but the hook's RuleOrder struct previously declared it as *string, causing
// json.Unmarshal to fail with "cannot unmarshal number into string".
func TestBeforeRequest_RuleOrderWithNumericRuleId(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"policy_type": "private-app",
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"]
		},
		"rule_order": {
			"order": "after",
			"rule_id": 4
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	// Verify the request body was re-marshaled successfully
	resultBody := readRequestBody(t, result)

	// Verify private apps were wrapped in brackets
	var parsed myPolicyRequest
	if err := json.Unmarshal([]byte(resultBody), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result body: %v", err)
	}

	if len(parsed.RuleData.PrivateApps) != 1 {
		t.Fatalf("expected 1 private app, got %d", len(parsed.RuleData.PrivateApps))
	}
	if parsed.RuleData.PrivateApps[0] != "[my-app]" {
		t.Errorf("expected private app '[my-app]', got %q", parsed.RuleData.PrivateApps[0])
	}

	// Verify rule_order was preserved through the round-trip
	if parsed.RuleOrder == nil {
		t.Fatal("expected rule_order to be preserved, got nil")
	}
	if parsed.RuleOrder.RuleID == nil {
		t.Fatal("expected rule_order.rule_id to be preserved, got nil")
	}
	if *parsed.RuleOrder.RuleID != 4 {
		t.Errorf("expected rule_order.rule_id=4, got %d", *parsed.RuleOrder.RuleID)
	}
	if parsed.RuleOrder.Order == nil || *parsed.RuleOrder.Order != "after" {
		t.Errorf("expected rule_order.order='after', got %v", parsed.RuleOrder.Order)
	}
}

// TestBeforeRequest_RuleOrderWithoutRuleId verifies that the hook works when
// rule_order has no rule_id (e.g., order = "top"). This is the happy path
// that already worked before BUG-003.
func TestBeforeRequest_RuleOrderWithoutRuleId(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"policy_type": "private-app",
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"]
		},
		"rule_order": {
			"order": "top"
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	var parsed myPolicyRequest
	if err := json.Unmarshal([]byte(resultBody), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result body: %v", err)
	}

	if parsed.RuleOrder == nil {
		t.Fatal("expected rule_order to be preserved")
	}
	if parsed.RuleOrder.Order == nil || *parsed.RuleOrder.Order != "top" {
		t.Errorf("expected rule_order.order='top', got %v", parsed.RuleOrder.Order)
	}
	if parsed.RuleOrder.RuleID != nil {
		t.Errorf("expected rule_order.rule_id to be nil, got %d", *parsed.RuleOrder.RuleID)
	}
}

// TestBeforeRequest_NoRuleOrder verifies that the hook works when rule_order
// is omitted entirely from the request body.
func TestBeforeRequest_NoRuleOrder(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"policy_type": "private-app",
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"]
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	var parsed myPolicyRequest
	if err := json.Unmarshal([]byte(resultBody), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result body: %v", err)
	}

	if parsed.RuleOrder != nil {
		t.Errorf("expected rule_order to remain nil, got %v", parsed.RuleOrder)
	}
}

// TestBeforeRequest_UpdateOperationWithRuleOrder verifies the hook also works
// for updateNPARules (not just createNPARules).
func TestBeforeRequest_UpdateOperationWithRuleOrder(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"policy_type": "private-app",
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["app-one", "app-two"],
			"access_method": ["Client"]
		},
		"rule_order": {
			"order": "after",
			"rule_id": 42
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "updateNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	var parsed myPolicyRequest
	if err := json.Unmarshal([]byte(resultBody), &parsed); err != nil {
		t.Fatalf("failed to unmarshal result body: %v", err)
	}

	// Verify both private apps were wrapped in brackets
	if len(parsed.RuleData.PrivateApps) != 2 {
		t.Fatalf("expected 2 private apps, got %d", len(parsed.RuleData.PrivateApps))
	}
	if parsed.RuleData.PrivateApps[0] != "[app-one]" {
		t.Errorf("expected '[app-one]', got %q", parsed.RuleData.PrivateApps[0])
	}
	if parsed.RuleData.PrivateApps[1] != "[app-two]" {
		t.Errorf("expected '[app-two]', got %q", parsed.RuleData.PrivateApps[1])
	}

	// Verify rule_order preserved
	if *parsed.RuleOrder.RuleID != 42 {
		t.Errorf("expected rule_order.rule_id=42, got %d", *parsed.RuleOrder.RuleID)
	}
}

// TestBeforeRequest_NonMatchingOperationPassthrough verifies that operations
// other than createNPARules/updateNPARules pass through without modification.
func TestBeforeRequest_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{"rule_name": "unchanged", "rule_order": {"order": "after", "rule_id": 99}}`

	ctx, req := buildFakePolicyRequest(t, body, "deleteNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)
	if resultBody != body {
		t.Errorf("expected body unchanged for non-matching operation, got %s", resultBody)
	}
}