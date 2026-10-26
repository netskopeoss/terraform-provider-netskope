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

type myBulkAppResponse struct {
	BulkApps models.BulkAppData `json:"data"`
	Status   string             `json:"status"`
}

var _ afterSuccessHook = (*myBulkAppResponse)(nil)

func (i *myBulkAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID != "listNPAPrivateApps" {
		return res, nil
	}

	return rewriteResponseJSON(res, func(body []byte) ([]byte, error) {
		var responseMap myBulkAppResponse
		if err := json.Unmarshal(body, &responseMap); err != nil {
			return nil, fmt.Errorf("unable to unmarshal response: %v", err)
		}

		for i := range responseMap.BulkApps.AppData {
			app := &responseMap.BulkApps.AppData[i]

			// Copy app_id → id. The list endpoint returns "app_id" but the SDK
			// struct maps private_app_id from "id". Without this, private_app_id
			// is always 0 in the data source. See docs/bugs/BUG-012.
			if app.AppID != 0 && app.ID == 0 {
				app.ID = app.AppID
			}

			app.AppName = strings.Trim(app.AppName, "[]")

			// Copy transport to type for each protocol.
			// API inconsistency: POST/PUT requests use "type", but GET responses return "transport".
			// The Terraform schema expects "type", so we copy transport→type on read.
			// See docs/KNOWN_API_ISSUES.md #9 for details.
			for j := range app.Protocols {
				if app.Protocols[j].Type == "" && app.Protocols[j].Transport != "" {
					app.Protocols[j].Type = app.Protocols[j].Transport
				}
			}

			// Sort protocols by type (alphabetically) then port (numerically) to ensure
			// deterministic ordering. The API returns protocols in non-deterministic order.
			// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
			sort.Slice(app.Protocols, func(a, b int) bool {
				ta := app.Protocols[a].Type
				tb := app.Protocols[b].Type
				if ta != tb {
					return ta < tb
				}
				pa, errA := strconv.Atoi(app.Protocols[a].Port)
				pb, errB := strconv.Atoi(app.Protocols[b].Port)
				if errA == nil && errB == nil {
					return pa < pb
				}
				return app.Protocols[a].Port < app.Protocols[b].Port
			})

			// Convert publisher_id from integer to string (API returns int, SDK expects string)
			// and trim whitespace from publisher_name (API sometimes returns leading spaces).
			for j := range app.ServicePublisherAssignments {
				if app.ServicePublisherAssignments[j].PublisherID != nil {
					switch v := app.ServicePublisherAssignments[j].PublisherID.(type) {
					case float64:
						app.ServicePublisherAssignments[j].PublisherID = fmt.Sprintf("%.0f", v)
					case int:
						app.ServicePublisherAssignments[j].PublisherID = fmt.Sprintf("%d", v)
					}
				}
				app.ServicePublisherAssignments[j].PublisherName =
					strings.TrimSpace(app.ServicePublisherAssignments[j].PublisherName)
			}

			// Sort publishers by publisher_id to ensure deterministic ordering.
			// See docs/bugs/BUG-001-publishers-perpetual-diff.md for details.
			sort.Slice(app.ServicePublisherAssignments, func(a, b int) bool {
				idA := fmt.Sprintf("%v", app.ServicePublisherAssignments[a].PublisherID)
				idB := fmt.Sprintf("%v", app.ServicePublisherAssignments[b].PublisherID)
				numA, errA := strconv.Atoi(idA)
				numB, errB := strconv.Atoi(idB)
				if errA == nil && errB == nil {
					return numA < numB
				}
				return idA < idB
			})

			// Sort tags by tag_id to ensure deterministic ordering.
			sort.Slice(app.Tags, func(a, b int) bool {
				return app.Tags[a].TagID < app.Tags[b].TagID
			})

			// Populate label_ids from labels array (same as single app hook).
			if len(app.Labels) > 0 {
				labelIds := make([]string, 0, len(app.Labels))
				for _, label := range app.Labels {
					labelIds = append(labelIds, label.LabelID)
				}
				sort.Strings(labelIds)
				app.LabelIds = labelIds
			}
		}

		return json.MarshalIndent(responseMap, "", "")
	})
}
