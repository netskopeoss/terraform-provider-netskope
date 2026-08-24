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
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "updateNPARules" ||
		hookCtx.OperationID == "getNPARules" || hookCtx.OperationID == "NPARules" {
		log.Print("Executing AfterSuccess myPolicyResponse hook....")
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

		// Defensive nil check to prevent crash if RuleData is nil
		if responseMap.Data.RuleData == nil {
			log.Print("WARNING: RuleData is nil, skipping transformation")
			// Restore body and return without modification
			res.Body = io.NopCloser(strings.NewReader(string(body)))
			return res, nil
		}

		// Restore template display name from cache.
		//
		// The Netskope API accepts display names (e.g. "tf-test-template") on
		// create/update but returns .html file names (e.g. "2.html") on GET.
		// Without correction, state would store the file name, causing a perpetual
		// diff against the display name in the user's config. suppressTemplateDrift
		// (nparules_resource_planmodify.go) suppresses the visual diff but still
		// causes update payloads to contain the file name, which the API then rejects
		// with "Undefined template: *.html" (block) or "template field is required"
		// (periodic_reauth, which strips nil'd fields differently).
		//
		// Fix: during createNPARules, BeforeRequest stored the display name in the
		// request context. Here we observe the file name in the API response, cache
		// the file→display mapping, and substitute the display name back into the
		// response. Subsequent getNPARules/updateNPARules responses are fixed using
		// the same cache lookup. State then always holds the display name, so:
		//   - Plan: config == state → no drift → suppressTemplateDrift is a no-op
		//   - Update payload: display name → hook doesn't strip → API accepts ✓
		//
		// See docs/bugs/BUG-019-block-rule-template-phantom-update.md
		// See https://github.com/netskopeoss/terraform-provider-netskope/issues/116
		if responseMap.Data.RuleData.MatchCriteriaAction != nil &&
			responseMap.Data.RuleData.MatchCriteriaAction.Template != nil {

			fileName := *responseMap.Data.RuleData.MatchCriteriaAction.Template
			if strings.HasSuffix(fileName, ".html") {
				// For createNPARules: BeforeRequest stored the display name in the
				// request context. Populate the cache with the observed mapping.
				if hookCtx.OperationID == "createNPARules" && res.Request != nil {
					if displayName, ok := npaTemplateDisplayNameFromCtx(res.Request.Context()); ok {
						npaTemplateCacheSet(fileName, displayName)
					}
				}
				// Substitute the cached display name into the response so that
				// Terraform state stores the display name rather than the file name.
				if displayName, ok := npaTemplateCacheGet(fileName); ok {
					dn := displayName
					responseMap.Data.RuleData.MatchCriteriaAction.Template = &dn
				}
			}
		}

		// Transform private app names: the API wraps them in brackets on GET
		// responses (e.g. "[my-app]"). Strip the brackets to match the plain
		// names used in config and in create/update requests.
		if responseMap.Data.RuleData.PrivateApps != nil {
			oldPrivateAppValue := responseMap.Data.RuleData.PrivateApps
			responseMap.Data.RuleData.PrivateApps = nil
			for _, untrimmedApp := range oldPrivateAppValue {
				trimmed := strings.Trim(untrimmedApp, "[]")
				responseMap.Data.RuleData.PrivateApps = append(responseMap.Data.RuleData.PrivateApps, trimmed)
			}
		}

		// Marshal the modified response back to JSON
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
