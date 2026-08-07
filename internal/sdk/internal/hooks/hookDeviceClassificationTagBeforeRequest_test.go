package hooks

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

func buildFakeDeviceClassificationTagRequest(body string, operationID string) (BeforeRequestContext, *http.Request) {
	req, _ := http.NewRequest("POST", "https://example.com/api/v2/deviceclassification/tags", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	ctx := BeforeRequestContext{
		HookContext: HookContext{
			OperationID: operationID,
		},
	}
	return ctx, req
}

// TestDeviceClassificationTag_CreateWrapsObjectInArray verifies that the single-object
// SDK payload is wrapped into a single-element array for the API.
// The API expects [{"name": "...", "description": "..."}] but the SDK sends a bare object.
func TestDeviceClassificationTag_CreateWrapsObjectInArray(t *testing.T) {
	hook := &deviceClassificationTagRequest{}

	body := `{"name": "my-tag", "description": "a test tag"}`
	ctx, req := buildFakeDeviceClassificationTagRequest(body, "createDeviceClassificationTag")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)

	// Output should be a JSON array
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		t.Fatalf("expected output to be a JSON array, got parse error: %v\nbody: %s", err, raw)
	}
	if len(arr) != 1 {
		t.Fatalf("expected array with 1 element, got %d", len(arr))
	}

	// The single element should contain the original fields
	var item map[string]interface{}
	if err := json.Unmarshal(arr[0], &item); err != nil {
		t.Fatalf("failed to unmarshal array element: %v", err)
	}
	if item["name"] != "my-tag" {
		t.Errorf("expected name = 'my-tag', got %v", item["name"])
	}
	if item["description"] != "a test tag" {
		t.Errorf("expected description = 'a test tag', got %v", item["description"])
	}
}

// TestDeviceClassificationTag_ContentLengthMatchesBody verifies that the Content-Length
// is set to the exact byte length of the modified body. If Content-Length is stale
// (still the original object length), the server receives truncated or misframed data.
func TestDeviceClassificationTag_ContentLengthMatchesBody(t *testing.T) {
	hook := &deviceClassificationTagRequest{}

	body := `{"name": "tag", "description": "desc"}`
	ctx, req := buildFakeDeviceClassificationTagRequest(body, "createDeviceClassificationTag")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed: %v", err)
	}

	raw, _ := io.ReadAll(result.Body)
	// ContentLength must match the actual bytes in the body (not the original unmodified length)
	if result.ContentLength != int64(len(raw)) {
		t.Errorf("ContentLength mismatch: header says %d but body is %d bytes", result.ContentLength, len(raw))
	}
	// Also sanity-check: the body is now a JSON array (starts with '[')
	if len(raw) == 0 || raw[0] != '[' {
		t.Errorf("expected body to be a JSON array starting with '[', got: %s", raw)
	}
}

// TestDeviceClassificationTag_NonMatchingOperationPassthrough verifies that operations
// other than createDeviceClassificationTag are passed through as bare objects (no wrapping).
func TestDeviceClassificationTag_NonMatchingOperationPassthrough(t *testing.T) {
	hook := &deviceClassificationTagRequest{}

	body := `{"name": "my-tag", "description": "desc"}`
	ctx, req := buildFakeDeviceClassificationTagRequest(body, "updateDeviceClassificationTag")

	result, err := hook.BeforeRequest(ctx, req)
	if err != nil {
		t.Fatalf("BeforeRequest failed for non-matching operation: %v", err)
	}

	// Read the body from the original request (not from result, since passthrough doesn't re-set body)
	// For passthrough, the hook returns the original req — body may still be on it
	_ = result // passthrough result is the original req

	// Re-read: since we read the body above to create the req, we need to verify by re-building
	// Just check that no array wrapping happened (if the hook re-sets body, check it isn't an array)
	if result.Body != nil {
		raw, _ := io.ReadAll(result.Body)
		if len(raw) > 0 {
			var arr []json.RawMessage
			if err := json.Unmarshal(raw, &arr); err == nil {
				t.Errorf("expected non-array body for non-matching operation, got array: %s", raw)
			}
		}
	}
}
