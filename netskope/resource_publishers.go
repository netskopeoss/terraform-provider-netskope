package netskope

import (
	"context"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/netskopeoss/netskope-api-client-go/nsgo"

func resourcePublisherCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	name := d.Get("name").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	//Init Options
	publisher := nsgo.PublisherOptions{
		Name: name,
	}

	newpublisher, err := nsclient.CreatePublisher(publisher)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.Itoa(newpublisher.ID))

	pubid := nsgo.PublisherOptions{
		Id: d.Get("id").(string),
	}

	token, err := nsclient.GetToken(pubid)
	if err != nil {
		return diag.FromErr(err)
	}

	d.Set("token", token.Token)

	return diags

}

func resourcePublisherRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics
	return diags
}

func resourcePublisherUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	name := d.Get("name").(string)
	id := d.Get("id").(string)

	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	//Init Options
	publisher := nsgo.PublisherOptions{
		Name: name,
		Id:   id,
	}

	_, err := nsclient.UpdatePublisher(publisher)
	if err != nil {
		return diag.FromErr(err)
	}
	return diags

}

func resourcePublisherDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	//Collect Diags
	var diags diag.Diagnostics

	//Get Vars from Schema
	id := d.Get("id").(string)
	//Init a client instance
	nsclient := m.(*nsgo.Client)
	//nsclient := nsgo.NewClient(BaseURL, ApiToken)

	//Init Options
	pubid := nsgo.PublisherOptions{
		Id: id,
	}

	//Delete Publisher
	_, err := nsclient.DeletePublisher(pubid)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return diags

}

func resourcePublishers() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePublisherCreate,
		ReadContext:   resourcePublisherRead,
		UpdateContext: resourcePublisherUpdate,
		DeleteContext: resourcePublisherDelete,
		Schema: map[string]*schema.Schema{
			"id": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"token": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}
