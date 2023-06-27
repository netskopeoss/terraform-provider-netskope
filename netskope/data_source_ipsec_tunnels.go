package netskope

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jeff-clearcover/netskope-api-client-go/nsgo"
)

func dataSourceIpsecTunnelsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get IPSec PoPs
	tunnels, err := nsclient.GetIpsecTunnels()
	if err != nil {
		return diag.FromErr(err)
	}

	jsonData, _ := json.Marshal(tunnels)

	tunnelsStruct := nsgo.IpsecTunnels{}
	json.Unmarshal(jsonData, &tunnelsStruct)

	newjsonData, _ := json.Marshal(tunnelsStruct)
	tunnelsMap := make([]map[string]interface{}, 0)
	json.Unmarshal(newjsonData, &tunnelsMap)

	if err := d.Set("ipsec_tunnels", tunnelsMap); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func dataSourceIpsecTunnels() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIpsecTunnelsRead,
		Schema: map[string]*schema.Schema{
			"ipsec_tunnels": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"site": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"template": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"sourcetype": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"notes": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"encryption": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"srcidentity": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"srcipidentity": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"status": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"status": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"since": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"throughput": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
									},
								},
							},
						},
						"pops": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"name": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"gateway": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"probeip": &schema.Schema{
										Type:     schema.TypeString,
										Computed: true,
									},
									"primary": &schema.Schema{
										Type:     schema.TypeBool,
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
