package netskope

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netskopeoss/netskope-api-client-go/nsgo"
)

type ExtendedPrivateAppList struct {
	PrivateApps *[]ExtendedPrivateAppListItem `json:"private_apps"`
}

type ExtendedPrivateAppListItem struct {
	AppID                       int                                              `json:"app_id"`
	AppName                     string                                           `json:"app_name"`
	ClientlessAccess            bool                                             `json:"clientless_access"`
	Host                        string                                           `json:"host"`
	PrivateAppProtocol          string                                           `json:"private_app_protocol"`
	Protocols                   *[]ExtendedPrivateAppProtocols                   `json:"protocols"`
	Reachability                ExtendedPrivateAppReachability                   `json:"reachability"`
	ServicePublisherAssignments *[]ExtendedPrivateAppServicePublisherAssignments `json:"service_publisher_assignments"`
	TrustSelfSignedCerts        bool                                             `json:"trust_self_signed_certs"`
	UsePublisherDNS             bool                                             `json:"use_publisher_dns"`
	Tags                        *[]ExtendedPrivateAppTags                        `json:"tags"`
}

type ExtendedPrivateAppTags struct {
	TagID   int    `json:"tag_id"`
	TagName string `json:"tag_name"`
}

type ExtendedPrivateAppProtocols struct {
	CreatedAt time.Time `json:"created_at"`
	ID        int       `json:"id"`
	Port      string    `json:"port"`
	ServiceID int       `json:"service_id"`
	Transport string    `json:"transport"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ExtendedPrivateAppReachability struct {
	//*Error
	Reachable bool `json:"reachable"`
}

type ExtendedPrivateAppServicePublisherAssignments struct {
	Primary      string                                                    `json:"primary"`
	PublisherID  int                                                       `json:"publisher_id"`
	Reachability ExtendedPrivateAppServicePublisherAssignmentsReachability `json:"reachability"`
	ServiceID    int                                                       `json:"service_id"`
}

type ExtendedPrivateAppServicePublisherAssignmentsReachability struct {
	//*Error
	Reachable bool `json:"reachable"`
}

type Error struct {
	ErrorCode   int    `json:"error_code"`
	ErrorString string `json:"error_string"`
}

func dataSourcePrivateAppsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Set Filter
	filter := d.Get("filter").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get Private Apps, marshal to ExtendedPrivateAppList type
	var privateAppList *ExtendedPrivateAppList
	appsInterface, err := nsclient.GetPrivateAppsWithFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}
	jsonData, err := json.Marshal(appsInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	json.Unmarshal(jsonData, &privateAppList)

	// Convert ExtendedPrivateAppList to map[string]interface{}
	privateAppsSchemaJson, _ := json.Marshal(privateAppList.PrivateApps)
	privateAppsSchemaMap := make([]map[string]interface{}, 0)
	json.Unmarshal(privateAppsSchemaJson, &privateAppsSchemaMap)

	if err := d.Set("private_apps", privateAppsSchemaMap); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func dataSourcePrivateApps() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePrivateAppsRead,
		Schema: map[string]*schema.Schema{
			"filter": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"private_apps": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"app_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"app_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"clientless_access": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"host": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"private_app_protocol": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						/*"reachability": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},*/
						"trust_self_signed_certs": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"use_publisher_dns": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"tags": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"tag_name": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"tag_id": &schema.Schema{
										Type:     schema.TypeFloat,
										Computed: true,
									},
								},
							},
						},
						"reachability": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeBool,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"reachable": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
						"protocols": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"created_at": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"updated_at": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
									"port": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"service_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
									"transport": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
								},
							},
						},
						"service_publisher_assignments": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"primary": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"publisher_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
									/*"reachability": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},*/
									"reachability": &schema.Schema{
										Type:     schema.TypeMap,
										Computed: true,
										Elem: &schema.Schema{
											Type: schema.TypeBool,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"reachable": &schema.Schema{
														Type:     schema.TypeString,
														Computed: true,
													},
												},
											},
										},
									},
									"service_id": &schema.Schema{
										Type:     schema.TypeInt,
										Computed: true,
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
