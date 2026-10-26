package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// privateAppRequestHook removes problematic empty fields from private app PUT requests.
// The Netskope API has a bug where empty objects ({}) and arrays ([]) for certain fields
// are stored as bytes instead of JSON, causing SQL serialization errors.
// See docs/KNOWN_API_ISSUES.md for details.
type privateAppRequestHook struct{}

var _ beforeRequestHook = (*privateAppRequestHook)(nil)

func (i *privateAppRequestHook) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "updateNPAPrivateApp" {
		return req, nil
	}

	return rewriteRequestJSON(req, func(body []byte) ([]byte, error) {
		var requestMap map[string]interface{}
		if err := json.Unmarshal(body, &requestMap); err != nil {
			return nil, fmt.Errorf("unable to unmarshal request: %v", err)
		}

		// Remove app_option when clientless_access is false.
		// The API rejects app_option for non-clientless (regular client) apps with error:
		// "App_option is only available for non-web browser access"
		clientlessAccess, hasClientless := requestMap["clientless_access"].(bool)
		if !hasClientless || !clientlessAccess {
			delete(requestMap, "app_option")
		}

		// Remove problematic empty fields that cause API errors.
		// These fields cause "Object of type bytes is not JSON serializable" errors
		// when empty on the API backend.
		for _, field := range []string{"paths", "bypass_uris"} {
			if val, exists := requestMap[field]; exists {
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

		// Also remove empty app_option even for clientless apps (prevents serialization errors).
		if val, exists := requestMap["app_option"]; exists {
			switch v := val.(type) {
			case nil:
				delete(requestMap, "app_option")
			case map[string]interface{}:
				if len(v) == 0 {
					delete(requestMap, "app_option")
				}
			}
		}

		// Remove uribypass_header_value when null or empty.
		if val, exists := requestMap["uribypass_header_value"]; exists {
			if val == nil || val == "" {
				delete(requestMap, "uribypass_header_value")
			}
		}

		return json.Marshal(requestMap)
	})
}
