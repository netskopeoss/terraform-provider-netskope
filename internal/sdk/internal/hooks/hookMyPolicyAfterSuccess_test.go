package hooks

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// buildFakePolicyResponse constructs a fake *http.Response with the given JSON body
// and operation ID, suitable for testing the AfterSuccess hook.
func buildFakePolicyResponse(body string, operationID string) (AfterSuccessContext, *http.Response) {
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

// TestAfterSuccess_ClassificationArrayUnmarshal verifies that the AfterSuccess hook
// correctly handles getNPARules responses where rule_data.classification is an array
// of strings. Previously the hook's RuleData struct had Classification *string which
// caused json.Unmarshal to fail with "cannot unmarshal array into Go struct field
// RuleData.data.rule_data.classification of type string".
// See docs/bugs/BUG-018.
func TestAfterSuccess_ClassificationArrayUnmarshal(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "224",
			"rule_name": "test-block-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"classification": ["unmanaged"],
				"privateApps": ["[my-app]"],
				"access_method": ["Client"]
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed with classification array: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response, got nil")
	}

	// Read the modified response body
	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}

	// The private app should have brackets stripped
	if strings.Contains(string(rawBody), `"[my-app]"`) {
		t.Errorf("expected private app brackets to be stripped, got: %s", string(rawBody))
	}
	if !strings.Contains(string(rawBody), `"my-app"`) {
		t.Errorf("expected 'my-app' in result body, got: %s", string(rawBody))
	}
}

// TestAfterSuccess_ClassificationAbsent verifies the hook processes rules that
// do not have the classification field at all (the common case).
func TestAfterSuccess_ClassificationAbsent(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "100",
			"rule_name": "allow-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"privateApps": ["[my-app]"],
				"access_method": ["Client"]
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed when classification is absent: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response, got nil")
	}
}

// TestAfterSuccess_NonMatchingOperationPassthrough verifies that the hook
// passes through responses for operations other than getNPARules/createNPARules.
func TestAfterSuccess_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{"data": {"rule_id": "1"}, "status": "success"}`
	ctx, res := buildFakePolicyResponse(body, "deleteNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed for non-matching operation: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response")
	}
}

// TestAfterSuccess_TemplateFilenameNotRewritten_BlockRule verifies that AfterSuccess
// no longer rewrites .html template filenames for block rules. Template reconciliation
// is now handled by private state in the plan modifier (suppressTemplateDrift).
// State will hold the .html filename returned by the API; the plan modifier detects
// whether this represents the same template as the configured display name.
// See docs/bugs/BUG-021-npa-template-private-state.md.
func TestAfterSuccess_TemplateFilenameNotRewritten_BlockRule(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "42",
			"rule_name": "block-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"match_criteria_action": {
					"action_name": "block",
					"emit_alert": true,
					"template": "99.html"
				},
				"privateApps": ["[my-app]"],
				"access_method": ["Client"]
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "createNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	bodyStr := string(rawBody)

	// Template must NOT be rewritten — state stores the filename returned by the API.
	// The plan modifier handles reconciliation with the configured display name.
	if !strings.Contains(bodyStr, "99.html") {
		t.Errorf("expected .html filename to be preserved in state (no AfterSuccess rewrite), got: %s", bodyStr)
	}
	if strings.Contains(bodyStr, "My Block Template") {
		t.Errorf("expected display name NOT to appear in response (no AfterSuccess rewrite), got: %s", bodyStr)
	}

	// But private app brackets must still be stripped.
	if strings.Contains(bodyStr, `"[my-app]"`) {
		t.Errorf("expected private app brackets to be stripped, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, `"my-app"`) {
		t.Errorf("expected 'my-app' in result body, got: %s", bodyStr)
	}
}

// TestAfterSuccess_TemplateFilenameNotRewritten_PeriodicReauth verifies that
// AfterSuccess does not rewrite .html filenames for periodic_reauth rules either.
// The plan modifier handles reconciliation via private state for all rule types.
// See docs/bugs/BUG-021-npa-template-private-state.md.
func TestAfterSuccess_TemplateFilenameNotRewritten_PeriodicReauth(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "77",
			"rule_name": "reauth-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"match_criteria_action": {
					"action_name": "periodic_reauth",
					"template": "10.html"
				},
				"privateApps": ["[my-app]"],
				"access_method": ["Client"],
				"os": ["Windows"],
				"periodic_reauth": {
					"reauth_interval": "8",
					"reauth_interval_unit": "hours"
				}
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	bodyStr := string(rawBody)

	// The .html filename must be preserved in state — AfterSuccess no longer rewrites.
	if !strings.Contains(bodyStr, "10.html") {
		t.Errorf("expected '10.html' to be preserved in state, got: %s", bodyStr)
	}
	if strings.Contains(bodyStr, "My Reauth Template") {
		t.Errorf("expected display name NOT to appear in response, got: %s", bodyStr)
	}
}

// TestAfterSuccess_PrivateAppsStripped_TemplateFilenameKept verifies the two
// remaining AfterSuccess transformations: bracket stripping for privateApps,
// and no rewriting of the template field. H13 from the test spec.
func TestAfterSuccess_PrivateAppsStripped_TemplateFilenameKept(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "1",
			"rule_name": "test",
			"enabled": "1",
			"rule_data": {
				"match_criteria_action": {
					"action_name": "block",
					"template": "5.html"
				},
				"privateApps": ["[app-one]", "[app-two]"],
				"access_method": ["Client"]
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	bodyStr := string(rawBody)

	// Template filename preserved
	if !strings.Contains(bodyStr, "5.html") {
		t.Errorf("expected template '5.html' to be preserved, got: %s", bodyStr)
	}
	// Private app brackets stripped
	if strings.Contains(bodyStr, "[app-one]") {
		t.Errorf("expected '[app-one]' brackets stripped, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, `"app-one"`) {
		t.Errorf("expected 'app-one' in result, got: %s", bodyStr)
	}
}

// TestAfterSuccess_NotifyTemplatesArrayOfObjects verifies that the hook correctly
// handles responses where notify.templates is an array of objects rather than
// the expected [][]string type. Using generic map parsing avoids unmarshal errors.
// H14 from the test spec (regression guard from 0.4.11).
func TestAfterSuccess_NotifyTemplatesArrayOfObjects(t *testing.T) {
	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "10",
			"rule_name": "notify-rule",
			"enabled": "1",
			"rule_data": {
				"match_criteria_action": {"action_name": "allow"},
				"privateApps": ["[my-app]"],
				"access_method": ["Client"],
				"notify": {
					"emails": ["user@example.com"],
					"templates": [{"name": "t1", "id": 1}, {"name": "t2", "id": 2}]
				}
			}
		},
		"status": "success"
	}`

	ctx, res := buildFakePolicyResponse(body, "getNPARules")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed with notify.templates as array-of-objects: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil response")
	}
}
