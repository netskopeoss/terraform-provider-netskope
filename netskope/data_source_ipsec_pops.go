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

func dataSourceIpsecPopsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get IPSec PoPs
	pops, err := nsclient.GetIpsecPops()
	if err != nil {
		return diag.FromErr(err)
	}

	jsonData, _ := json.Marshal(pops)

	popsStruct := nsgo.IpsecPops{}
	json.Unmarshal(jsonData, &popsStruct)

	newjsonData, _ := json.Marshal(popsStruct)
	popsMap := make([]map[string]interface{}, 0)
	json.Unmarshal(newjsonData, &popsMap)

	if err := d.Set("ipsec_pops", popsMap); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags
}

func dataSourceIpsecPops() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceIpsecPopsRead,
		Schema: map[string]*schema.Schema{
			"ipsec_pops": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"gateway": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"closestpop": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"location": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"probeip": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"region": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						/*"options": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeMap,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"phase1": &schema.Schema{
											Type:     schema.TypeList,
											Computed: true,
											Elem: &schema.Resource{
												Schema: map[string]*schema.Schema{
													"dhgroup": &schema.Schema{
														Type:     schema.TypeString,
														Computed: true,
													},
													"encryptionalgo": &schema.Schema{
														Type:     schema.TypeString,
														Computed: true,
													},
													"ikeversion": &schema.Schema{
														Type:     schema.TypeString,
														Computed: true,
													},
													"integrityalgo": &schema.Schema{
														Type:     schema.TypeString,
														Computed: true,
													},
													"salifetime": &schema.Schema{
														Type:     schema.TypeInt,
														Computed: true,
													},
												},
											},
										},
									},
								},
							},
						},*/
					},
				},
			},
		},
	}
}
