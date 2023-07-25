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

func dataSourcePrivateAppsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Set Filter
	filter := d.Get("filter").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get Publishers
	apps, err := nsclient.GetPrivateAppsWithFilter(filter)
	if err != nil {
		return diag.FromErr(err)
	}

	jsonData, _ := json.Marshal(apps)

	appsStruct := nsgo.PrivateAppsList{}
	json.Unmarshal(jsonData, &appsStruct)

	newjsonData, _ := json.Marshal(appsStruct.PrivateApps)
	appsMap := make([]map[string]interface{}, 0)
	json.Unmarshal(newjsonData, &appsMap)

	if err := d.Set("private_apps", appsMap); err != nil {
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
