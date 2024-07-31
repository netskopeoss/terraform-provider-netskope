package hooks

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
	"log"
)

type MyAppResponse struct{
	data map[string]interface{}
}

var (
    _ afterSuccessHook = (*MyAppResponse)(nil)
)

func (i *MyAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	log.Print(hookCtx)
	log.Print("--------------------")
	log.Print(res.Body)
	if hookCtx.OperationID == "createNPAPrivateApps" || hookCtx.OperationID == "getNPAPrivateApp" {
        log.Print(hookCtx.OperationID)
        log.Print("inside the if statement")
		
		var responseMap MyAppResponse

        // Read and unmarshal the response body
        body, err := io.ReadAll(res.Body)
        if err != nil {
            return nil, fmt.Errorf("Error reading response body: %w", err)
        }

		log.Print("--------------------")
		log.Print(body)

        // defer res.Body.Close()

        // Unmarshal the raw response into a map
        if err := json.Unmarshal(body, &responseMap.data); err != nil {
            return nil, fmt.Errorf("failed to unmarshal response: %w", err)
        }

		log.Print("--------------------")
		log.Print(responseMap.data)

        // Modify the response as needed
        if val, ok := responseMap.data["app_name"]; ok {
            if strVal, isString := val.(string); isString {
                cleanedValue := strings.Trim(strVal, "[]")
                responseMap.data["app_name"] = cleanedValue
            }
        }

        // Marshal the modified response back to json.RawMessage
        modifiedBody, err := json.Marshal(responseMap.data)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal modified response: %w", err)
        }

        s := string(modifiedBody)
        res.Body = io.NopCloser(strings.NewReader(s))

        return res, nil
    }
	log.Print("outside the if statement")
    return res, nil
}