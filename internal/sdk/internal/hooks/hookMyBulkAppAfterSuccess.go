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

type myBulkAppResponse struct {
	BulkApps models.BulkAppData `json:"data"`
	Status   string             `json:"status"`
}

var (
	_                      afterSuccessHook = (*myBulkAppResponse)(nil)
	myBulkAppResponseDebug bool             = false
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
			// Copy app_id → id. The list endpoint returns "app_id" but the SDK
			// struct maps private_app_id from "id". Without this, private_app_id
			// is always 0 in the data source. See docs/bugs/BUG-012.
			if responseMap.BulkApps.AppData[i].AppID != 0 && responseMap.BulkApps.AppData[i].ID == 0 {
				responseMap.BulkApps.AppData[i].ID = responseMap.BulkApps.AppData[i].AppID
			}

			oldAppNameValue := responseMap.BulkApps.AppData[i].AppName
			responseMap.BulkApps.AppData[i].AppName = strings.Trim(oldAppNameValue, "[]")

			// Copy transport to type for each protocol.
			// API inconsistency: POST/PUT requests use "type", but GET responses return "transport".
			// The Terraform schema expects "type", so we copy transport→type on read.
			// See docs/KNOWN_API_ISSUES.md #9 for details.
			for j := range responseMap.BulkApps.AppData[i].Protocols {
				if responseMap.BulkApps.AppData[i].Protocols[j].Type == "" && responseMap.BulkApps.AppData[i].Protocols[j].Transport != "" {
					responseMap.BulkApps.AppData[i].Protocols[j].Type = responseMap.BulkApps.AppData[i].Protocols[j].Transport
				}
			}

			// Sort protocols by type (alphabetically) then port (numerically) to ensure
			// deterministic ordering. The API returns protocols in non-deterministic order.
			// ModifyPlan handles config-vs-state order differences (BUG-002).
			// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
			sort.Slice(responseMap.BulkApps.AppData[i].Protocols, func(a, b int) bool {
				ta := responseMap.BulkApps.AppData[i].Protocols[a].Type
				tb := responseMap.BulkApps.AppData[i].Protocols[b].Type
				if ta != tb {
					return ta < tb
				}
				pa, errA := strconv.Atoi(responseMap.BulkApps.AppData[i].Protocols[a].Port)
				pb, errB := strconv.Atoi(responseMap.BulkApps.AppData[i].Protocols[b].Port)
				if errA == nil && errB == nil {
					return pa < pb
				}
				return responseMap.BulkApps.AppData[i].Protocols[a].Port < responseMap.BulkApps.AppData[i].Protocols[b].Port
			})

			// Convert publisher_id from integer to string (API returns int, SDK expects string)
			// and trim whitespace from publisher_name (API sometimes returns leading spaces)
			for j := range responseMap.BulkApps.AppData[i].ServicePublisherAssignments {
				if responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID != nil {
					switch v := responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID.(type) {
					case float64:
						responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID = fmt.Sprintf("%.0f", v)
						if myBulkAppResponseDebug {
							log.Printf("Converted publisher_id from float64 to string: %s", responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID)
						}
					case int:
						responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID = fmt.Sprintf("%d", v)
						if myBulkAppResponseDebug {
							log.Printf("Converted publisher_id from int to string: %s", responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherID)
						}
					}
				}
				responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherName = strings.TrimSpace(responseMap.BulkApps.AppData[i].ServicePublisherAssignments[j].PublisherName)
			}

			// Sort publishers by publisher_id to ensure deterministic ordering.
			// The API returns publishers in non-deterministic order, which causes
			// perpetual diffs because the schema uses ListNestedAttribute (order-sensitive).
			// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
			sort.Slice(responseMap.BulkApps.AppData[i].ServicePublisherAssignments, func(a, b int) bool {
				idA := fmt.Sprintf("%v", responseMap.BulkApps.AppData[i].ServicePublisherAssignments[a].PublisherID)
				idB := fmt.Sprintf("%v", responseMap.BulkApps.AppData[i].ServicePublisherAssignments[b].PublisherID)
				numA, errA := strconv.Atoi(idA)
				numB, errB := strconv.Atoi(idB)
				if errA == nil && errB == nil {
					return numA < numB
				}
				return idA < idB
			})

			// Sort tags by tag_id to ensure deterministic ordering.
			// The API returns tags in non-deterministic order, which causes
			// perpetual diffs because the schema uses ListNestedAttribute (order-sensitive).
			sort.Slice(responseMap.BulkApps.AppData[i].Tags, func(a, b int) bool {
				return responseMap.BulkApps.AppData[i].Tags[a].TagID < responseMap.BulkApps.AppData[i].Tags[b].TagID
			})

			// Populate label_ids from labels array (same as single app hook).
			if len(responseMap.BulkApps.AppData[i].Labels) > 0 {
				labelIds := make([]string, 0, len(responseMap.BulkApps.AppData[i].Labels))
				for _, label := range responseMap.BulkApps.AppData[i].Labels {
					labelIds = append(labelIds, label.LabelID)
				}
				sort.Strings(labelIds)
				responseMap.BulkApps.AppData[i].LabelIds = labelIds
			}

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
