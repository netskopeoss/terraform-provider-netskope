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
