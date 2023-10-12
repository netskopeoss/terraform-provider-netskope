package netskope

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netskopeoss/netskope-api-client-go/nsgo"
)

func resourcePublisherUpgradeProfileCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	profile_name := d.Get("name").(string)
	timezone := d.Get("timezone").(string)
	release_type := d.Get("release_type").(string)
	docker_tag := d.Get("docker_tag").(string)
	frequency := d.Get("frequency").(string)
	profile_enabled := d.Get("enabled").(bool)

	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	upgradeprofileoptions := nsgo.PublisherUpgradeProfileOptions{
		Name:        profile_name,
		Timezone:    timezone,
		ReleaseType: release_type,
		DockerTag:   docker_tag,
		Frequency:   frequency,
		Enabled:     profile_enabled,
	}

	newupgradeprofile, err := nsclient.CreatePublisherUpgradeProfile(upgradeprofileoptions)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(newupgradeprofile.ID))
	d.Set("created_at", newupgradeprofile.CreatedAt)
	d.Set("next_update_time", newupgradeprofile.NextUpdateTime)
	d.Set("updated_at", newupgradeprofile.UpdatedAt)
	d.Set("will_start", newupgradeprofile.WillStart)
	d.Set("upgrading_stage", newupgradeprofile.UpgradingStage)

	return diags
}

func resourcePublisherUpgradeProfileRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics
	return diags
}

func resourcePublisherUpgradeProfileUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	profile_name := d.Get("name").(string)
	timezone := d.Get("timezone").(string)
	release_type := d.Get("release_type").(string)
	docker_tag := d.Get("docker_tag").(string)
	frequency := d.Get("frequency").(string)
	profile_enabled := d.Get("enabled").(bool)
	external_id := d.Get("id").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	upgradeprofileoptions := nsgo.PublisherUpgradeProfileOptions{
		Name:        profile_name,
		Timezone:    timezone,
		ReleaseType: release_type,
		DockerTag:   docker_tag,
		Frequency:   frequency,
		Enabled:     profile_enabled,
		ID:          external_id,
	}

	updatedprofile, err := nsclient.UpdatePublisherUpgradeProfile(upgradeprofileoptions)
	if err != nil {
		return diag.FromErr(err)
	}

	//d.SetId(strconv.Itoa(upgradeupgradeprofile.ID))
	//d.Set("created_at", newupgradeprofile.CreatedAt)
	d.Set("next_update_time", updatedprofile.NextUpdateTime)
	d.Set("updated_at", updatedprofile.UpdatedAt)
	d.Set("will_start", updatedprofile.WillStart)
	d.Set("upgrading_stage", updatedprofile.UpgradingStage)

	return diags
}

func resourcePublisherUpgradeProfileDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	id := d.Get("id").(string)
	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	//Init Options
	profileid := nsgo.PublisherUpgradeProfileOptions{
		ExternalID: id,
	}

	//Delete Publisher
	_, err := nsclient.DeletePublisherUpgradeProfile(profileid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourcePublisherUpgradeProfiles() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePublisherUpgradeProfileCreate,
		ReadContext:   resourcePublisherUpgradeProfileRead,
		UpdateContext: resourcePublisherUpgradeProfileUpdate,
		DeleteContext: resourcePublisherUpgradeProfileDelete,
		Schema: map[string]*schema.Schema{
			"created_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"docker_tag": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": &schema.Schema{
				Type:     schema.TypeBool,
				Required: true,
			},
			"frequency": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"next_update_time": &schema.Schema{
				Type:     schema.TypeInt,
				Computed: true,
			},
			"release_type": {
				Type:     schema.TypeString,
				Required: true,
			},
			"updated_at": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"timezone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
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
	}
}
