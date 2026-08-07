package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
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
	AccessMethod                []string                    `json:"access_method,omitempty"`
	BNegateNetLocation          *bool                       `json:"b_negateNetLocation,omitempty"`
	BNegateSrcCountries         *bool                       `json:"b_negateSrcCountries,omitempty"`
	Classification              []string                    `json:"classification,omitempty"` // See docs/bugs/BUG-018: API returns array, not string
	Description                 *string                     `json:"description,omitempty"`
	DeviceClassificationID      []string                    `json:"device_classification_id,omitempty"`
	DlpActions                  []NpaPolicyRuleDlp          `json:"dlp_actions,omitempty"`
	ExternalDlp                 *bool                       `json:"external_dlp,omitempty"`
	JSONVersion                 *int64                      `json:"json_version,omitempty"`
	MatchCriteriaAction         *MatchCriteriaAction        `json:"match_criteria_action,omitempty"`
	NetLocationObj              []string                    `json:"net_location_obj,omitempty"`
	Notify                      *Notify                     `json:"notify,omitempty"`
	OrganizationUnits           []string                    `json:"organization_units,omitempty"`
	Os                          []string                    `json:"os,omitempty"`
	PeriodicReauth              *PeriodicReauth             `json:"periodic_reauth,omitempty"`
	PrivateAppTagIds            []string                    `json:"privateAppTagIds,omitempty"`
	PrivateAppTags              []string                    `json:"privateAppTags,omitempty"`
	PrivateApps                 []string                    `json:"privateApps,omitempty"`
	PrivateAppsWithActivities   []PrivateAppsWithActivities `json:"privateAppsWithActivities,omitempty"`
	Schedule                    []ScheduleItem              `json:"schedule,omitempty"`
	ShowDlpProfileActionTable   *bool                       `json:"show_dlp_profile_action_table,omitempty"`
	SrcCountries                []string                    `json:"srcCountries,omitempty"`
	TssActions                  []NpaPolicyRuleTss          `json:"tss_actions,omitempty"`
	TssProfile                  []string                    `json:"tss_profile,omitempty"`
	UserConfidence              *UserConfidence             `json:"user_confidence,omitempty"`
	UserGroups                  []string                    `json:"userGroups,omitempty"`
	UserType                    *string                     `json:"userType,omitempty"`
	Users                       []string                    `json:"users,omitempty"`
	Version                     *int64                      `json:"version,omitempty"`
}

type RuleOrder struct {
	Order    *string `json:"order,omitempty"`
	Position *int64  `json:"position,omitempty"`
	RuleID   *int64  `json:"rule_id,omitempty"`
	RuleName *string `json:"rule_name,omitempty"`
}

type Notify struct {
	Emails    []string   `json:"emails,omitempty"`
	FromUser  *string    `json:"from_user,omitempty"`
	Interval  *string    `json:"interval,omitempty"`
	Templates [][]string `json:"templates,omitempty"`
	ToUsers   []string   `json:"to_users,omitempty"`
}

type PeriodicReauth struct {
	ReauthInterval     *string `json:"reauth_interval,omitempty"`
	ReauthIntervalUnit *string `json:"reauth_interval_unit,omitempty"`
}

type ScheduleItem struct {
	TimeIntervalObj []string    `json:"time_interval_obj,omitempty"`
	TimeRange       []TimeRange `json:"time_range,omitempty"`
}

type TimeRange struct {
	EndDate   *string `json:"end_date,omitempty"`
	EndTime   *string `json:"end_time,omitempty"`
	StartDate *string `json:"start_date,omitempty"`
	StartTime *string `json:"start_time,omitempty"`
}

type UserConfidence struct {
	Index    *string `json:"index,omitempty"`
	Operator *string `json:"operator,omitempty"`
}

type NpaPolicyRuleDlp struct {
	Actions    []string `json:"actions"`
	DlpProfile *string  `json:"dlp_profile"`
}

