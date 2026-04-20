package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
)

// deviceClassificationTagsResponse wraps the raw array response from
// GET /deviceclassification/tags into the {"tags": [...]} shape that
// the Speakeasy-generated SDK expects.
type deviceClassificationTagsResponse struct{}

var _ afterSuccessHook = (*deviceClassificationTagsResponse)(nil)

func (d *deviceClassificationTagsResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID != "listDeviceClassificationTags" {
		return res, nil
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("deviceClassificationTags hook: failed to read response body: %w", err)
	}

	// The API returns a raw JSON array: [{...}, {...}]
	// Wrap it into {"tags": [...]} to match the OAS schema.
	var tags []json.RawMessage
	if err := json.Unmarshal(body, &tags); err != nil {
		return nil, fmt.Errorf("deviceClassificationTags hook: failed to unmarshal array: %w", err)
	}

	wrapped, err := json.Marshal(map[string]interface{}{
		"tags": tags,
	})
	if err != nil {
		return nil, fmt.Errorf("deviceClassificationTags hook: failed to marshal wrapped response: %w", err)
	}

	res.Body = io.NopCloser(bytes.NewReader(wrapped))
	res.ContentLength = int64(len(wrapped))
	res.Header.Set("Content-Length", strconv.Itoa(len(wrapped)))

	return res, nil
}
