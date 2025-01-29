package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type myPolicyResponse struct {
	Data   Data   `json:"data"`
	Status string `json:"status"`
}

type Data struct {
	AccessMethod              []string                    `json:"access_method"`
	BNegateNetLocation        bool                        `json:"b_negate_net_location"`
	BNegateSrcCountries       bool                        `json:"b_negate_src_countries"`
	Classification            string                      `json:"classification"`
	DeviceClassificationID    []int                       `json:"device_classification_id"`
	DlpActions                []NpaPolicyRuleDlp          `json:"dlp_actions"`
	ExternalDlp               bool                        `json:"external_dlp"`
	JSONVersion               int                         `json:"json_version"`
	MatchCriteriaAction       *MatchCriteriaAction        `json:"match_criteria_action"`
	NetLocationObj            []string                    `json:"net_location_obj"`
	OrganizationUnits         []string                    `json:"organization_units"`
	PolicyType                string                      `json:"policy_type"`
	PrivateApps               []string                    `json:"private_apps"`
	PrivateAppsWithActivities []PrivateAppsWithActivities `json:"private_apps_with_activities"`
	PrivateAppTagIds          []string                    `json:"private_app_tag_ids"`
	PrivateAppTags            []string                    `json:"private_app_tags"`
	ShowDlpProfileActionTable bool                        `json:"show_dlp_profile_action_table"`
	SrcCountries              []string                    `json:"src_countries"`
	TssActions                []NpaPolicyRuleTss          `json:"tss_actions"`
	TssProfile                []string                    `json:"tss_profile"`
	UserGroups                []string                    `json:"user_groups"`
	Users                     []string                    `json:"users"`
	UserType                  string                      `json:"user_type"`
	Version                   int                         `json:"version"`
}

type NpaPolicyRuleDlp struct {
	Actions    []string `json:"actions"`
	DlpProfile string   `json:"dlp_profile"`
}

type MatchCriteriaAction struct {
	ActionName string `json:"action_name"`
}

type PrivateAppsWithActivities struct {
	Activities []Activities `json:"activities"`
	AppID      []string     `json:"app_id"`
	AppName    string       `json:"app_name"`
}

type NpaPolicyRuleTss struct {
	Actions    []NpaPolicyRuleTssActions `json:"actions"`
	TssProfile []string                  `json:"tss_profile"`
}

type Activities struct {
	Activity          string   `json:"activity"`
	ListOfConstraints []string `json:"list_of_constraints"`
}

type NpaPolicyRuleTssActions struct {
	ActionName         string `json:"action_name"`
	RemediationProfile string `json:"remediation_profile"`
	Severity           string `json:"severity"`
	Template           string `json:"template"`
}

var (
	_ afterSuccessHook = (*myPolicyResponse)(nil)
)

func (i *myPolicyResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	log.Print("Executing AfterSucess hook....")
	if hookCtx.OperationID == "createNPARules" || hookCtx.OperationID == "getNPARules" || hookCtx.OperationID == "NPARules" {
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
		oldValue := responseMap.Data.PrivateApps
		for _, untrimmedApp := range oldValue {
			trimmed := strings.Trim(untrimmedApp, "[]")
			responseMap.Data.PrivateApps = append(responseMap.Data.PrivateApps, trimmed)
		}
		// responseMap.Data.AppName = strings.Trim(oldValue, "[]")
		// Marshal the modified response back to json.RawMessage
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
