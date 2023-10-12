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

func dataSourcePublisherUpgradeProfilesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Get the list of publisher upgrade profiles
	upgradeprofiles, err := nsclient.GetPublisherUpgradeProfiles()
	if err != nil {
		return diag.FromErr(err)
	}

	// Marshall Data into the PublisherUpgradeProfiles struct
	jsonData, _ := json.Marshal(upgradeprofiles)
	profileStruct := nsgo.PublisherUpgradeProfiles{}
	json.Unmarshal(jsonData, &profileStruct)

	//Unmarshal the upgrade profiles into a map
	newjsonData, _ := json.Marshal(profileStruct.UpgradeProfiles)
	profilesMap := make([]map[string]interface{}, 0)
	json.Unmarshal(newjsonData, &profilesMap)

	// Set the upgrade profiles in the Terraform schema
	if err := d.Set("upgrade_profiles", profilesMap); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))
	return diags

}

func dataSourcePublisherUpgradeProfiles() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourcePublisherUpgradeProfilesRead,
		Schema: map[string]*schema.Schema{
			"upgrade_profiles": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"created_at": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"docker_tag": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"enabled": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
						"external_id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"frequency": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"next_update_time": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"num_associated_publisher": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"release_type": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"updated_at": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"timezone": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"upgrading_stage": &schema.Schema{
							Type:     schema.TypeInt,
							Computed: true,
						},
						"will_start": &schema.Schema{
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}
