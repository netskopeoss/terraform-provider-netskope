package hooks

import (
	"net/http"
	"strings"
	"encoding/json"
	"fmt"
	"context"


)

type MyAppResponse struct{}

var (
	_ afterSuccessHook  = (*MyAppResponse)(nil)
)

func (i *MyAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	var responseMap map[string]interface{}

	// Unmarshal the raw response
	if err := json.Unmarshal(*resp, &responseMap); err != nil {
		return fm.Error("failed to unmarshal response: %w", err)
	}

	// Modify the response
	if val, ok := responseMap["app_name"]; ok {
		if strVal, isString := val.(string); {
			cleanedValue := strings.Trim(strVal, "[]")
			responseMap["app_name"] = cleanedValue
		}
	}

	// Marshal the modified response back to the json Raw Message
	modifiedResp, err := json.Marshal(responseMap)
	of err != nil {
		return fmt.Error("failed to marshal modified response: %w", err)
	}

	// Set the modified response back to the ResourceData
	if err := data.Set("modified_response", string(modifiedResp)); err != nil {
		return fmt.Error("failed to set modified response: %w", err)
	}
	
	return res, nil
}
