package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// npaPrivateAppTagHook handles the non-standard request/response shape for the
// NPA private app tag create endpoint (POST /steering/apps/private/tags).
//
// The API requires the request body to be wrapped in a {"tags": [...]} envelope
// and returns a response in the shape {"data": {"id": null, "tags": [...]}, "status": "success"}.
// Terraform expects a flat {"tag_name": "..."} request and a standard
// {"data": {"tag_id": N, "tag_name": "..."}, "status": "success"} response.
//
// This hook:
//   - BeforeRequest: wraps {tag_name: "..."} → {"tags": [{"tag_name": "..."}]}
//   - AfterSuccess:  unwraps data.tags[0] → {"data": {tag_id, tag_name}, "status": "success"}
type npaPrivateAppTagHook struct{}

var _ beforeRequestHook = (*npaPrivateAppTagHook)(nil)
var _ afterSuccessHook = (*npaPrivateAppTagHook)(nil)

func (h *npaPrivateAppTagHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "createNPAPrivateAppTag" {
		return req, nil
	}

	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to read request body: %w", err)
	}

	// Parse the flat {tag_name: "..."} body
	var flat map[string]interface{}
	if err := json.Unmarshal(body, &flat); err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to unmarshal request: %w", err)
	}

	// Wrap into {"tags": [{tag_name: "..."}]}
	wrapped, err := json.Marshal(map[string]interface{}{
		"tags": []interface{}{flat},
	})
	if err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to marshal wrapped request: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(wrapped))
	req.ContentLength = int64(len(wrapped))
	req.Header.Set("Content-Length", fmt.Sprintf("%d", len(wrapped)))

	return req, nil
}

func (h *npaPrivateAppTagHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID != "createNPAPrivateAppTag" {
		return res, nil
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to read response body: %w", err)
	}

	// Parse the actual API response: {"data": {"id": null, "tags": [{tag_id, tag_name}]}, "status": "success"}
	var apiResp struct {
		Data struct {
			Tags []struct {
				TagID   int    `json:"tag_id"`
				TagName string `json:"tag_name"`
			} `json:"tags"`
		} `json:"data"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to unmarshal response: %w", err)
	}

	if len(apiResp.Data.Tags) == 0 {
		return nil, fmt.Errorf("npaTag create hook: API returned no tags in response: %s", string(body))
	}

	// Normalize to {"data": {tag_id, tag_name}, "status": "success"} — matches npa_tag_get_response
	normalized, err := json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"tag_id":   apiResp.Data.Tags[0].TagID,
			"tag_name": apiResp.Data.Tags[0].TagName,
		},
		"status": apiResp.Status,
	})
	if err != nil {
		return nil, fmt.Errorf("npaTag create hook: failed to marshal normalized response: %w", err)
	}

	return replaceBody(res, normalized), nil
}
