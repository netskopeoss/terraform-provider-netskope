package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// aigRateLimitRequestHook strips criteria fields that are excluded by the API
// based on the value of apply_on:
//   - apply_on = "ai"  → mcp_server_ids, tools, resources, prompts are excluded
//   - apply_on = "mcp" → ai_provider_ids, models are excluded
//
// The SDK serializes Computed+Optional list fields as [] when unset, but the
// API rejects any excluded field even when empty.
type aigRateLimitRequestHook struct{}

var _ beforeRequestHook = (*aigRateLimitRequestHook)(nil)

func (h *aigRateLimitRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "createAigRateLimit" && hookCtx.OperationID != "updateAigRateLimit" {
		return req, nil
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("aigRateLimitRequestHook: unable to read request body: %w", err)
	}

	var requestMap map[string]interface{}
	if err := json.Unmarshal(body, &requestMap); err != nil {
		return nil, fmt.Errorf("aigRateLimitRequestHook: unable to unmarshal request: %w", err)
	}

	criteria, ok := requestMap["criteria"].(map[string]interface{})
	if !ok {
		// No criteria object — nothing to strip.
		req.Body = io.NopCloser(strings.NewReader(string(body)))
		return req, nil
	}

	applyOn, _ := criteria["apply_on"].(string)

	switch applyOn {
	case "ai":
		// MCP-specific fields are excluded when apply_on = "ai"
		for _, field := range []string{"mcp_server_ids", "tools", "resources", "prompts"} {
			delete(criteria, field)
		}
	case "mcp":
		// AI-specific fields are excluded when apply_on = "mcp"
		for _, field := range []string{"ai_provider_ids", "models"} {
			delete(criteria, field)
		}
	}

	requestMap["criteria"] = criteria

	modifiedBody, err := json.Marshal(requestMap)
	if err != nil {
		return nil, fmt.Errorf("aigRateLimitRequestHook: unable to marshal modified request: %w", err)
	}

	req.Body = io.NopCloser(strings.NewReader(string(modifiedBody)))
	req.ContentLength = int64(len(modifiedBody))

	return req, nil
}
