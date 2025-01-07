package netskope

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netskopeoss/netskope-api-client-go/nsgo"
)

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
	private_app_protocol := d.Get("private_app_protocol").(string)
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
		PrivateAppProtocol:   private_app_protocol,
	}

	newapp, err := nsclient.CreatePrivateApp(appStruct)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(newapp.Id))

	return diags

}

func resourcePrivateAppsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics
	return diags
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
	private_app_protocol := d.Get("private_app_protocol").(string)
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
		PrivateAppProtocol:   private_app_protocol
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
			"private_app_protocol": &schema.Schema{
               		 Type:        schema.TypeString,
              		  Optional:    true,
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
