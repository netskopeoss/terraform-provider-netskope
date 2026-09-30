package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// deviceTagRequestHook rewrites read operations for the device tags API.
//
// The /device/tags API does not expose conventional GET endpoints for reading
// tags. Instead, all reads go through POST /device/tags/gettags. This hook
// intercepts the two synthetic GET operations declared in the OAS and rewrites
// them into the correct POST calls before the request is sent:
//
//   - getDeviceTag:   GET /device/tags/{id}  → POST /device/tags/gettags  {"id": <id>}
//   - listDeviceTags: GET /device/tags       → POST /device/tags/gettags  {"limit": 100}
type deviceTagRequestHook struct{}

var _ beforeRequestHook = (*deviceTagRequestHook)(nil)

func (d *deviceTagRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	switch hookCtx.OperationID {
	case "getDeviceTag":
		return d.rewriteSingleGet(req)
	case "listDeviceTags":
		return d.rewriteListGet(req)
	default:
		return req, nil
	}
}

// rewriteSingleGet converts GET /device/tags/{id} → POST /device/tags/gettags with {"id": <id>}.
func (d *deviceTagRequestHook) rewriteSingleGet(req *http.Request) (*http.Request, error) {
	// Extract tag ID from path: /api/v2/device/tags/{id}
	parts := splitPath(req.URL.Path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("deviceTag read hook: cannot extract tag ID from URL: %s", req.URL.Path)
	}
	tagIDStr := parts[len(parts)-1]
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		return nil, fmt.Errorf("deviceTag read hook: invalid tag ID %q in URL %s: %w", tagIDStr, req.URL.Path, err)
	}

	body, err := json.Marshal(map[string]interface{}{"id": tagID})
	if err != nil {
		return nil, fmt.Errorf("deviceTag read hook: failed to marshal request body: %w", err)
	}

	return d.rewriteToPost(req, "/api/v2/devices/device/tags/gettags", body)
}

// rewriteListGet converts GET /devices/device/tags → POST /devices/device/tags/gettags with {"limit": 100}.
func (d *deviceTagRequestHook) rewriteListGet(req *http.Request) (*http.Request, error) {
	body, err := json.Marshal(map[string]interface{}{"limit": 100})
	if err != nil {
		return nil, fmt.Errorf("deviceTag list hook: failed to marshal request body: %w", err)
	}
	return d.rewriteToPost(req, "/api/v2/devices/device/tags/gettags", body)
}

// rewriteToPost creates a new POST request with the given path and body,
// copying auth headers from the original request.
func (d *deviceTagRequestHook) rewriteToPost(req *http.Request, path string, body []byte) (*http.Request, error) {
	newURL := *req.URL
	newURL.Path = path
	newURL.RawQuery = ""

	newReq, err := http.NewRequestWithContext(req.Context(), http.MethodPost, newURL.String(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("deviceTag hook: failed to build POST request: %w", err)
	}

	// Copy all headers from the original request
	for k, vv := range req.Header {
		for _, v := range vv {
			newReq.Header.Add(k, v)
		}
	}
	newReq.Header.Set("Content-Type", "application/json")
	newReq.Header.Set("Content-Length", strconv.Itoa(len(body)))
	newReq.ContentLength = int64(len(body))

	return newReq, nil
}

// deviceTagAfterSuccessHook unwraps the non-standard API response envelopes for
// device tag operations so the SDK sees a flat tag item (or a {tags:[...]} wrapper
// for list operations).
//
// The API wraps all responses in {"success": true, "data": {...}}. For reads via
// gettags the data is further wrapped in a pagination envelope:
// {"success": true, "data": {"data": [{...}], "total_count": 1, ...}}.
type deviceTagAfterSuccessHook struct{}

var _ afterSuccessHook = (*deviceTagAfterSuccessHook)(nil)

func (d *deviceTagAfterSuccessHook) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createDeviceTag", "updateDeviceTag":
		return d.unwrapItem(res)
	case "getDeviceTag":
		return d.unwrapSingleFromPaginated(res)
	case "listDeviceTags":
		return d.unwrapListFromPaginated(res)
	default:
		return res, nil
	}
}

// deviceTagEnvelope is the top-level response shape for create/update:
// {"success": true, "data": {"id": 1, "name": "...", "description": "..."}}
type deviceTagEnvelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
}

// deviceTagPaginatedEnvelope is the shape returned by POST /device/tags/gettags:
// {"success": true, "data": {"data": [{...}], "total_count": 1, "offset": 0, "limit": 1}}
type deviceTagPaginatedEnvelope struct {
	Success bool `json:"success"`
	Data    struct {
		Data       []json.RawMessage `json:"data"`
		TotalCount int               `json:"total_count"`
		Offset     int               `json:"offset"`
		Limit      int               `json:"limit"`
	} `json:"data"`
}

// unwrapItem handles create/update responses: {success, data: {id, name, description}} → flat item.
func (d *deviceTagAfterSuccessHook) unwrapItem(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceTag hook: failed to read response body: %w", err)
	}

	var envelope deviceTagEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("deviceTag hook: failed to unmarshal response envelope: %w", err)
	}

	if !envelope.Success || len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, fmt.Errorf("deviceTag hook: API returned failure or empty data: %s", string(body))
	}

	return replaceBody(res, envelope.Data), nil
}

// unwrapSingleFromPaginated handles getDeviceTag responses:
// {success, data: {data: [{...}], total_count: 1}} → flat tag item.
// Synthesises a 404 if the tag is not found in the results.
func (d *deviceTagAfterSuccessHook) unwrapSingleFromPaginated(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceTag read hook: failed to read response body: %w", err)
	}

	var envelope deviceTagPaginatedEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("deviceTag read hook: failed to unmarshal paginated response: %w", err)
	}

	if !envelope.Success || len(envelope.Data.Data) == 0 {
		// Tag not found — return 404 so the Terraform resource is removed from state
		res.StatusCode = http.StatusNotFound
		res.Status = "404 Not Found"
		notFound := []byte(`{"success":false,"error":{"message":"Tag not found"}}`)
		return replaceBody(res, notFound), nil
	}

	return replaceBody(res, envelope.Data.Data[0]), nil
}

// unwrapListFromPaginated handles listDeviceTags responses:
// {success, data: {data: [{...}], total_count, ...}} → {tags: [{...}]}.
func (d *deviceTagAfterSuccessHook) unwrapListFromPaginated(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceTag list hook: failed to read response body: %w", err)
	}

	var envelope deviceTagPaginatedEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, fmt.Errorf("deviceTag list hook: failed to unmarshal paginated response: %w", err)
	}

	if !envelope.Success {
		return nil, fmt.Errorf("deviceTag list hook: API returned failure: %s", string(body))
	}

	wrapped, err := json.Marshal(map[string]interface{}{
		"tags": envelope.Data.Data,
	})
	if err != nil {
		return nil, fmt.Errorf("deviceTag list hook: failed to marshal wrapped response: %w", err)
	}

	return replaceBody(res, wrapped), nil
}
