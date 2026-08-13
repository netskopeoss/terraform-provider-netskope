package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// deviceClassificationTagsResponse handles response transformations for all
// device classification endpoints:
//   - listDeviceClassificationTags: wraps raw array into {"tags": [...]}
//   - createDeviceClassificationTag: transforms {status, data:[id]} into a
//     tag item by doing a follow-up GET
//   - updateDeviceClassificationTag: transforms {status: true} into a tag
//     item by doing a follow-up GET
//   - listDeviceClassificationOptions: wraps raw array into {"options": [...]}
type deviceClassificationTagsResponse struct{}

var _ afterSuccessHook = (*deviceClassificationTagsResponse)(nil)

func (d *deviceClassificationTagsResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "listDeviceClassificationTags":
		return d.wrapArray(res, "tags")
	case "listDeviceClassificationOptions":
		return d.wrapArray(res, "options")
	case "createDeviceClassificationTag":
		return d.handleCreateResponse(res)
	case "updateDeviceClassificationTag":
		return d.handleUpdateResponse(res)
	default:
		return res, nil
	}
}

// wrapArray wraps a raw JSON array response [...] into {key: [...]}
func (d *deviceClassificationTagsResponse) wrapArray(res *http.Response, key string) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceClassification hook: failed to read response body: %w", err)
	}

	var items []json.RawMessage
	if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("deviceClassification hook: failed to unmarshal array: %w", err)
	}

	wrapped, err := json.Marshal(map[string]interface{}{
		key: items,
	})
	if err != nil {
		return nil, fmt.Errorf("deviceClassification hook: failed to marshal wrapped response: %w", err)
	}

	return replaceBody(res, wrapped), nil
}

// handleCreateResponse transforms {status: true, data: [id]} into the full
// tag object by doing a follow-up GET /deviceclassification/tags/{id}.
func (d *deviceClassificationTagsResponse) handleCreateResponse(res *http.Response) (*http.Response, error) {
	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceClassification create hook: failed to read response body: %w", err)
	}

	var createResp struct {
		Status bool  `json:"status"`
		Data   []int `json:"data"`
	}
	if err := json.Unmarshal(body, &createResp); err != nil {
		return nil, fmt.Errorf("deviceClassification create hook: failed to unmarshal response: %w", err)
	}

	if !createResp.Status || len(createResp.Data) == 0 {
		return nil, fmt.Errorf("deviceClassification create hook: create failed or no ID returned: %s", string(body))
	}

	tagID := createResp.Data[0]
	return d.fetchTag(res, tagID)
}

// handleUpdateResponse transforms {status: true} into the full tag object
// by doing a follow-up GET using the tag ID from the request URL.
func (d *deviceClassificationTagsResponse) handleUpdateResponse(res *http.Response) (*http.Response, error) {
	// Drain the response body (we don't need {status: true})
	io.ReadAll(res.Body)
	res.Body.Close()

	// Extract tag ID from the request URL path: /deviceclassification/tags/{id}
	path := res.Request.URL.Path
	parts := splitPath(path)
	if len(parts) == 0 {
		return nil, fmt.Errorf("deviceClassification update hook: cannot extract tag ID from URL: %s", path)
	}
	tagIDStr := parts[len(parts)-1]
	tagID, err := strconv.Atoi(tagIDStr)
	if err != nil {
		return nil, fmt.Errorf("deviceClassification update hook: invalid tag ID %q: %w", tagIDStr, err)
	}

	return d.fetchTag(res, tagID)
}

// fetchTag does a GET /deviceclassification/tags/{id} and replaces the
// response body with the returned tag object.
func (d *deviceClassificationTagsResponse) fetchTag(origRes *http.Response, tagID int) (*http.Response, error) {
	// Build the GET URL from the original request
	getURL := *origRes.Request.URL
	getURL.Path = fmt.Sprintf("/api/v2/deviceclassification/tags/%d", tagID)
	getURL.RawQuery = ""

	getReq, err := http.NewRequestWithContext(origRes.Request.Context(), "GET", getURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("deviceClassification hook: failed to build GET request: %w", err)
	}

	// Copy auth headers from original request
	for _, h := range []string{"Netskope-Api-Token", "Authorization"} {
		if v := origRes.Request.Header.Get(h); v != "" {
			getReq.Header.Set(h, v)
		}
	}

	client := &http.Client{}
	getRes, err := client.Do(getReq)
	if err != nil {
		return nil, fmt.Errorf("deviceClassification hook: follow-up GET failed: %w", err)
	}

	tagBody, err := io.ReadAll(getRes.Body)
	getRes.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceClassification hook: failed to read GET response: %w", err)
	}

	return replaceBody(origRes, tagBody), nil
}

// replaceBody returns the response with a new body and ensures Content-Type is
// set to application/json so SDK content-type checks succeed even when the
// original response had an empty or absent Content-Type (e.g. 201 with no body).
func replaceBody(res *http.Response, body []byte) *http.Response {
	res.Body = io.NopCloser(bytes.NewReader(body))
	res.ContentLength = int64(len(body))
	if res.Header == nil {
		res.Header = make(http.Header)
	}
	res.Header.Set("Content-Length", strconv.Itoa(len(body)))
	res.Header.Set("Content-Type", "application/json")
	return res
}

// splitPath splits a URL path into segments, skipping empty strings.
func splitPath(path string) []string {
	var parts []string
	for _, p := range bytes.Split([]byte(path), []byte("/")) {
		if len(p) > 0 {
			parts = append(parts, string(p))
		}
	}
	return parts
}
