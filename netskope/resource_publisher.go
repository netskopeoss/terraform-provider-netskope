package netskope

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/ns-sbrown/nsgo"
)

func resourcePublisherCreate(d *schema.ResourceData, m interface{}) error {

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
		return err
	}

	d.SetId(strconv.Itoa(newpublisher.ID))

	/*
		pubid := nsgo.PublisherOptions{
			Id: d.Get("id").(string),
		}

			token, err := nsclient.GetToken(pubid)
			if err != nil {
				return err
			}

			d.Set("token", token.Token)
	*/
	return nil

}

func resourcePublisherRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourcePublisherUpdate(d *schema.ResourceData, m interface{}) error {
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
		return err
	}
	return nil

}

func resourcePublisherDelete(d *schema.ResourceData, m interface{}) error {
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
		return err
	}
	d.SetId("")
	return nil

}

func resourcePublisher() *schema.Resource {
	return &schema.Resource{
		Create:   resourcePublisherCreate,
		Read:     resourcePublisherRead,
		Update:   resourcePublisherUpdate,
		Delete:   resourcePublisherDelete,
		Importer: &schema.ResourceImporter{State: schema.ImportStatePassthrough},
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
