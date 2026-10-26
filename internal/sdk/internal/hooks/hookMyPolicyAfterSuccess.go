package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type myPolicyResponse struct{}

var _ afterSuccessHook = (*myPolicyResponse)(nil)

func (i *myPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createNPARules", "updateNPARules", "getNPARules", "NPARules":
	default:
		return res, nil
	}

	return rewriteResponseJSON(res, func(body []byte) ([]byte, error) {
		// Unmarshal into a generic map to avoid type conflicts in nested fields.
		// The API returns notify.templates as an array of objects on some rule types,
		// which cannot be unmarshalled into the generated [][]string model type.
		// Using map[string]interface{} handles any JSON shape without errors.
		var responseRaw map[string]interface{}
		if err := json.Unmarshal(body, &responseRaw); err != nil {
			return nil, fmt.Errorf("unable to unmarshal response: %v", err)
		}

		data, _ := responseRaw["data"].(map[string]interface{})
		if data == nil {
			return body, nil
		}

		ruleData, _ := data["rule_data"].(map[string]interface{})
		if ruleData == nil {
			return body, nil
		}

		// Transform private app names: the API wraps them in brackets on GET
		// responses (e.g. "[my-app]"). Strip the brackets to match the plain
		// names used in config and in create/update requests.
		if apps, ok := ruleData["privateApps"].([]interface{}); ok {
			for idx, app := range apps {
				if s, ok := app.(string); ok {
					apps[idx] = strings.Trim(s, "[]")
				}
			}
		}

		return json.Marshal(responseRaw)
	})
}
