package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakeAIGRateLimitRequest(body string, operationID string) (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/aig/ratelimit", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, req
}

// TestAIGRateLimit_ApplyOnAIStripesMCPFields verifies that when apply_on = "ai",
// the MCP-specific fields (mcp_server_ids, tools, resources, prompts) are stripped.
func TestAIGRateLimit_ApplyOnAIStripesMCPFields(t *testing.T) {
	hook := &aigRateLimitRequestHook{}

	body := `{
		"name": "test",
		"criteria": {
			"apply_on": "ai",
			"ai_provider_ids": ["provider-1"],
			"models": ["gpt-4"],
			"mcp_server_ids": ["server-1"],
			"tools": ["tool-a"],
			"resources": ["res-b"],
			"prompts": ["prompt-c"]
		}
	}`
	ctx, req := buildFakeAIGRateLimitRequest(body, "createAigRateLimit")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	criteria := out["criteria"].(map[string]interface{})
	for _, field := range []string{"mcp_server_ids", "tools", "resources", "prompts"} {
		if _, exists := criteria[field]; exists {
			t.Errorf("expected %q stripped for apply_on=ai, still present in criteria", field)
		}
	}
	// AI-specific fields should remain
	for _, field := range []string{"ai_provider_ids", "models"} {
		if _, exists := criteria[field]; !exists {
			t.Errorf("expected %q preserved for apply_on=ai, missing from criteria", field)
		}
	}
}

// TestAIGRateLimit_ApplyOnMCPStripesAIFields verifies that when apply_on = "mcp",
// the AI-specific fields (ai_provider_ids, models) are stripped.
func TestAIGRateLimit_ApplyOnMCPStripesAIFields(t *testing.T) {
	hook := &aigRateLimitRequestHook{}

	body := `{
		"name": "test",
		"criteria": {
			"apply_on": "mcp",
			"ai_provider_ids": ["provider-1"],
			"models": ["gpt-4"],
			"mcp_server_ids": ["server-1"],
			"tools": ["tool-a"],
			"resources": ["res-b"],
			"prompts": ["prompt-c"]
		}
	}`
	ctx, req := buildFakeAIGRateLimitRequest(body, "updateAigRateLimit")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	criteria := out["criteria"].(map[string]interface{})
	for _, field := range []string{"ai_provider_ids", "models"} {
		if _, exists := criteria[field]; exists {
			t.Errorf("expected %q stripped for apply_on=mcp, still present in criteria", field)
		}
	}
	// MCP-specific fields should remain
	for _, field := range []string{"mcp_server_ids", "tools", "resources", "prompts"} {
		if _, exists := criteria[field]; !exists {
			t.Errorf("expected %q preserved for apply_on=mcp, missing from criteria", field)
		}
	}
}

// TestAIGRateLimit_NoCriteriaPassesThrough verifies that requests without a criteria
// object are passed through unchanged without crashing.
func TestAIGRateLimit_NoCriteriaPassesThrough(t *testing.T) {
	hook := &aigRateLimitRequestHook{}

	body := `{"name": "test"}`
	ctx, req := buildFakeAIGRateLimitRequest(body, "createAigRateLimit")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil request")
	}
}

// TestAIGRateLimit_UnknownApplyOnPreservesCriteria verifies that an unknown apply_on
// value leaves all criteria fields intact (no accidental stripping).
func TestAIGRateLimit_UnknownApplyOnPreservesCriteria(t *testing.T) {
	hook := &aigRateLimitRequestHook{}

	body := `{
		"name": "test",
		"criteria": {
			"apply_on": "all",
			"ai_provider_ids": ["p1"],
			"mcp_server_ids": ["s1"]
		}
	}`
	ctx, req := buildFakeAIGRateLimitRequest(body, "createAigRateLimit")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("failed to unmarshal result: %v", err)
	}

	criteria := out["criteria"].(map[string]interface{})
	for _, field := range []string{"ai_provider_ids", "mcp_server_ids"} {
		if _, exists := criteria[field]; !exists {
			t.Errorf("expected %q preserved for unknown apply_on, missing from criteria", field)
		}
	}
}

// TestAIGRateLimit_NonMatchingOperationPassthrough verifies that operations other
// than createAigRateLimit / updateAigRateLimit are passed through unchanged.
func TestAIGRateLimit_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &aigRateLimitRequestHook{}

	body := `{"criteria": {"apply_on": "ai", "mcp_server_ids": ["s1"]}}`
	ctx, req := buildFakeAIGRateLimitRequest(body, "deleteAigRateLimit")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	// mcp_server_ids should NOT be stripped for non-matching op
	raw, _ := io.ReadAll(result.Body)
	if !strings.Contains(string(raw), "mcp_server_ids") {
		t.Errorf("expected mcp_server_ids unchanged for non-matching operation, got: %s", raw)
	}
}
