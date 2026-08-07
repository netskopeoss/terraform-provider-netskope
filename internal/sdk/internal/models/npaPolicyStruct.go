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
	ShowAisecProfileActionTable *bool                       `json:"show_aisec_profile_action_table,omitempty"`
	ShowDlpProfileActionTable   *bool                       `json:"show_dlp_profile_action_table,omitempty"`
	SrcCountries                []string                    `json:"srcCountries,omitempty"`
	TssActions                  []NpaPolicyRuleTss          `json:"tss_actions,omitempty"`
	TssProfile                  []string                    `json:"tss_profile,omitempty"`
	UserConfidence              *UserConfidence             `json:"user_confidence,omitempty"`
	UserGroupObjects            []UserGroupObject           `json:"userGroupObjects,omitempty"`
	UserGroups                  []string                    `json:"userGroups,omitempty"`
	UserType                    *string                     `json:"userType,omitempty"`
	Users                       []string                    `json:"users,omitempty"`
	Version                     *int64                      `json:"version,omitempty"`
}

type Notify struct {
	Emails    []string     `json:"emails,omitempty"`
	FromUser  *string      `json:"from_user,omitempty"`
	Interval  *string      `json:"interval,omitempty"`
	Templates [][]string   `json:"templates,omitempty"`
	ToUsers   []string     `json:"to_users,omitempty"`
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

type UserGroupObject struct {
	Disabled *string `json:"disabled,omitempty"`
	ID       *string `json:"id,omitempty"`
	Name     *string `json:"name,omitempty"`
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
