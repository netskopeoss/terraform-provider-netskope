package netskope

import (
	"context"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jeff-clearcover/netskope-api-client-go/nsgo"
)

func resourceIpsecTunnelsCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	encryption := d.Get("encryption").(string)
	site := d.Get("site").(string)
	srcidentity := d.Get("srcidentity").(string)
	srcipidentity := d.Get("srcipidentity").(string)
	psk := d.Get("psk").(string)
	notes := d.Get("notes").(string)
	sourcetype := d.Get("sourcetype").(string)
	pops := d.Get("pops").(([]interface{}))
	bandwidth := d.Get("bandwidth").(int)
	enable := d.Get("enable").(bool)

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Define New Tunnel Struct
	newtunnel := nsgo.NewIpsecTunnel{
		Encryption:    encryption,
		Site:          site,
		Srcidentity:   srcidentity,
		Srcipidentity: srcipidentity,
		Psk:           psk,
		Pops:          pops,
		Bandwidth:     bandwidth,
		Notes:         notes,
		Sourcetype:    sourcetype,
		Enable:        enable,
	}

	_, err := nsclient.CreateIpsecTunnel(newtunnel)
	if err != nil {
		return diag.FromErr(err)
	}

	//Fix this once API is updated
	//d.SetId(strconv.Itoa(tunnel.ID))
	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}

func resourceIpsecTunnelsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	return diags
}

func resourceIpsecTunnelsUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	encryption := d.Get("encryption").(string)
	site := d.Get("site").(string)
	srcidentity := d.Get("srcidentity").(string)
	srcipidentity := d.Get("srcipidentity").(string)
	psk := d.Get("psk").(string)
	notes := d.Get("notes").(string)
	sourcetype := d.Get("sourcetype").(string)
	pops := d.Get("pops").(([]interface{}))
	bandwidth := d.Get("bandwidth").(int)
	enable := d.Get("enable").(bool)
	id := d.Get("id").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Set Request Options
	//Options Struct
	tunnelid := nsgo.RequestOptions{
		Id: id,
	}

	//Define New Tunnel Struct
	updatetunnel := nsgo.NewIpsecTunnel{
		Encryption:    encryption,
		Site:          site,
		Srcidentity:   srcidentity,
		Srcipidentity: srcipidentity,
		Psk:           psk,
		Pops:          pops,
		Bandwidth:     bandwidth,
		Notes:         notes,
		Sourcetype:    sourcetype,
		Enable:        enable,
	}

	//Update Tunnel
	_, err := nsclient.UpdateIpsecTunnel(tunnelid, updatetunnel)
	if err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIpsecTunnelsDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Request Vars
	id := d.Get("id").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)

	//Options Struct
	tunnelid := nsgo.RequestOptions{
		Id: id,
	}

	//Delete tunnel
	_, err := nsclient.DeleteIpsecTunnel(tunnelid)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

func resourceIpsecTunnels() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIpsecTunnelsCreate,
		ReadContext:   resourceIpsecTunnelsRead,
		UpdateContext: resourceIpsecTunnelsUpdate,
		DeleteContext: resourceIpsecTunnelsDelete,
		Schema: map[string]*schema.Schema{
			//['site', 'srcidentity', 'psk', 'encryption', 'pops', 'bandwidth']
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"encryption": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"site": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"srcidentity": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"srcipidentity": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"psk": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"notes": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"sourcetype": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"bandwidth": &schema.Schema{
				Type:     schema.TypeInt,
				Required: true,
			},
			"enable": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},
			"pops": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}
