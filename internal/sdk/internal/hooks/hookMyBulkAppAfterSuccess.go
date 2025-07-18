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

type myBulkAppResponse struct {
	BulkApps models.BulkAppData `json:"data"`
	Status   string             `json:"status"`
}

var (
	_                      afterSuccessHook = (*myBulkAppResponse)(nil)
	myBulkAppResponseDebug bool             = true
)

func (i *myBulkAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID == "listNPAPrivateApps" {
		if myBulkAppResponseDebug {
			log.Print("Executing AfterSucess hook BULK APPS")
		}
		var responseMap myBulkAppResponse
		// Read and unmarshal the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read response body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read response body: %w", err)
		}
		if myBulkAppResponseDebug {
			log.Printf("SUCCESS: Successfully read response body")
		}
		// Unmarshal the raw response into a map
		if err := json.Unmarshal(body, &responseMap); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}
		if myBulkAppResponseDebug {
			log.Printf("SUCCESS: Successfully unmarshalled response")
			log.Print("--------------------")
			log.Print(responseMap)
			log.Print("--------------------")
		}
		for i := range responseMap.BulkApps.AppData {
			oldAppNameValue := responseMap.BulkApps.AppData[i].AppName
			responseMap.BulkApps.AppData[i].AppName = strings.Trim(oldAppNameValue, "[]")
			if myBulkAppResponseDebug {
				log.Print("--------------------")
				log.Print(responseMap.BulkApps.AppData[i].AppName)
				log.Print("--------------------")
				log.Print(responseMap.BulkApps.AppData[i])
			}
		}
		modifiedBody, err := json.MarshalIndent(responseMap, "", "")
		if myBulkAppResponseDebug {
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
