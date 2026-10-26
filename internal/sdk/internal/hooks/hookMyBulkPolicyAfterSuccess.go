package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/internal/models"
)

type myBulkPolicyResponse struct {
	BulkPolicy []models.PolicyData `json:"data"`
	Status     string              `json:"status"`
}

var _ afterSuccessHook = (*myBulkPolicyResponse)(nil)

func (i *myBulkPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID != "getNPARulesList" {
		return res, nil
	}

	return rewriteResponseJSON(res, func(body []byte) ([]byte, error) {
		var responseMap myBulkPolicyResponse
		if err := json.Unmarshal(body, &responseMap); err != nil {
			return nil, fmt.Errorf("unable to unmarshal response: %v", err)
		}

		for i := range responseMap.BulkPolicy {
			old := responseMap.BulkPolicy[i].RuleData.PrivateApps
			responseMap.BulkPolicy[i].RuleData.PrivateApps = nil
			for _, app := range old {
				responseMap.BulkPolicy[i].RuleData.PrivateApps = append(
					responseMap.BulkPolicy[i].RuleData.PrivateApps,
					strings.Trim(app, "[]"),
				)
			}
		}

		return json.MarshalIndent(responseMap, "", "")
	})
}
