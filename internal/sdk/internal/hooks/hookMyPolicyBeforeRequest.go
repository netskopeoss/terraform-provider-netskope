package hooks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
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
	Classification            []string                    `json:"classification,omitempty"` // See docs/bugs/BUG-018: API returns array, not string
	Description               *string                     `json:"description,omitempty"`
	DeviceClassificationID    []string                    `json:"device_classification_id,omitempty"`
	DlpActions                []NpaPolicyRuleDlp          `json:"dlp_actions,omitempty"`
	ExternalDlp               *bool                       `json:"external_dlp,omitempty"`
	JSONVersion               *int64                      `json:"json_version,omitempty"`
	MatchCriteriaAction       *MatchCriteriaAction        `json:"match_criteria_action,omitempty"`
	NetLocationObj            []string                    `json:"net_location_obj,omitempty"`
	Notify                    *Notify                     `json:"notify,omitempty"`
	OrganizationUnits         []string                    `json:"organization_units,omitempty"`
	Os                        []string                    `json:"os,omitempty"`
	PeriodicReauth            *PeriodicReauth             `json:"periodic_reauth,omitempty"`
	PrivateAppTagIds          []string                    `json:"privateAppTagIds,omitempty"`
	PrivateAppTags            []string                    `json:"privateAppTags,omitempty"`
	PrivateApps               []string                    `json:"privateApps,omitempty"`
	PrivateAppsWithActivities []PrivateAppsWithActivities `json:"privateAppsWithActivities,omitempty"`
	Schedule                  []ScheduleItem              `json:"schedule,omitempty"`
	ShowDlpProfileActionTable *bool                       `json:"show_dlp_profile_action_table,omitempty"`
	SrcCountries              []string                    `json:"srcCountries,omitempty"`
	TssActions                []NpaPolicyRuleTss          `json:"tss_actions,omitempty"`
	TssProfile                []string                    `json:"tss_profile,omitempty"`
	UserConfidence            *UserConfidence             `json:"user_confidence,omitempty"`
	UserGroups                []string                    `json:"userGroups,omitempty"`
	UserType                  *string                     `json:"userType,omitempty"`
	Users                     []string                    `json:"users,omitempty"`
	Version                   *int64                      `json:"version,omitempty"`
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

// ruleDataWithIntIDs overrides DeviceClassificationID as []int64 for API writes.
// The API returns strings on GET but expects integers on POST/PATCH (#64/#74).
type ruleDataWithIntIDs struct {
	RuleData
	DeviceClassificationID []int64 `json:"device_classification_id,omitempty"`
}

// requestWithIntIDs is the wire shape sent to the API. It mirrors myPolicyRequest
// but uses ruleDataWithIntIDs so device_classification_id is serialised as integers.
type requestWithIntIDs struct {
	Enabled    *string             `json:"enabled,omitempty"`
	ModifyBy   *string             `json:"modify_by,omitempty"`
	ModifyTime *string             `json:"modify_time,omitempty"`
	ModifyType *string             `json:"modify_type,omitempty"`
	PolicyType *string             `json:"policy_type,omitempty"`
	GroupID    *string             `json:"group_id,omitempty"`
	RuleData   *ruleDataWithIntIDs `json:"rule_data,omitempty"`
	RuleID     *string             `json:"rule_id,omitempty"`
	RuleName   *string             `json:"rule_name,omitempty"`
	RuleOrder  *RuleOrder          `json:"rule_order,omitempty"`
}

var _ beforeRequestHook = (*myPolicyRequest)(nil)

