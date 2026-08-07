package hooks

import (
	"io"
	"strings"
	"testing"
)

// TestBulkAfterSuccess_ClassificationArrayInList verifies that the bulk AfterSuccess
// hook handles getNPARulesList responses where rule_data.classification is an array
// of strings. Uses the same models.RuleData struct as the single-rule hook, so the
// same type fix (Classification []string) is exercised here.
// See docs/bugs/BUG-018.
func TestBulkAfterSuccess_ClassificationArrayInList(t *testing.T) {
	hook := &myBulkPolicyResponse{}

	body := `{
		"data": [
			{
				"rule_id": "10",
				"rule_name": "allow-rule",
				"enabled": "1",
				"rule_data": {
					"policy_type": "private-app",
					"privateApps": ["[app-one]"],
					"access_method": ["Client"]
				}
			},
			{
				"rule_id": "20",
				"rule_name": "block-rule",
				"enabled": "1",
				"rule_data": {
					"policy_type": "private-app",
					"classification": ["unmanaged"],
					"privateApps": ["[app-two]"],
					"access_method": ["Client"]
				}
			}
		],
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARulesList")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed with classification array in list: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestBulkAfterSuccess_BracketStripping verifies that the bulk AfterSuccess hook
// strips brackets from privateApps entries in list responses.
func TestBulkAfterSuccess_BracketStripping(t *testing.T) {
	hook := &myBulkPolicyResponse{}

	body := `{
		"data": [
			{
				"rule_id": "1",
				"rule_name": "test-rule",
				"rule_data": {
					"privateApps": ["[my-app]", "[other-app]"],
					"access_method": ["Client"]
				}
			}
		],
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARulesList")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	body2 := string(rawBody)

	if strings.Contains(body2, `"[my-app]"`) {
		t.Errorf("expected brackets to be stripped from 'my-app', got: %s", body2)
	}
	if strings.Contains(body2, `"[other-app]"`) {
		t.Errorf("expected brackets to be stripped from 'other-app', got: %s", body2)
	}
	if !strings.Contains(body2, `"my-app"`) {
		t.Errorf("expected 'my-app' in result, got: %s", body2)
	}
}

// TestBulkAfterSuccess_ClassificationAbsent verifies the bulk hook handles list
// responses where rules have no classification field at all.
func TestBulkAfterSuccess_ClassificationAbsent(t *testing.T) {
	hook := &myBulkPolicyResponse{}

	body := `{
		"data": [
			{
				"rule_id": "1",
				"rule_name": "allow-rule",
				"rule_data": {
					"policy_type": "private-app",
					"privateApps": ["[my-app]"],
					"access_method": ["Client"]
				}
			}
		],
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARulesList")

	_, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed when classification absent: %v", err)
	}
}

// TestBulkAfterSuccess_EmptyList verifies the bulk hook handles an empty data array
// without panicking.
func TestBulkAfterSuccess_EmptyList(t *testing.T) {
	hook := &myBulkPolicyResponse{}

	body := `{"data": [], "status": "success"}`

	ctx, res := buildFakePolicyResponse(body, "getNPARulesList")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed on empty list: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestBulkAfterSuccess_NonMatchingOperation verifies that the bulk hook passes
// through responses for operations other than getNPARulesList unchanged.
func TestBulkAfterSuccess_NonMatchingOperation(t *testing.T) {
	hook := &myBulkPolicyResponse{}

	body := `{"data": [{"rule_id": "1"}], "status": "success"}`
	ctx, res := buildFakePolicyResponse(body, "deleteNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed for non-matching operation: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	if string(rawBody) != body {
		t.Errorf("expected body unchanged for non-matching operation")
	}
}
