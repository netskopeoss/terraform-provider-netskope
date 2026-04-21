package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// deviceClassificationTagRequest wraps the create request body from a single
// object into the array format the API expects.
type deviceClassificationTagRequest struct{}

var _ beforeRequestHook = (*deviceClassificationTagRequest)(nil)

func (d *deviceClassificationTagRequest) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "createDeviceClassificationTag" {
		return req, nil
	}

	body, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceClassificationTag create hook: failed to read request body: %w", err)
	}

	// The SDK sends a single object: {"name": "...", "description": "..."}
	// The API expects an array: [{"name": "...", "description": "..."}]
	var item json.RawMessage
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("deviceClassificationTag create hook: failed to unmarshal request: %w", err)
	}

	wrapped, err := json.Marshal([]json.RawMessage{item})
	if err != nil {
		return nil, fmt.Errorf("deviceClassificationTag create hook: failed to marshal array: %w", err)
	}

	req.Body = io.NopCloser(bytes.NewReader(wrapped))
	req.ContentLength = int64(len(wrapped))
	req.Header.Set("Content-Length", strconv.Itoa(len(wrapped)))

	return req, nil
}
