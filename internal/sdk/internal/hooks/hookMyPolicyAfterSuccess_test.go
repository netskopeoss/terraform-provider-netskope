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
// createNPARules is processed, the AfterSuccess hook:
//  1. Reads the display name from the request context (set by BeforeRequest)
//  2. Populates npaTemplateCache with the file→display mapping
//  3. Substitutes the display name into the response body
//
// This ensures Terraform state stores the display name rather than the API-returned
// .html file name, eliminating the perpetual drift for block and periodic_reauth rules.
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

// TestAfterSuccess_ReadUsesTemplateCacheToFixResponse verifies that getNPARules
// responses have their .html template file name replaced with the display name
// from the cache (populated by a prior create). This ensures state stays consistent
// after any Read/refresh following the initial create.
func TestAfterSuccess_ReadUsesTemplateCacheToFixResponse(t *testing.T) {
	hook := &myPolicyResponse{}

	// Seed the cache as if a prior create had run
	npaTemplateCacheSet("7.html", "Periodic Reauth Template")

	body := `{
		"data": {
			"rule_id": "55",
			"rule_name": "reauth-rule",
			"enabled": "1",
			"rule_data": {
				"policy_type": "private-app",
				"match_criteria_action": {
					"action_name": "periodic_reauth",
					"template": "7.html"
				},
				"privateApps": ["[my-app]"],
				"access_method": ["Client"],
				"os": ["Windows"]
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
		t.Errorf("expected .html file name to be replaced by cache lookup, got: %s", bodyStr)
	}
	if !strings.Contains(bodyStr, "Periodic Reauth Template") {
		t.Errorf("expected 'Periodic Reauth Template' in response, got: %s", bodyStr)
	}
}
