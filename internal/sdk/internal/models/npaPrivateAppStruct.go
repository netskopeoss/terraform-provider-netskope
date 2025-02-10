package models

type BulkAppData struct {
	AppData []AppData `json:"private_apps"`
}

type AppData struct {
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
