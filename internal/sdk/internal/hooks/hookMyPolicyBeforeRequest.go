package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

//	type myPolicyRequest struct {
//		Data models.PolicyData `json:"data"`
//	}
type myPolicyRequest struct {
	Enabled    *string    `json:"enabled,omitempty"`
	ModifyBy   *string    `json:"modify_by,omitempty"`
	ModifyTime *string    `json:"modify_time,omitempty"`
	ModifyType *string    `json:"modify_type,omitempty"`
	PolicyType *string    `json:"policy_type,omitempty"`
	GroupID    *string    `json:"group_id,omitempty"`
	RuleData   *RuleData  `json:"rule_data,omitempty"`
	RuleID     *string    `json:"rule_id,omitempty"`
	RuleName   *string    `json:"rule_name,omitempty"`
	RuleOrder  *RuleOrder `json:"rule_order,omitempty"`
}

type RuleData struct {
	AccessMethod              []string                    `json:"access_method,omitempty"`
	BNegateNetLocation        *bool                       `json:"b_negateNetLocation,omitempty"`
	BNegateSrcCountries       *bool                       `json:"b_negateSrcCountries,omitempty"`
	Classification            *string                     `json:"classification,omitempty"`
	DlpActions                []NpaPolicyRuleDlp          `json:"dlp_actions,omitempty"`
	TssActions                []NpaPolicyRuleTss          `json:"tss_actions,omitempty"`
	TssProfile                []string                    `json:"tss_profile,omitempty"`
	ExternalDlp               *bool                       `json:"external_dlp,omitempty"`
	JSONVersion               *int64                      `json:"json_version,omitempty"`
	DeviceClassificationID    []int64                     `json:"device_classification_id,omitempty"`
	MatchCriteriaAction       *MatchCriteriaAction        `json:"match_criteria_action,omitempty"`
	NetLocationObj            []string                    `json:"net_location_obj,omitempty"`
	OrganizationUnits         []string                    `json:"organization_units,omitempty"`
	PrivateAppTagIds          []string                    `json:"privateAppTagIds,omitempty"`
	PrivateAppTags            []string                    `json:"privateAppTags,omitempty"`
	PrivateApps               []string                    `json:"privateApps,omitempty"`
	PrivateAppsWithActivities []PrivateAppsWithActivities `json:"privateAppsWithActivities,omitempty"`
	ShowDlpProfileActionTable *bool                       `json:"show_dlp_profile_action_table,omitempty"`
	SrcCountries              []string                    `json:"srcCountries,omitempty"`
	UserGroups                []string                    `json:"userGroups,omitempty"`
	UserType                  *string                     `json:"userType,omitempty"`
	Users                     []string                    `json:"users,omitempty"`
	Version                   *int64                      `json:"version,omitempty"`
}

type RuleOrder struct {
	Order    *string `json:"order"`
	Position *int64  `json:"position"`
	RuleID   *string `json:"rule_id"`
	RuleName *string `json:"rule_name"`
}

type NpaPolicyRuleDlp struct {
	Actions    []string `json:"actions"`
	DlpProfile *string  `json:"dlp_profile"`
}

type MatchCriteriaAction struct {
	ActionName *string `json:"action_name"`
}

type PrivateAppsWithActivities struct {
	Activities []Activities `json:"activities"`
	AppID      []string     `json:"app_id"`
	AppName    *string      `json:"app_name"`
}

type NpaPolicyRuleTss struct {
	Actions    []NpaPolicyRuleTssActions `json:"actions"`
	TssProfile []string                  `json:"tss_profile"`
}

type Activities struct {
	Activity          *string  `json:"activity"`
	ListOfConstraints []string `json:"list_of_constraints"`
}

type NpaPolicyRuleTssActions struct {
	ActionName         string `json:"action_name"`
	RemediationProfile string `json:"remediation_profile"`
	Severity           string `json:"severity"`
	Template           string `json:"template"`
}

var (
	_                    beforeRequestHook = (*myPolicyRequest)(nil)
	myPolicyRequestDebug bool              = true
)

func (i *myPolicyRequest) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "updateNPARulesById" {
		if myPolicyRequestDebug {
			log.Print("Executing BeforeRequest hook....")
		}
		var requestMap myPolicyRequest
		// Read and unmarshal the response body
		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Printf("ERROR: Unable to read request body: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to read request body: %w", err)
		}
		if myPolicyRequestDebug {
			log.Printf("SUCCESS: Successfully read response body")
		}
		// Unmarshal the raw response into a map
		if myPolicyRequestDebug {
			log.Printf("--------Body--------")
			log.Println(string(body))
			log.Printf("--------------------")
		}
		if err := json.Unmarshal(body, &requestMap); err != nil {
			log.Printf("ERROR: Unable to unmarshal response: %v", err)
			return nil, fmt.Errorf("ERROR: Unable to unmarshal response: %v", err)
		}

		if myPolicyRequestDebug {
			log.Printf("SUCCESS: Successfully unmarshalled response")
			log.Print(requestMap)
			log.Print("--------------------")

		}

		oldPrivateAppValue := requestMap.RuleData.PrivateApps
		requestMap.RuleData.PrivateApps = nil
		for _, trimmedApp := range oldPrivateAppValue {
			untrimmedApp := "[" + trimmedApp + "]"
			requestMap.RuleData.PrivateApps = append(requestMap.RuleData.PrivateApps, untrimmedApp)
		}
		//modifiedBody, err := json.MarshalIndent(requestMap, "", "")
		modifiedBody, err := json.Marshal(requestMap)
		if err != nil {
			log.Printf("Error: Unable to marshal modified response: %v", err)
			return nil, fmt.Errorf("Error: Unable to marshal modified response: %v", err)
		}
		if myPolicyRequestDebug {
			log.Print("=======exit==========")
			log.Println(string(modifiedBody))
			log.Print("=================")
		}
		if myPolicyRequestDebug {
			log.Printf("Modified body length: %d", len(modifiedBody))
			log.Printf("Request Content-Length header: %d", req.ContentLength)
		}
		s := string(modifiedBody)
		req.Body = io.NopCloser(strings.NewReader(s))
		req.ContentLength = int64(len(modifiedBody))
		return req, nil
	}
	return req, nil
}
