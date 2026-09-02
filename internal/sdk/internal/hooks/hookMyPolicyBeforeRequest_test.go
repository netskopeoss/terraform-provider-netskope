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

// TestBeforeRequest_DeviceClassificationIDCoercion verifies that
// device_classification_id string values are coerced to integers in the
// marshalled output. The API returns strings (e.g. ["5871"]) but expects
// integers on write (e.g. [5871]).
func TestBeforeRequest_DeviceClassificationIDCoercion(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"device_classification_id": ["5871", "42"]
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	// Parse as raw JSON to verify device_classification_id contains integers
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(resultBody), &raw); err != nil {
		t.Fatalf("failed to unmarshal result body: %v", err)
	}
	var ruleData map[string]json.RawMessage
	if err := json.Unmarshal(raw["rule_data"], &ruleData); err != nil {
		t.Fatalf("failed to unmarshal rule_data: %v", err)
	}

	var ids []json.Number
	if err := json.Unmarshal(ruleData["device_classification_id"], &ids); err != nil {
		t.Fatalf("failed to unmarshal device_classification_id: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("expected 2 device_classification_id values, got %d", len(ids))
	}

	// Verify they are valid integers (not quoted strings)
	expected := []int64{5871, 42}
	for i, id := range ids {
		n, err := id.Int64()
		if err != nil {
			t.Errorf("device_classification_id[%d] = %s is not an integer: %v", i, id, err)
		}
		if n != expected[i] {
			t.Errorf("device_classification_id[%d] = %d, want %d", i, n, expected[i])
		}
	}

	// Also verify the raw JSON contains unquoted integers, not strings
	rawDCI := string(ruleData["device_classification_id"])
	if strings.Contains(rawDCI, `"5871"`) {
		t.Errorf("device_classification_id should contain unquoted integers, got %s", rawDCI)
	}
}

// TestBeforeRequest_DeviceClassificationIDEmpty verifies that an empty
// device_classification_id is omitted from the output (omitempty).
func TestBeforeRequest_DeviceClassificationIDEmpty(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
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

	if strings.Contains(resultBody, "device_classification_id") {
		t.Errorf("expected device_classification_id to be omitted when empty, got %s", resultBody)
	}
}

// TestBeforeRequest_DeviceClassificationIDInvalidValue verifies that a
// non-numeric device_classification_id value causes the hook to return an error.
func TestBeforeRequest_DeviceClassificationIDInvalidValue(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"device_classification_id": ["not-a-number"]
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	_, err := hook.BeforeRequest(ctx, req)
	if err == nil {
		t.Fatal("expected error for non-numeric device_classification_id, got nil")
	}
	if !strings.Contains(err.Error(), "not-a-number") {
		t.Errorf("expected error to mention the bad value, got: %v", err)
	}
}

// TestBeforeRequest_NegateNetLocationStrippedWhenEmpty verifies that
// b_negateNetLocation is removed from the request when net_location_obj is
// empty or absent. Tenants with the source-IP-criteria feature flag disabled
// reject b_negateNetLocation even when set to false.
func TestBeforeRequest_NegateNetLocationStrippedWhenEmpty(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"b_negateNetLocation": false,
			"b_negateSrcCountries": false
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(resultBody), &out); err != nil {
		t.Fatalf("failed to parse result body: %v", err)
	}
	ruleData, ok := out["rule_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("rule_data missing from result body")
	}
	if _, found := ruleData["b_negateNetLocation"]; found {
		t.Errorf("expected b_negateNetLocation to be stripped when net_location_obj is empty, got %s", resultBody)
	}
	if _, found := ruleData["b_negateSrcCountries"]; found {
		t.Errorf("expected b_negateSrcCountries to be stripped when srcCountries is empty, got %s", resultBody)
	}
}

// TestBeforeRequest_NegateNetLocationPreservedWhenPopulated verifies that
// b_negateNetLocation is kept in the request when net_location_obj has entries.
func TestBeforeRequest_NegateNetLocationPreservedWhenPopulated(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "test-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {"action_name": "allow"},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"net_location_obj": ["192.168.0.0/16"],
			"b_negateNetLocation": true
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(resultBody), &out); err != nil {
		t.Fatalf("failed to parse result body: %v", err)
	}
	ruleData, ok := out["rule_data"].(map[string]interface{})
	if !ok {
		t.Fatalf("rule_data missing from result body")
	}
	if v, found := ruleData["b_negateNetLocation"]; !found {
		t.Errorf("expected b_negateNetLocation to be present when net_location_obj is non-empty, got %s", resultBody)
	} else if v != true {
		t.Errorf("expected b_negateNetLocation=true, got %v", v)
	}
}

