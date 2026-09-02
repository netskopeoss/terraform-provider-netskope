package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type myPolicyResponse struct{}

var (
	_ afterSuccessHook = (*myPolicyResponse)(nil)
)

func (i *myPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "updateNPARules" ||
		hookCtx.OperationID == "getNPARules" || hookCtx.OperationID == "NPARules" {
		log.Print("Executing AfterSuccess myPolicyResponse hook....")

		body, err := io.ReadAll(res.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read response body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read response body: %w", err)
		}
		log.Printf("SUCCESS: Successfully read response body")

		// Unmarshal into a generic map to avoid type conflicts in nested fields.
		// The API returns notify.templates as an array of objects on some rule types,
		// which cannot be unmarshalled into the generated [][]string model type.
		// Using map[string]interface{} handles any JSON shape without errors.
		var responseRaw map[string]interface{}
		if err := json.Unmarshal(body, &responseRaw); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}
		log.Printf("SUCCESS: Successfully unmarshalled response")

		data, _ := responseRaw["data"].(map[string]interface{})
		if data == nil {
			log.Print("WARNING: data is nil, skipping transformation")
			res.Body = io.NopCloser(strings.NewReader(string(body)))
			return res, nil
		}

		ruleData, _ := data["rule_data"].(map[string]interface{})
		if ruleData == nil {
			log.Print("WARNING: RuleData is nil, skipping transformation")
			res.Body = io.NopCloser(strings.NewReader(string(body)))
			return res, nil
		}

		// Restore template display name from the templates API cache or (for block
		// rules) the session-based fallback cache.
		//
		// Primary path (all action types): look up the .html filename in the templates
		// API bidirectional cache (npaTemplatesAPICache). This cache is populated on
		// first use from /api/v2/templates/usernotifications and covers all action types,
		// including periodic_reauth. When a match is found, the display name is
		// substituted so Terraform state stores the display name rather than the file
		// name — eliminating drift for both block and periodic_reauth rules.
		//
		// Fallback path (block rules only): if the templates API cache has no entry,
		// use the session-based npaTemplateCache populated by BeforeRequest context
		// storage on create. This preserves backward compatibility for tenants where
		// the templates API is unavailable.
		//
		// See docs/bugs/BUG-019-block-rule-template-phantom-update.md
		// See docs/bugs/BUG-020-periodic-reauth-template.md
		// See https://github.com/netskopeoss/terraform-provider-netskope/issues/118
		if mca, ok := ruleData["match_criteria_action"].(map[string]interface{}); ok {
			if fileName, ok := mca["template"].(string); ok && strings.HasSuffix(fileName, ".html") {
				actionName, _ := mca["action_name"].(string)

				// Ensure the templates API cache is populated (lazy, once per session).
				if res.Request != nil {
					ensureNPATemplatesAPILoaded(extractAPIBase(res.Request), res.Request.Header.Get("Netskope-Api-Token"))
				}

				if displayName, ok := npaTemplatesAPIDisplayName(actionName, fileName); ok {
					// Primary: templates API cache covers all action types.
					mca["template"] = displayName
				} else if actionName == "block" {
					// Fallback: session cache for block rules. Populated by BeforeRequest
					// context storage during createNPARules when the templates API was
					// unavailable (user supplied a display name that was stored in ctx).
					if hookCtx.OperationID == "createNPARules" && res.Request != nil {
						if displayName, ok := npaTemplateDisplayNameFromCtx(res.Request.Context()); ok {
							npaTemplateCacheSet(fileName, displayName)
						}
					}
					if displayName, ok := npaTemplateCacheGet(fileName); ok {
						mca["template"] = displayName
					}
				}
			}
		}

		// Transform private app names: the API wraps them in brackets on GET
		// responses (e.g. "[my-app]"). Strip the brackets to match the plain
		// names used in config and in create/update requests.
		if apps, ok := ruleData["privateApps"].([]interface{}); ok {
			for idx, app := range apps {
				if s, ok := app.(string); ok {
					apps[idx] = strings.Trim(s, "[]")
				}
			}
		}

		modifiedBody, err := json.Marshal(responseRaw)
		if err != nil {
			log.Printf("Error: Unable to marshal modified response: %v", err)
			return nil, fmt.Errorf("Error: Unable to marshal modified response: %v", err)
		}
		res.Body = io.NopCloser(strings.NewReader(string(modifiedBody)))
		return res, nil
	}
	return res, nil
}
