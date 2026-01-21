package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// privateAppRequestHook removes problematic empty fields from private app PUT requests
// The Netskope API has a bug where empty objects ({}) and arrays ([]) for certain fields
// are stored as bytes instead of JSON, causing SQL serialization errors.
// See docs/KNOWN_API_ISSUES.md for details.
type privateAppRequestHook struct{}

var (
	_                      beforeRequestHook = (*privateAppRequestHook)(nil)
	privateAppRequestDebug bool              = false
)

func (i *privateAppRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	// Only process private app update operations
	if hookCtx.OperationID != "updateNPAPrivateApp" {
		return req, nil
	}

	if privateAppRequestDebug {
		log.Print("Executing privateAppRequestHook BeforeRequest hook....")
	}

	// Read the request body
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read request body: %w", err)
	}

	if privateAppRequestDebug {
		log.Printf("Original body: %s", string(body))
	}

	// Unmarshal into a generic map to manipulate
	var requestMap map[string]interface{}
	if err := json.Unmarshal(body, &requestMap); err != nil {
		return nil, fmt.Errorf("unable to unmarshal request: %v", err)
	}

	// Remove problematic empty fields that cause API errors
	// These fields cause "Object of type bytes is not JSON serializable" errors
	// when empty on the API backend
	fieldsToRemove := []string{"app_option", "paths"}
	for _, field := range fieldsToRemove {
		if val, exists := requestMap[field]; exists {
			// Remove if nil, empty object, or empty array
			switch v := val.(type) {
			case nil:
				delete(requestMap, field)
			case map[string]interface{}:
				if len(v) == 0 {
					delete(requestMap, field)
				}
			case []interface{}:
				if len(v) == 0 {
					delete(requestMap, field)
				}
			}
		}
	}

	// Also handle uribypass_header_value if it's null/empty
	if val, exists := requestMap["uribypass_header_value"]; exists {
		if val == nil || val == "" {
			delete(requestMap, "uribypass_header_value")
		}
	}

	// Marshal back to JSON
	modifiedBody, err := json.Marshal(requestMap)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal modified request: %v", err)
	}

	if privateAppRequestDebug {
		log.Printf("Modified body: %s", string(modifiedBody))
	}

	// Update request body
	req.Body = io.NopCloser(strings.NewReader(string(modifiedBody)))
	req.ContentLength = int64(len(modifiedBody))

	return req, nil
}
