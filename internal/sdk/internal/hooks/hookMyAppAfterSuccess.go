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

type myAppResponse struct {
	Data   models.AppData `json:"data"`
	Status string         `json:"status"`
}

var (
	_                  afterSuccessHook = (*myAppResponse)(nil)
	myAppResponseDebug bool             = false
)

func (i *myAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID == "createNPAPrivateApps" || hookCtx.OperationID == "getNPAPrivateApp" || hookCtx.OperationID == "updateNPAPrivateApp" {
		if myAppResponseDebug {
			log.Print("Executing AfterSucess single app hook....")
		}
		var responseMap myAppResponse
		// Read and unmarshal the response body
		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read response body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read response body: %w", err)
		}
		if myAppResponseDebug {
			log.Printf("SUCCESS: Successfully read response body")
		}
		// Unmarshal the raw response into a map
		if err := json.Unmarshal(body, &responseMap); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}
		if myAppResponseDebug {
			log.Printf("SUCCESS: Successfully unmarshalled response")
			log.Print("--------------------")
			log.Print(responseMap)
			log.Print("--------------------")
		}
		oldAppNameValue := responseMap.Data.AppName
		responseMap.Data.AppName = strings.Trim(oldAppNameValue, "[]")

		// Copy transport to type for each protocol (API returns transport but schema uses type)
		for i := range responseMap.Data.Protocols {
			if responseMap.Data.Protocols[i].Type == "" && responseMap.Data.Protocols[i].Transport != "" {
				responseMap.Data.Protocols[i].Type = responseMap.Data.Protocols[i].Transport
				if myAppResponseDebug {
					log.Printf("Copied transport '%s' to type for protocol at index %d", responseMap.Data.Protocols[i].Transport, i)
				}
			}
		}

		// Convert publisher_id from integer to string (API returns int, SDK expects string)
		for i := range responseMap.Data.ServicePublisherAssignments {
			if responseMap.Data.ServicePublisherAssignments[i].PublisherID != nil {
				switch v := responseMap.Data.ServicePublisherAssignments[i].PublisherID.(type) {
				case float64:
					responseMap.Data.ServicePublisherAssignments[i].PublisherID = fmt.Sprintf("%.0f", v)
					if myAppResponseDebug {
						log.Printf("Converted publisher_id from float64 to string: %s", responseMap.Data.ServicePublisherAssignments[i].PublisherID)
					}
				case int:
					responseMap.Data.ServicePublisherAssignments[i].PublisherID = fmt.Sprintf("%d", v)
					if myAppResponseDebug {
						log.Printf("Converted publisher_id from int to string: %s", responseMap.Data.ServicePublisherAssignments[i].PublisherID)
					}
				}
			}
		}

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
