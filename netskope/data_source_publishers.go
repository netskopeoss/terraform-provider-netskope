package netskope

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ns-sbrown/nsgo"
)

func dataSourcePublishersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get Publishers
	pubs, err := nsclient.GetPublishers()
	if err != nil {
		return diag.FromErr(err)
	}

	jsonData, _ := json.Marshal(pubs)

	pubsStruct := nsgo.PublishersList{}
	json.Unmarshal(jsonData, &pubsStruct)

	newjsonData, _ := json.Marshal(pubsStruct.Publishers)
	pubsMap := make([]map[string]interface{}, 0)
	json.Unmarshal(newjsonData, &pubsMap)

	if err := d.Set("publishers", pubsMap); err != nil {
		/*
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create HashiCups client",
				Detail:   string([]byte(newjsonData)),
			})
			return diags
		*/
		return diag.FromErr(err)
	}

	/*
		//Get Publisher
		pubs, err := nsclient.GetPublishers()
		if err != nil {
			return diag.FromErr(err)
		}

		pubsMap := make([]map[string]interface{}, 0)
		err = json.NewDecoder(pubs).Decode(&pubsMap)
		if err != nil {
			return diag.FromErr(err)
		}

		if err := d.Set("publishers", pubsMap); err != nil {
			return diag.FromErr(err)
		}

		// always run
		d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	*/
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags

}

func dataSourcePublishers() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePublishersRead,
		Schema: map[string]*schema.Schema{
			"publishers": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"publisher_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"publisher_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"publisher_upgrade_profiles_id": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"upgrade_failed_reason": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						}, "upgrade_request": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"common_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"registered": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"status": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"stitcher_id": &schema.Schema{
							Type:     schema.TypeFloat,
							Computed: true,
						},
						"assessment": &schema.Schema{
							Type:     schema.TypeMap,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
								Elem: &schema.Resource{
									Schema: map[string]*schema.Schema{
										"eee_support": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"hdd_free": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"hdd_total": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"ip_address": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
										"version": &schema.Schema{
											Type:     schema.TypeString,
											Computed: true,
										},
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
