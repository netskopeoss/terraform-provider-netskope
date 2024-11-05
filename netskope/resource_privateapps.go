package netskope

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netskopeoss/netskope-api-client-go/nsgo"
)

type PublisherIdentity struct {
	PublisherID   float64 `json:"publisher_id"`
	PublisherName string  `json:"publisher_name"`
}

type Protocol struct {
	Type string `json:"transport"`
	Port string `json:"port"`
}

type ExtendedPrivateApp struct {
	*nsgo.PrivateApp
	Protocols  []Protocol          `json:"protocols"`
	Publishers []PublisherIdentity `json:"service_publisher_assignments"`
}

func resourcePrivateAppsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	app_name := d.Get("app_name").(string)
	host := d.Get("host").(string)
	protocols := d.Get("protocols").([]interface{})
	publisher_id := d.Get("publisher").([]interface{})
	tags := d.Get("tags").([]interface{})
	use_publisher_dns := d.Get("use_publisher_dns").(bool)
	clientless_access := d.Get("clientless_access").(bool)
	trust_self_signed_certs := d.Get("trust_self_signed_certs").(bool)
	//log.Println(app_name)
	//log.Println(host)

	//Init Client
	nsclient := m.(*nsgo.Client)

	//Marshall Protocols
	protocols_json, _ := json.Marshal(protocols)
	protocolStruct := []nsgo.Protocol{}
	json.Unmarshal(protocols_json, &protocolStruct)
	//log.Println(string(jsondata))
	//log.Println(protocolStruct)

	//Marshall Publishers
	publishers_json, _ := json.Marshal(publisher_id)
	publishersStruct := []nsgo.PublisherIdentity{}
	json.Unmarshal(publishers_json, &publishersStruct)

	//Marshall Tags
	tags_json, _ := json.Marshal(tags)
	tagsStruct := []nsgo.PrivateAppTags{}
	json.Unmarshal(tags_json, &tagsStruct)

	appStruct := nsgo.PrivateApp{
		AppName:              app_name,
		Host:                 host,
		Protocols:            protocolStruct,
		Publishers:           publishersStruct,
		Tags:                 tagsStruct,
		UsePublisherDNS:      use_publisher_dns,
		ClientlessAccess:     clientless_access,
		TrustSelfSignedCerts: trust_self_signed_certs,
	}

	newapp, err := nsclient.CreatePrivateApp(appStruct)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(newapp.Id))

	return diags

}

func resourcePrivateAppsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Collect Diags
	var diags diag.Diagnostics

	// Init Client
	nsclient, ok := m.(*nsgo.Client)
	if !ok {
		return diag.Errorf("Failed to convert interface to nsgo.Client")
	}

	// Get the resource ID from State
	appId := d.Id()

	// Query the API for the privateapp using the ID - marshal the JSON body to the ExtendedPrivateApp struct
	var privateApp *ExtendedPrivateApp
	appsIdInterface, err := nsclient.GetPrivateAppId(nsgo.PrivateAppOptions{Id: appId})
	if err != nil {
		if err.Error() != fmt.Sprintf("No private app with id '%s' is found.", appId) {
			return diag.FromErr(err)
		}
	}
	jsonIdData, err := json.Marshal(appsIdInterface)
	if err != nil {
		return diag.FromErr(err)
	}
	err = json.Unmarshal(jsonIdData, &privateApp)
	if err != nil {
		return diag.FromErr(err)
	}

	// Remove the private app if it does not exist, otherwise set the values
	if privateApp == nil {
		// Remove the private app from state
		d.SetId("")
		return diags
	} else {
		// Update the values in state
		d.SetId(strconv.Itoa(privateApp.Id))
		d.Set("app_name", strings.Trim(privateApp.AppName, "[]"))
		d.Set("host", privateApp.Host)
		d.Set("use_publisher_dns", privateApp.UsePublisherDNS)
		d.Set("clientless_access", privateApp.ClientlessAccess)
		d.Set("trust_self_signed_certs", privateApp.TrustSelfSignedCerts)
		d.Set("tags", flattenPrivateAppTags(privateApp.Tags))
		d.Set("protocols", flattenPrivateAppProtocols(privateApp.Protocols))
		d.Set("publisher", flattenPrivateAppPublishers(privateApp.Publishers))

		return diags
	}
}

func resourcePrivateAppsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	id := d.Get("id").(string)
	app_name := d.Get("app_name").(string)
	host := d.Get("host").(string)
	protocols := d.Get("protocols").([]interface{})
	publisher_id := d.Get("publisher").([]interface{})
	use_publisher_dns := d.Get("use_publisher_dns").(bool)
	tags := d.Get("tags").([]interface{})
	clientless_access := d.Get("clientless_access").(bool)
	trust_self_signed_certs := d.Get("trust_self_signed_certs").(bool)
	//log.Println(app_name)
	//log.Println(host)

	//Convert ID to Int
	id_int, _ := strconv.Atoi(id)

	//Init Options
	appid := nsgo.PrivateAppOptions{
		Id: id,
	}

	//Init Client
	nsclient := m.(*nsgo.Client)

	//Marshall Protocols
	protocols_json, _ := json.Marshal(protocols)
	protocolStruct := []nsgo.Protocol{}
	json.Unmarshal(protocols_json, &protocolStruct)
	//log.Println(string(jsondata))
	//log.Println(protocolStruct)

	//Marshall Publishers
	publishers_json, _ := json.Marshal(publisher_id)
	publishersStruct := []nsgo.PublisherIdentity{}
	json.Unmarshal(publishers_json, &publishersStruct)

	//Marshall Tags
	tags_json, _ := json.Marshal(tags)
	tagsStruct := []nsgo.PrivateAppTags{}
	json.Unmarshal(tags_json, &tagsStruct)

	appStruct := nsgo.PrivateApp{
		Id:                   id_int,
		AppName:              app_name,
		Host:                 host,
		Protocols:            protocolStruct,
		Publishers:           publishersStruct,
		Tags:                 tagsStruct,
		UsePublisherDNS:      use_publisher_dns,
		ClientlessAccess:     clientless_access,
		TrustSelfSignedCerts: trust_self_signed_certs,
	}

	_, err := nsclient.ReplacePrivateApp(appid, appStruct)
	if err != nil {
		return diag.FromErr(err)
	}

	//d.SetId(strconv.Itoa(newapp.Id))
	return diags

}

func resourcePrivateAppsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	id := d.Get("id").(string)
	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	//Init Options
	appid := nsgo.PrivateAppOptions{
		Id: id,
	}

	//Delete Publisher
	_, err := nsclient.DeletePrivateApp(appid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags
}

func resourcePrivateApps() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePrivateAppsCreate,
		ReadContext:   resourcePrivateAppsRead,
		UpdateContext: resourcePrivateAppsUpdate,
		DeleteContext: resourcePrivateAppsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"app_name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"host": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"use_publisher_dns": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"clientless_access": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"trust_self_signed_certs": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"tags": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tag_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"protocols": &schema.Schema{
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"port": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"publisher": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"publisher_id": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"publisher_name": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func flattenPrivateAppProtocols(protocols []Protocol) []interface{} {
	flattened := make([]interface{}, len(protocols))
	for i, protocol := range protocols {
		flattened[i] = map[string]interface{}{
			"type": protocol.Type,
			"port": protocol.Port,
		}
	}
	return flattened
}

func flattenPrivateAppPublishers(publishers []PublisherIdentity) []interface{} {
	flattened := make([]interface{}, len(publishers))
	for i, publisher := range publishers {
		flattened[i] = map[string]interface{}{
			"publisher_id":   strconv.Itoa(int(publisher.PublisherID)),
			"publisher_name": publisher.PublisherName,
		}
	}
	return flattened
}

func flattenPrivateAppTags(tags []nsgo.PrivateAppTags) []interface{} {
	flattened := make([]interface{}, len(tags))
	for i, tag := range tags {
		flattened[i] = map[string]interface{}{
			"tag_name": tag.TagName,
		}
	}
	return flattened
}
