package hooks

import (
	"encoding/json"
	"fmt"
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

var _ afterSuccessHook = (*myAppResponse)(nil)

func (i *myAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	switch hookCtx.OperationID {
	case "createNPAPrivateApps", "getNPAPrivateApp", "updateNPAPrivateApp":
	default:
		return res, nil
	}

	return rewriteResponseJSON(res, func(body []byte) ([]byte, error) {
		var responseMap myAppResponse
		if err := json.Unmarshal(body, &responseMap); err != nil {
			return nil, fmt.Errorf("unable to unmarshal response: %v", err)
		}

		responseMap.Data.AppName = strings.Trim(responseMap.Data.AppName, "[]")

		// Copy transport to type for each protocol.
		// API inconsistency: POST/PUT requests use "type", but GET responses return "transport".
		// The Terraform schema expects "type", so we copy transport→type on read.
		// See docs/KNOWN_API_ISSUES.md #9 for details.
		for i := range responseMap.Data.Protocols {
			if responseMap.Data.Protocols[i].Type == "" && responseMap.Data.Protocols[i].Transport != "" {
				responseMap.Data.Protocols[i].Type = responseMap.Data.Protocols[i].Transport
			}
		}

		// Sort protocols by type (alphabetically) then port (numerically) to ensure
		// deterministic ordering. The API returns protocols in non-deterministic order.
		// ModifyPlan handles config-vs-state order differences (BUG-002).
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
		// and trim whitespace from publisher_name (API sometimes returns leading spaces).
		for i := range responseMap.Data.ServicePublisherAssignments {
			if responseMap.Data.ServicePublisherAssignments[i].PublisherID != nil {
				switch v := responseMap.Data.ServicePublisherAssignments[i].PublisherID.(type) {
				case float64:
					responseMap.Data.ServicePublisherAssignments[i].PublisherID = fmt.Sprintf("%.0f", v)
				case int:
					responseMap.Data.ServicePublisherAssignments[i].PublisherID = fmt.Sprintf("%d", v)
				}
			}
			responseMap.Data.ServicePublisherAssignments[i].PublisherName =
				strings.TrimSpace(responseMap.Data.ServicePublisherAssignments[i].PublisherName)
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
		sort.Slice(responseMap.Data.Tags, func(i, j int) bool {
			return responseMap.Data.Tags[i].TagID < responseMap.Data.Tags[j].TagID
		})

		// Populate label_ids from labels array.
		// The API returns labels as [{label_id, permission}] but the Terraform schema
		// uses label_ids (a flat string array) for input. We extract the IDs so the
		// read state matches what was configured.
		if len(responseMap.Data.Labels) > 0 {
			labelIds := make([]string, 0, len(responseMap.Data.Labels))
			for _, label := range responseMap.Data.Labels {
				labelIds = append(labelIds, label.LabelID)
			}
			sort.Strings(labelIds)
			responseMap.Data.LabelIds = labelIds
		}

		return json.MarshalIndent(responseMap, "", "")
	})
}
