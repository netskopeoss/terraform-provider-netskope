package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// cciDataAfterSuccess flattens the CCI data response.
//
// The API returns:
//
//	{ "data": { "category": [...] }, "status": "Success", "status_code": 200 }
//
// The generated SDK struct expects:
//
//	{ "categories": [...] }
//
// This hook rewrites the response body so the SDK can unmarshal it correctly.
type cciDataAfterSuccess struct{}

var _ afterSuccessHook = (*cciDataAfterSuccess)(nil)

func (h *cciDataAfterSuccess) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID != "getCCICategoryList" {
		return res, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("cciDataAfterSuccess: failed to read response body: %w", err)
	}

	// Parse the raw response envelope
	var envelope struct {
		Data       map[string]json.RawMessage `json:"data"`
		Status     string                     `json:"status"`
		StatusCode int                        `json:"status_code"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		// Not the expected shape — pass through unchanged
		res.Body = io.NopCloser(strings.NewReader(string(body)))
		return res, nil
	}

	categoryRaw, ok := envelope.Data["category"]
	if !ok {
		// No category field — return empty categories list
		categoryRaw = json.RawMessage("[]")
	}

	// Rewrite to the flat shape the SDK struct expects
	flat := map[string]json.RawMessage{
		"categories": categoryRaw,
	}
	rewritten, err := json.Marshal(flat)
	if err != nil {
		return nil, fmt.Errorf("cciDataAfterSuccess: failed to marshal rewritten body: %w", err)
	}

	res.Body = io.NopCloser(strings.NewReader(string(rewritten)))
	res.ContentLength = int64(len(rewritten))
	return res, nil
}
