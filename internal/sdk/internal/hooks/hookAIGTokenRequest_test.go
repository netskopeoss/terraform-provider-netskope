package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakeAIGTokenRequest(body string, operationID string) (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/aig/token", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, req
}

// TestAIGToken_ZeroExpireInRemoved verifies that expire_in with value=0 and unit=""
// is stripped from the request body. The Terraform framework populates these zero values
// when the attribute is null in the plan — the API rejects them as invalid.
func TestAIGToken_ZeroExpireInRemoved(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "expire_in": {"value": 0, "unit": ""}}`
	ctx, req := buildFakeAIGTokenRequest(body, "createAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["expire_in"]; exists {
		t.Errorf("expected expire_in stripped for zero value, still present in: %s", raw)
	}
}

// TestAIGToken_ValidExpireInPreserved verifies that a properly configured expire_in
// (non-zero value + non-empty unit) is preserved in the request body.
func TestAIGToken_ValidExpireInPreserved(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "expire_in": {"value": 30, "unit": "days"}}`
	ctx, req := buildFakeAIGTokenRequest(body, "createAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["expire_in"]; !exists {
		t.Errorf("expected expire_in preserved for valid value, missing from: %s", raw)
	}
	expireIn := out["expire_in"].(map[string]interface{})
	if expireIn["unit"] != "days" {
		t.Errorf("expected expire_in.unit = 'days', got %v", expireIn["unit"])
	}
}

// TestAIGToken_NullExpireInRemoved verifies that a null expire_in is removed.
func TestAIGToken_NullExpireInRemoved(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "expire_in": null}`
	ctx, req := buildFakeAIGTokenRequest(body, "updateAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["expire_in"]; exists {
		t.Errorf("expected null expire_in stripped, still present in: %s", raw)
	}
}

// TestAIGToken_ZeroValueNonEmptyUnitRemoved verifies that expire_in with value=0
// (even if unit is set) is stripped since value=0 is an invalid expiry.
func TestAIGToken_ZeroValueNonEmptyUnitRemoved(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "expire_in": {"value": 0, "unit": "days"}}`
	ctx, req := buildFakeAIGTokenRequest(body, "createAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}
	if _, exists := out["expire_in"]; exists {
		t.Errorf("expected expire_in stripped when value=0, still present in: %s", raw)
	}
}

// TestAIGToken_NoExpireInPreservesBody verifies that requests without expire_in
// are passed through unchanged.
func TestAIGToken_NoExpireInPreservesBody(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "description": "test"}`
	ctx, req := buildFakeAIGTokenRequest(body, "createAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil request")
	}

	raw, _ := io.ReadAll(result.Body)
	if !strings.Contains(string(raw), "my-token") {
		t.Errorf("expected body preserved, got: %s", raw)
	}
}

// TestAIGToken_NonMatchingOperationPassthrough verifies that operations other than
// createAigToken / updateAigToken are passed through unchanged.
func TestAIGToken_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &aigTokenRequestHook{}

	body := `{"name": "my-token", "expire_in": {"value": 0, "unit": ""}}`
	ctx, req := buildFakeAIGTokenRequest(body, "deleteAigToken")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed for non-matching operation: %v", err)
	}

	// expire_in should NOT be stripped for non-matching op
	raw, _ := io.ReadAll(result.Body)
	if !strings.Contains(string(raw), "expire_in") {
		t.Errorf("expected expire_in unchanged for non-matching operation, got: %s", raw)
	}
}
