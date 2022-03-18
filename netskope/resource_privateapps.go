package netskope

import (
	"encoding/json"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ns-sbrown/nsgo"
)

func resourcePrivateAppsCreate(d *schema.ResourceData, m interface{}) error {

	//Get Vars from Schema
	app_name := d.Get("app_name").(string)
	host := d.Get("host").(string)
	protocols := d.Get("protocols").([]interface{})
	publisher_id := d.Get("publisher").([]interface{})
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

	appStruct := nsgo.PrivateApp{
		AppName:              app_name,
		Host:                 host,
		Protocols:            protocolStruct,
		Publishers:           publishersStruct,
		UsePublisherDNS:      use_publisher_dns,
		ClientlessAccess:     clientless_access,
		TrustSelfSignedCerts: trust_self_signed_certs,
	}

	newapp, err := nsclient.CreatePrivateApp(appStruct)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(newapp.Id))

	return nil

}

func resourcePrivateAppsRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourcePrivateAppsUpdate(d *schema.ResourceData, m interface{}) error {
	//Get Vars from Schema
	id := d.Get("id").(string)
	app_name := d.Get("app_name").(string)
	host := d.Get("host").(string)
	protocols := d.Get("protocols").([]interface{})
	publisher_id := d.Get("publisher").([]interface{})
	use_publisher_dns := d.Get("use_publisher_dns").(bool)
	clientless_access := d.Get("clientless_access").(bool)
	trust_self_signed_certs := d.Get("trust_self_signed_certs").(bool)
	//log.Println(app_name)
	//log.Println(host)

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

	appStruct := nsgo.PrivateApp{
		AppName:              app_name,
		Host:                 host,
		Protocols:            protocolStruct,
		Publishers:           publishersStruct,
		UsePublisherDNS:      use_publisher_dns,
		ClientlessAccess:     clientless_access,
		TrustSelfSignedCerts: trust_self_signed_certs,
	}

	_, err := nsclient.UpdatePrivateApp(appid, appStruct)
	if err != nil {
		return err
	}

	//d.SetId(strconv.Itoa(newapp.Id))
	return nil

}

func resourcePrivateAppsDelete(d *schema.ResourceData, m interface{}) error {
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
		return err
	}
	d.SetId("")
	return nil
}

func resourcePrivateApps() *schema.Resource {
	return &schema.Resource{
		Create: resourcePrivateAppsCreate,
		Read:   resourcePrivateAppsRead,
		Update: resourcePrivateAppsUpdate,
		Delete: resourcePrivateAppsDelete,
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