type MatchCriteriaAction struct {
	ActionName *string `json:"action_name"`
	EmitAlert  *bool   `json:"emit_alert,omitempty"`
	Template   *string `json:"template,omitempty"`
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
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "updateNPARules" {
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

		if requestMap.RuleData != nil {
			oldPrivateAppValue := requestMap.RuleData.PrivateApps
			requestMap.RuleData.PrivateApps = nil
			for _, trimmedApp := range oldPrivateAppValue {
				untrimmedApp := "[" + trimmedApp + "]"
				requestMap.RuleData.PrivateApps = append(requestMap.RuleData.PrivateApps, untrimmedApp)
			}

			// Strip negate flags when their associated criteria lists are empty.
			// The API rejects b_negateNetLocation / b_negateSrcCountries even set to
			// false on tenants where the source-IP-criteria feature flag is disabled.
			// The flags are meaningless without a populated criteria list anyway.
			if len(requestMap.RuleData.NetLocationObj) == 0 {
				requestMap.RuleData.BNegateNetLocation = nil
			}
			if len(requestMap.RuleData.SrcCountries) == 0 {
				requestMap.RuleData.BNegateSrcCountries = nil
			}

			// Strip .html file names from update payloads.
			// suppressTemplateDrift (in nparules_resource_planmodify.go) sets the
			// planned template to the state value (a .html file name) to suppress
			// the display-name vs file-name diff. When a legitimate update fires for
			// any reason, the PUT body would otherwise contain the .html file name
			// which the API rejects with "Undefined template: *.html".
			// Omitting template on update is safe — the API preserves the existing one.
			// Create payloads are left unchanged (display names work on create).
			// See docs/bugs/BUG-019-block-rule-template-phantom-update.md
			if hookCtx.OperationID == "updateNPARules" &&
				requestMap.RuleData.MatchCriteriaAction != nil &&
				requestMap.RuleData.MatchCriteriaAction.Template != nil &&
				strings.HasSuffix(*requestMap.RuleData.MatchCriteriaAction.Template, ".html") {
				requestMap.RuleData.MatchCriteriaAction.Template = nil
			}
		}

		// Convert device_classification_id from []string to []int64 for the API
		// The API returns strings but expects integers on write
		var deviceClassificationIDInts []int64
		if requestMap.RuleData != nil {
			for _, s := range requestMap.RuleData.DeviceClassificationID {
				n, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					log.Printf("ERROR: Unable to convert device_classification_id value %q to int: %v", s, err)
					return nil, fmt.Errorf("ERROR: Unable to convert device_classification_id value %q to int: %w", s, err)
				}
				deviceClassificationIDInts = append(deviceClassificationIDInts, n)
			}
		}

		// Marshal using a temporary wrapper that outputs device_classification_id as []int64
		type ruleDataWithIntIDs struct {
			RuleData
			DeviceClassificationID []int64 `json:"device_classification_id,omitempty"`
		}
		type requestWithIntIDs struct {
			Enabled    *string              `json:"enabled,omitempty"`
			ModifyBy   *string              `json:"modify_by,omitempty"`
			ModifyTime *string              `json:"modify_time,omitempty"`
			ModifyType *string              `json:"modify_type,omitempty"`
			PolicyType *string              `json:"policy_type,omitempty"`
			GroupID    *string              `json:"group_id,omitempty"`
			RuleData   *ruleDataWithIntIDs  `json:"rule_data,omitempty"`
			RuleID     *string              `json:"rule_id,omitempty"`
			RuleName   *string              `json:"rule_name,omitempty"`
			RuleOrder  *RuleOrder           `json:"rule_order,omitempty"`
		}

		marshalReq := requestWithIntIDs{
			Enabled:    requestMap.Enabled,
			ModifyBy:   requestMap.ModifyBy,
			ModifyTime: requestMap.ModifyTime,
			ModifyType: requestMap.ModifyType,
			PolicyType: requestMap.PolicyType,
			GroupID:    requestMap.GroupID,
			RuleID:     requestMap.RuleID,
			RuleName:   requestMap.RuleName,
			RuleOrder:  requestMap.RuleOrder,
		}
		if requestMap.RuleData != nil {
			rd := *requestMap.RuleData
			rd.DeviceClassificationID = nil // Clear string version so embedded field doesn't conflict
			marshalReq.RuleData = &ruleDataWithIntIDs{
				RuleData:               rd,
				DeviceClassificationID: deviceClassificationIDInts,
			}
		}

		modifiedBody, err := json.Marshal(marshalReq)
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