// TestBeforeRequest_HtmlTemplateStrippedOnUpdate verifies that the BeforeRequest
// hook removes the template field from update payloads when it contains a .html
// file name. suppressTemplateDrift sets the planned template to the API-returned
// file name to suppress phantom diffs; stripping it here prevents the PUT from
// sending a file name the API rejects with "Undefined template: *.html".
// See docs/bugs/BUG-019-block-rule-template-phantom-update.md
func TestBeforeRequest_HtmlTemplateStrippedOnUpdate(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "block-rule",
		"enabled": "0",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "block",
				"emit_alert": true,
				"template": "23.html"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"]
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "updateNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	if strings.Contains(resultBody, `"template"`) {
		t.Errorf("expected .html template to be stripped from update payload, got: %s", resultBody)
	}
	if strings.Contains(resultBody, "23.html") {
		t.Errorf("expected file name '23.html' to be absent from update payload, got: %s", resultBody)
	}
}

// TestBeforeRequest_HtmlTemplatePreservedForPeriodicReauthOnUpdate verifies that
// the .html template filename is NOT stripped from periodic_reauth update payloads.
// Only block rule update payloads have .html filenames stripped (to prevent the API
// rejecting them). For periodic_reauth, the display name is sent as-is on update and
// the API stores it verbatim — the template applied from CREATE persists.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestBeforeRequest_HtmlTemplatePreservedForPeriodicReauthOnUpdate(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "reauth-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "periodic_reauth",
				"template": "10.html"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"os": ["Windows"],
			"periodic_reauth": {
				"reauth_interval": "8",
				"reauth_interval_unit": "hours"
			}
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "updateNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	// Template must be preserved — the .html strip only applies to block rules.
	if !strings.Contains(resultBody, `"template"`) {
		t.Errorf("expected template to be preserved in periodic_reauth update payload, got: %s", resultBody)
	}
	if !strings.Contains(resultBody, "10.html") {
		t.Errorf("expected '10.html' to be present in periodic_reauth update payload, got: %s", resultBody)
	}
}

// TestBeforeRequest_DisplayNamePreservedOnUpdate verifies that display name
// template values (no .html suffix) are NOT stripped from update payloads.
// A user intentionally changing their template should have the new display name sent.
func TestBeforeRequest_DisplayNamePreservedOnUpdate(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "block-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "block",
				"emit_alert": true,
				"template": "Generic Block"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"]
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "updateNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	if !strings.Contains(resultBody, `"Generic Block"`) {
		t.Errorf("expected display name 'Generic Block' to be preserved in update payload, got: %s", resultBody)
	}
}

// TestBeforeRequest_DisplayNamePreservedOnCreate verifies that display name
// template values are NOT stripped from create payloads. The API accepts display
// names on create and rejects file names — create payloads must never be stripped.
func TestBeforeRequest_DisplayNamePreservedOnCreate(t *testing.T) {
	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "block-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "block",
				"emit_alert": true,
				"template": "Default Template"
			},
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

	if !strings.Contains(resultBody, `"Default Template"`) {
		t.Errorf("expected 'Default Template' to be preserved in create payload, got: %s", resultBody)
	}
}

// TestBeforeRequest_PeriodicReauthDisplayNamePreservedOnCreate verifies that a
// display name in a periodic_reauth template field is NOT translated to a filename
// in BeforeRequest. The periodic_reauth API (like block) accepts display names
// natively and translates them server-side — sending a filename causes the API to
// reject with "Undefined template". The display name must reach the API as-is.
// AfterSuccess translates the returned .html filename back to the display name.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestBeforeRequest_PeriodicReauthDisplayNamePreservedOnCreate(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	npaTemplatesAPISeedForTest([]npaTemplatesAPIEntry{
		{FileName: "10.html", Name: "My Reauth Template", ActionType: "periodic_reauth"},
	})

	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "reauth-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "periodic_reauth",
				"template": "My Reauth Template"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"os": ["Windows"],
			"periodic_reauth": {
				"reauth_interval": "8",
				"reauth_interval_unit": "hours"
			}
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	// Display name must be preserved — BeforeRequest must not translate to filename.
	// The API accepts display names for periodic_reauth just like block rules.
	if !strings.Contains(resultBody, `"My Reauth Template"`) {
		t.Errorf("expected display name 'My Reauth Template' to be preserved in create payload, got: %s", resultBody)
	}
	if strings.Contains(resultBody, "10.html") {
		t.Errorf("expected filename NOT to appear — BeforeRequest must not translate for periodic_reauth, got: %s", resultBody)
	}
}

