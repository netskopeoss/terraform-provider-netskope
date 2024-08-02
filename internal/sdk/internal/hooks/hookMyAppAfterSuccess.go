package hooks

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type MyAppResponse struct {
	Data   Data   `json:"data"`
	Status string `json:"status"`
}
type Data struct {
	AllowUnauthenticatedCors    bool                         `json:"allow_unauthenticated_cors"`
	AppID                       int                          `json:"app_id"`
	AppName                     string                       `json:"app_name"`
	AppOption                   map[string]interface{}       `json:"app_option"`
	ClientlessAccess            bool                         `json:"clientless_access"`
	Host                        string                       `json:"host"`
	ID                          int                          `json:"id"`
	IsUserPortalApp             bool                         `json:"is_user_portal_app"`
	ModifiedBy                  string                       `json:"modified_by"`
	ModifyTime                  string                       `json:"modify_time"`
	Name                        string                       `json:"name"`
	Policies                    []string                     `json:"policies"`
	PrivateAppProtocol          string                       `json:"private_app_protocol"`
	Protocols                   []Protocol                   `json:"protocols"`
	PublicHost                  string                       `json:"public_host"`
	Reachability                Reachability                 `json:"reachability"`
	RealHost                    *string                      `json:"real_host"`
	ServicePublisherAssignments []ServicePublisherAssignment `json:"service_publisher_assignments"`
	SteeringConfigs             []string                     `json:"steering_configs"`
	SupplementDNSForOsx         bool                         `json:"supplement_dns_for_osx"`
	Tags                        []Tag                        `json:"tags"`
	TrustSelfSignedCerts        *bool                        `json:"trust_self_signed_certs"`
	UsePublisherDNS             *bool                        `json:"use_publisher_dns"`
}
type Reachability struct {
	Reachable bool `json:"reachable"`
}
type Protocol struct {
	CreatedAt string `json:"created_at"`
	ID        int    `json:"id"`
	Port      string `json:"port"`
	ServiceID int    `json:"service_id"`
	Transport string `json:"transport"`
	UpdatedAt string `json:"updated_at"`
}
type ServicePublisherAssignment struct {
	Primary       *bool         `json:"primary"`
	PublisherID   int           `json:"publisher_id"`
	PublisherName string        `json:"publisher_name"`
	Reachability  *Reachability `json:"reachability"`
	ServiceID     int           `json:"service_id"`
}
type Tag struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

var (
	_ afterSuccessHook = (*MyAppResponse)(nil)
)

func (i *MyAppResponse) AfterSuccess(hookCtx AfterSuccessContext, res *http.Response) (*http.Response, error) {
	log.Print("Executing AfterSucess hook....")
	if hookCtx.OperationID == "createNPAPrivateApps" || hookCtx.OperationID == "getNPAPrivateApp" {
		var responseMap MyAppResponse
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
		oldValue := responseMap.Data.AppName
		responseMap.Data.AppName = strings.Trim(oldValue, "[]")
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
