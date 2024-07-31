package hooks

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
)

type MyAppResponse struct{}

var (
    _ afterSuccessHook = (*MyAppResponse)(nil)
)

func (i *MyAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
    if hookCtx.OperationID == "createNPAPrivateApps" || hookCtx.OperationID == "getNPAPrivateApp" {
        var responseMap map[string]interface{}

        // Read and unmarshal the response body
        body, err := io.ReadAll(res.Body)
        if err != nil {
            return nil, fmt.Errorf("Error reading response body: %w", err)
        }

        defer res.Body.Close()

        // Unmarshal the raw response into a map
        if err := json.Unmarshal(body, &responseMap); err != nil {
            return nil, fmt.Errorf("failed to unmarshal response: %w", err)
        }

        // Modify the response as needed
        if val, ok := responseMap["app_name"]; ok {
            if strVal, isString := val.(string); isString {
                cleanedValue := strings.Trim(strVal, "[]")
                responseMap["app_name"] = cleanedValue
            }
        }

        // Marshal the modified response back to json.RawMessage
        modifiedBody, err := json.Marshal(responseMap)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal modified response: %w", err)
        }

        s := string(modifiedBody)
        res.Body = io.NopCloser(strings.NewReader(s))

        return res, nil
    }
    return res, nil
}