// TestBeforeRequest_PeriodicReauthDisplayNamePreservedWhenTemplatesAPIEmpty
// verifies that display names are preserved in create payloads regardless of
// templates API cache state. BeforeRequest never translates for periodic_reauth.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestBeforeRequest_PeriodicReauthDisplayNamePreservedWhenTemplatesAPIEmpty(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "reauth-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "periodic_reauth",
				"template": "Unknown Template"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"os": ["Windows"],
			"periodic_reauth": {
				"reauth_interval": "8",
				"reauth_interval_unit": "hours"
			}
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "createNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	// Display name must be preserved — BeforeRequest never translates for periodic_reauth.
	if !strings.Contains(resultBody, `"Unknown Template"`) {
		t.Errorf("expected display name to be preserved in create payload, got: %s", resultBody)
	}
}

// TestBeforeRequest_PeriodicReauthDisplayNamePreservedOnUpdate verifies that the
// display name template value is NOT stripped or translated on periodic_reauth
// update payloads. Translation (display name → filename) is only done on CREATE.
// On UPDATE the display name is sent as-is; the API stores it verbatim, and
// subsequent GET responses return it unchanged — no drift.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestBeforeRequest_PeriodicReauthDisplayNamePreservedOnUpdate(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	npaTemplatesAPISeedForTest([]npaTemplatesAPIEntry{
		{FileName: "10.html", Name: "My Template", ActionType: "periodic_reauth"},
	})

	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "reauth-rule",
		"enabled": "0",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "periodic_reauth",
				"template": "My Template"
			},
			"privateApps": ["my-app"],
			"access_method": ["Client"],
			"os": ["Windows"],
			"periodic_reauth": {
				"reauth_interval": "8",
				"reauth_interval_unit": "hours"
			}
		}
	}`

	ctx, req := buildFakePolicyRequest(t, body, "updateNPARules")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	resultBody := readRequestBody(t, result)

	// Display name must be preserved — no stripping or translation on update for non-block.
	if !strings.Contains(resultBody, `"template"`) {
		t.Errorf("expected template to be present in periodic_reauth update payload, got: %s", resultBody)
	}
	if !strings.Contains(resultBody, `"My Template"`) {
		t.Errorf("expected display name 'My Template' to be preserved in update payload, got: %s", resultBody)
	}
	// Must NOT be translated to filename on update (translation only happens on create)
	if strings.Contains(resultBody, "10.html") {
		t.Errorf("expected filename NOT to appear on update (no translation), got: %s", resultBody)
	}
}

// TestBeforeRequest_BlockRuleDisplayNameNotTranslatedViaTemplatesAPI verifies
// that block rule display names are NOT translated to filenames even when the
// templates API cache has an entry. Block rules accept display names natively
// server-side, so translation is unnecessary and would cause the .html strip to
// unnecessarily suppress the value.
func TestBeforeRequest_BlockRuleDisplayNameNotTranslatedViaTemplatesAPI(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	npaTemplatesAPISeedForTest([]npaTemplatesAPIEntry{
		{FileName: "5.html", Name: "Generic Block", ActionType: "block"},
	})

	hook := &myPolicyRequest{}

	body := `{
		"rule_name": "block-rule",
		"enabled": "1",
		"group_id": "5",
		"rule_data": {
			"match_criteria_action": {
				"action_name": "block",
				"emit_alert": true,
				"template": "Generic Block"
			},
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

	// Block rule display name must NOT be translated — API accepts it as-is
	if !strings.Contains(resultBody, `"Generic Block"`) {
		t.Errorf("expected block rule display name to be preserved, got: %s", resultBody)
	}
	if strings.Contains(resultBody, "5.html") {
		t.Errorf("expected .html filename NOT to appear for block rule create, got: %s", resultBody)
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