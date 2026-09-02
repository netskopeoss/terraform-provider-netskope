package hooks

import (
	"context"
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

// buildFakePolicyResponseWithDisplayName constructs a fake *http.Response whose
// associated Request carries the given template display name in its context, as
// BeforeRequest does for createNPARules. Use this for testing the cache population
// path in AfterSuccess.
func buildFakePolicyResponseWithDisplayName(body string, operationID string, displayName string) (AfterSuccessContext, *http.Response) {
	req, _ := http.NewRequestWithContext(
		withNPATemplateDisplayName(context.Background(), displayName),
		http.MethodPost,
		"https://example.com/api/v2/policy/npa/rules",
		nil,
	)
	hookCtx, res := buildFakePolicyResponse(body, operationID)
	res.Request = req
	return hookCtx, res
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

// TestAfterSuccess_CreatePopulatesTemplateCacheAndFixesResponse verifies that when
// createNPARules is processed for a block rule, the AfterSuccess hook:
//  1. Reads the display name from the request context (set by BeforeRequest)
//  2. Populates npaTemplateCache with the file→display mapping
//  3. Substitutes the display name into the response body
//
// This ensures Terraform state stores the display name rather than the API-returned
// .html file name, eliminating the perpetual drift for block rules.
// Cache substitution is NOT applied to periodic_reauth rules — see
// TestAfterSuccess_PeriodicReauthFilenamePreservedInResponse.
// See https://github.com/netskopeoss/terraform-provider-netskope/issues/116
func TestAfterSuccess_CreatePopulatesTemplateCacheAndFixesResponse(t *testing.T) {
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

	ctx, res := buildFakePolicyResponseWithDisplayName(body, "createNPARules", "My Block Template")

	result, err := hook.AfterSuccess(ctx, res)
	if err != nil {
		t.Fatalf("AfterSuccess failed: %v", err)
	}

	rawBody, err := io.ReadAll(result.Body)
	if err != nil {
		t.Fatalf("failed to read result body: %v", err)
	}
	bodyStr := string(rawBody)

	// Template should be replaced with the display name
	if strings.Contains(bodyStr, "99.html") {
		t.Errorf("expected .html file name to be replaced in response, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "My Block Template") {
		t.Errorf("expected display name 'My Block Template' in response, got: %s", bodyStr)
	}

	// Cache should be populated
	displayName, ok := npaTemplateCacheGet("99.html")
	if !ok {
		t.Error("expected cache to be populated with file→display mapping")
	}
	if displayName != "My Block Template" {
		t.Errorf("expected cache entry 'My Block Template', got %q", displayName)
	}
}

// TestAfterSuccess_ReadUsesTemplateCacheToFixBlockResponse verifies that getNPARules
// responses for block rules have their .html template file name replaced with the
// display name from the cache (populated by a prior create). This ensures state
// stays consistent after any Read/refresh following the initial create.
func TestAfterSuccess_ReadUsesTemplateCacheToFixBlockResponse(t *testing.T) {
	hook := &myPolicyResponse{}

	// Seed the cache as if a prior block rule create had run
	npaTemplateCacheSet("7.html", "My Block Template")

	body := `{
		"data": {
			"rule_id": "55",
			"rule_name": "block-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"match_criteria_action": {
					"action_name": "block",
					"emit_alert": true,
					"template": "7.html"
				},
				"privateApps": ["[my-app]"],
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

	if strings.Contains(bodyStr, "7.html") {
		t.Errorf("expected .html file name to be replaced by cache lookup for block rule, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "My Block Template") {
		t.Errorf("expected 'My Block Template' in response, got: %s", bodyStr)
	}
}

// TestAfterSuccess_PeriodicReauthDisplayNameRestoredViaTemplatesAPI verifies that
// getNPARules responses for periodic_reauth rules have their .html filename
// replaced with the display name from the templates API cache. This allows state
// to store the display name, eliminating drift when users configure templates by
// display name. The templates API cache is keyed on action_type so block and
// periodic_reauth entries do not collide.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestAfterSuccess_PeriodicReauthDisplayNameRestoredViaTemplatesAPI(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	npaTemplatesAPISeedForTest([]npaTemplatesAPIEntry{
		{FileName: "10.html", Name: "My Reauth Template", ActionType: "periodic_reauth"},
	})

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

	// The filename must be replaced with the display name from the templates API
	if strings.Contains(bodyStr, "10.html") {
		t.Errorf("expected .html filename to be replaced by templates API lookup, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "My Reauth Template") {
		t.Errorf("expected display name 'My Reauth Template' in response, got: %s", bodyStr)
	}
}

// TestAfterSuccess_PeriodicReauthDisplayNameRestoredViaGlobalFallback verifies that
// AfterSuccess can translate a .html filename to a display name even when the
// templates API cache entry uses a DIFFERENT action_type than the rule's action_name.
// This covers the real-world case where a "block"-typed template is used in a
// periodic_reauth rule: the filename is stored under "block:1.html" in the cache
// but the rule's action_name is "periodic_reauth".
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestAfterSuccess_PeriodicReauthDisplayNameRestoredViaGlobalFallback(t *testing.T) {
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	// Seed the cache with a "block"-typed template — as the templates API returns it.
	// (User created "tf-test-template" as a block template but uses it for periodic_reauth.)
	npaTemplatesAPISeedForTest([]npaTemplatesAPIEntry{
		{FileName: "1.html", Name: "tf-test-template", ActionType: "block"},
	})

	hook := &myPolicyResponse{}

	body := `{
		"data": {
			"rule_id": "88",
			"rule_name": "reauth-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"match_criteria_action": {
					"action_name": "periodic_reauth",
					"template": "1.html"
				},
				"privateApps": ["[my-app]"],
				"access_method": ["Client"],
				"os": ["Windows"],
				"periodic_reauth": {
					"reauth_interval": "60",
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

	// Global fallback must translate "1.html" → "tf-test-template"
	// even though the cache entry has action_type "block", not "periodic_reauth".
	if strings.Contains(bodyStr, "1.html") {
		t.Errorf("expected .html filename to be replaced via global fallback, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "tf-test-template") {
		t.Errorf("expected display name 'tf-test-template' in response via global fallback, got: %s", bodyStr)
	}
}

// TestAfterSuccess_PeriodicReauthFilenamePreservedInResponse verifies that getNPARules
// responses for periodic_reauth rules have their .html filename preserved as-is in
// state when the templates API cache has no entry. This is the graceful fallback for
// tenants where /api/v2/templates/usernotifications is unavailable.
// See docs/bugs/BUG-020-periodic-reauth-template.md
func TestAfterSuccess_PeriodicReauthFilenamePreservedInResponse(t *testing.T) {
	// Ensure the templates API cache is empty for this test — no substitution should occur.
	npaTemplatesAPIResetForTest()
	t.Cleanup(npaTemplatesAPIResetForTest)

	hook := &myPolicyResponse{}

	// Session cache must NOT be applied to periodic_reauth rules either.
	npaTemplateCacheSet("10.html", "Should Not Be Substituted")

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

	// The .html filename must remain unchanged in state for periodic_reauth
	if !strings.Contains(bodyStr, "10.html") {
		t.Errorf("expected filename '10.html' to be preserved in periodic_reauth response, got: %s", bodyStr)
	}
	if strings.Contains(bodyStr, "Should Not Be Substituted") {
		t.Errorf("expected cache NOT to be applied to periodic_reauth response, got: %s", bodyStr)
	}
}