func (i *myPolicyRequest) BeforeRequest(hookCtx BeforeRequestContext, req *http.Request) (*http.Request, error) {
	if hookCtx.OperationID != "createNPARules" && hookCtx.OperationID != "updateNPARules" {
		return req, nil
	}

	return rewriteRequestJSON(req, func(body []byte) ([]byte, error) {
		var requestMap myPolicyRequest
		if err := json.Unmarshal(body, &requestMap); err != nil {
			return nil, fmt.Errorf("unable to unmarshal request: %v", err)
		}

		stripRuleOrderOnUpdate(hookCtx.OperationID, &requestMap)
		if requestMap.RuleData != nil {
			bracketPrivateApps(requestMap.RuleData)
			stripNegateFlagsWhenEmpty(requestMap.RuleData)
		}

		// Template field — no transformation for ANY action type.
		//
		// Verified against bespin and policydemo on 2026-10-21 (docs/api-audit/npa-rules-template):
		//   block           — the API resolves the display name to a filename on POST and
		//                     PATCH and returns the filename on GET. A filename on write is
		//                     rejected ("Undefined template"); an absent field is rejected
		//                     ("template field is required").
		//   periodic_reauth — the API stores the display name VERBATIM on POST and PATCH and
		//                     returns it on GET. A filename on write is rejected
		//                     ("Undefined template: 3.html") on both tenants.
		// So the display name from config is the only value the API accepts for either
		// action, and it must be sent unchanged. The 0.4.11–0.4.13-beta translation of
		// periodic_reauth names to filenames (via /api/v2/templates/usernotifications) was
		// based on a false premise and produced "Undefined template" on tenants where it
		// ran; it has been removed, and the provider no longer calls that endpoint at all.
		// The display-name/filename asymmetry for block rules is reconciled in the plan
		// modifier via private state (suppressTemplateDrift).
		// See https://github.com/netskopeoss/terraform-provider-netskope/issues/118
		// See https://github.com/netskopeoss/terraform-provider-netskope/issues/124

		idInts, err := coerceDeviceClassificationIDs(requestMap.RuleData)
		if err != nil {
			return nil, err
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
			rd.DeviceClassificationID = nil // clear string version; embedded field takes over
			marshalReq.RuleData = &ruleDataWithIntIDs{
				RuleData:               rd,
				DeviceClassificationID: idInts,
			}
		}

		return json.Marshal(marshalReq)
	})
}

// stripRuleOrderOnUpdate removes rule_order from full-rule update payloads.
// rule_order is create-time placement only. Sending it on a netskope_npa_rules
// update silently moves the rule to the configured position, which can demote
// block rules below allow rules (#129).
//
// Order-only PATCHes sent by netskope_npa_rules_order carry only rule_order
// (no rule_data), so their rule_order is preserved.
// See docs/bugs/BUG-021-npa-template-private-state.md.
func stripRuleOrderOnUpdate(op string, m *myPolicyRequest) {
	if op == "updateNPARules" && m.RuleData != nil {
		m.RuleOrder = nil
	}
}

// bracketPrivateApps wraps each private app name in brackets for the API write path.
// The API expects private app names wrapped in brackets on write and returns them
// wrapped on read; the AfterSuccess hook strips them so config and state hold plain names.
func bracketPrivateApps(rd *RuleData) {
	old := rd.PrivateApps
	rd.PrivateApps = nil
	for _, app := range old {
		rd.PrivateApps = append(rd.PrivateApps, "["+app+"]")
	}
}

// stripNegateFlagsWhenEmpty removes b_negateNetLocation / b_negateSrcCountries
// when their associated criteria lists are empty. The API rejects these flags
// even when false on tenants where the source-IP-criteria feature flag is disabled.
// The flags are meaningless without a populated criteria list.
func stripNegateFlagsWhenEmpty(rd *RuleData) {
	if len(rd.NetLocationObj) == 0 {
		rd.BNegateNetLocation = nil
	}
	if len(rd.SrcCountries) == 0 {
		rd.BNegateSrcCountries = nil
	}
}

// coerceDeviceClassificationIDs converts device_classification_id from []string
// to []int64 for the API write path. The API returns strings on GET but expects
// integers on POST/PATCH. Returns nil slice (not an error) when rd is nil.
func coerceDeviceClassificationIDs(rd *RuleData) ([]int64, error) {
	if rd == nil {
		return nil, nil
	}
	var out []int64
	for _, s := range rd.DeviceClassificationID {
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("unable to convert device_classification_id value %q to int: %w", s, err)
		}
		out = append(out, n)
	}
	return out, nil
}
