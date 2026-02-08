package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/internal/models"
)

type myAppResponse struct {
	Data   models.AppData `json:"data"`
	Status string         `json:"status"`
}

var (
	_ afterSuccessHook = (*myAppResponse)(nil)

	myAppResponseDebug = false
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

		// Copy transport to type for each protocol.
		// API inconsistency: POST/PUT requests use "type", but GET responses return "transport".
		// The Terraform schema expects "type", so we copy transport→type on read.
		// See docs/KNOWN_API_ISSUES.md #9 for details.
		for i := range responseMap.Data.Protocols {
			if responseMap.Data.Protocols[i].Type == "" && responseMap.Data.Protocols[i].Transport != "" {
				responseMap.Data.Protocols[i].Type = responseMap.Data.Protocols[i].Transport
				if myAppResponseDebug {
					log.Printf("Copied transport '%s' to type for protocol at index %d", responseMap.Data.Protocols[i].Transport, i)
				}
			}
		}

		// Sort protocols by type (alphabetically) then port (numerically) to ensure
		// deterministic ordering. The API returns protocols in non-deterministic order,
		// which causes perpetual drift because the schema uses ListNestedAttribute.
		// Sort order: tcp before udp (alphabetical), then port ascending (22, 443, 8080).
		// Users should configure HCL in this same order to avoid drift.
		// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
		sort.Slice(responseMap.Data.Protocols, func(i, j int) bool {
			ti := responseMap.Data.Protocols[i].Type
			tj := responseMap.Data.Protocols[j].Type
			if ti != tj {
				return ti < tj
			}
			pi, errI := strconv.Atoi(responseMap.Data.Protocols[i].Port)
			pj, errJ := strconv.Atoi(responseMap.Data.Protocols[j].Port)
			if errI == nil && errJ == nil {
				return pi < pj
			}
			return responseMap.Data.Protocols[i].Port < responseMap.Data.Protocols[j].Port
		})

		// Convert publisher_id from integer to string (API returns int, SDK expects string)
		// and trim whitespace from publisher_name (API sometimes returns leading spaces)
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
			responseMap.Data.ServicePublisherAssignments[i].PublisherName = strings.TrimSpace(responseMap.Data.ServicePublisherAssignments[i].PublisherName)
		}

		// Sort publishers by publisher_id to ensure deterministic ordering.
		// The API returns publishers in non-deterministic order, which causes
		// perpetual diffs because the schema uses ListNestedAttribute (order-sensitive).
		// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
		sort.Slice(responseMap.Data.ServicePublisherAssignments, func(i, j int) bool {
			idI := fmt.Sprintf("%v", responseMap.Data.ServicePublisherAssignments[i].PublisherID)
			idJ := fmt.Sprintf("%v", responseMap.Data.ServicePublisherAssignments[j].PublisherID)
			numI, errI := strconv.Atoi(idI)
			numJ, errJ := strconv.Atoi(idJ)
			if errI == nil && errJ == nil {
				return numI < numJ
			}
			return idI < idJ
		})

		// Sort tags by tag_id to ensure deterministic ordering.
		// The API returns tags in non-deterministic order, which causes
		// perpetual diffs because the schema uses ListNestedAttribute (order-sensitive).
		sort.Slice(responseMap.Data.Tags, func(i, j int) bool {
			return responseMap.Data.Tags[i].TagID < responseMap.Data.Tags[j].TagID
		})

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
