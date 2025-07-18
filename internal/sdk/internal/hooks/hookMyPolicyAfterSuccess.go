package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/internal/models"
)

type myPolicyResponse struct {
	Data   models.PolicyData `json:"data"`
	Status string            `json:"status"`
}

var (
	_ afterSuccessHook = (*myPolicyResponse)(nil)
)

func (i *myPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "getNPARules" || hookCtx.OperationID == "NPARules" {
		log.Print("Executing AfterSucess myPolicyResponse hook....")
		var responseMap myPolicyResponse
		// Read and unmarshal the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read response body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read response body: %w", err)
		}
		log.Printf("SUCCESS: Successfully read response body")
		// Unmarshal the raw response into a map
		if err := json.Unmarshal(body, &responseMap); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}
		log.Printf("SUCCESS: Successfully unmarshalled response")
		log.Print("--------------------")
		log.Print(responseMap)
		log.Print("--------------------")
		oldPrivateAppValue := responseMap.Data.RuleData.PrivateApps
		responseMap.Data.RuleData.PrivateApps = nil
		for _, untrimmedApp := range oldPrivateAppValue {
			trimmed := strings.Trim(untrimmedApp, "[]")
			responseMap.Data.RuleData.PrivateApps = append(responseMap.Data.RuleData.PrivateApps, trimmed)
		}
		// responseMap.Data.AppName = strings.Trim(oldValue, "[]")
		// Marshal the modified response back to json.RawMessage
		modifiedBody, err := json.MarshalIndent(responseMap, "", "")
		if err != nil {
			log.Printf("Error: Unable to marshal modified response: %v", err)
			return nil, fmt.Errorf("Error: Unable to marshal modified response: %v", err)
		}
		s := string(modifiedBody)
		res.Body = io.NopCloser(strings.NewReader(s))
		return res, nil
	}
	return res, nil
}
