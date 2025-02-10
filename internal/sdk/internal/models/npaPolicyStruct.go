package models

type PolicyData struct {
	Enabled    *string   `json:"enabled,omitempty"`
	ModifyBy   *string   `json:"modify_by,omitempty"`
	ModifyTime *string   `json:"modify_time,omitempty"`
	ModifyType *string   `json:"modify_type,omitempty"`
	PolicyType *string   `json:"policy_type,omitempty"`
	GroupID    *string   `json:"group_id,omitempty"`
	RuleData   *RuleData `json:"rule_data,omitempty"`
	RuleID     *string   `json:"rule_id,omitempty"`
	RuleName   *string   `json:"rule_name,omitempty"`
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
