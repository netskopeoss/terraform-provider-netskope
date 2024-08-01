package hooks
import (
    "bytes"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "strings"
    "github.com/netskope/terraform-provider-ns/internal/sdk/models/shared"
)
type MyAppResponse struct {
    PrivateAppsResponse *shared.PrivateAppsResponse
}
var (
    _ afterSuccessHook = (*MyAppResponse)(nil)
)
func (i *MyAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
    log.Print("AfterSucess hook....")
    log.Print("-------- before if  --------")
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
        // Unmarshal the raw response into a map
        if err := json.Unmarshal(body, &responseMap); err != nil {
            return nil, fmt.Errorf("failed to unmarshal response: %w", err)
        }
        log.Print("--------------------")
        log.Print(responseMap)
        oldValue := *responseMap.PrivateAppsResponse.Data.AppName
        newValue := strings.Trim(oldValue, "[]")
        responseMap.PrivateAppsResponse.Data.AppName = &(newValue)
        // Marshal the modified response back to json.RawMessage
        modifiedBody, err := json.Marshal(responseMap)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal modified response: %w", err)
        }
        res.Body = io.NopCloser(bytes.NewReader(modifiedBody))
        return res, nil
    }
    log.Print("outside the if statement")
    return res, nil
}