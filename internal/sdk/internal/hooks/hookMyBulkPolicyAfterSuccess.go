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

type myBulkPolicyResponse struct {
	BulkPolicy []models.PolicyData `json:"data"`
	Status     string              `json:"status"`
}

var (
	_                 afterSuccessHook = (*myBulkPolicyResponse)(nil)
	myBulkPolicyDebug bool             = true
)

func (i *myBulkPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if myBulkPolicyDebug {
		log.Print("Executing AfterSuccess myBulkPolicyResponse hook....")
	}
	if hookCtx.OperationID == "getNPARulesList" {
		var responseMap myBulkPolicyResponse
		// Read and unmarshal the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read response body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read response body: %w", err)
		}
		if myBulkPolicyDebug {
			log.Printf("SUCCESS: Successfully read response body")
		}
		// Unmarshal the raw response into a map
		if err := json.Unmarshal(body, &responseMap); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}
		if myBulkPolicyDebug {
			log.Printf("SUCCESS: Successfully unmarshalled response")
			log.Print("--------------------")
			log.Print(responseMap)
			log.Print("--------------------")
		}
		for i := range responseMap.BulkPolicy {
			oldPrivateAppValue := responseMap.BulkPolicy[i].RuleData.PrivateApps
			responseMap.BulkPolicy[i].RuleData.PrivateApps = nil
			for _, untrimmedApp := range oldPrivateAppValue {
				trimmed := strings.Trim(untrimmedApp, "[]")
				responseMap.BulkPolicy[i].RuleData.PrivateApps = append(responseMap.BulkPolicy[i].RuleData.PrivateApps, trimmed)
			}
		}
		modifiedBody, err := json.MarshalIndent(responseMap, "", "")
		if myBulkPolicyDebug {
			log.Print("=================")
			log.Println(string(modifiedBody))
			log.Print("=================")
		}

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
