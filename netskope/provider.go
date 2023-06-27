package netskope

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jeff-clearcover/netskope-api-client-go/nsgo"
)

// Provider - Netskope APIv2 Provider
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"baseurl": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NS_BaseURL", nil),
			},
			"apitoken": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("NS_ApiToken", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"netskope_publishers":    resourcePublishers(),
			"netskope_privateapps":   resourcePrivateApps(),
			"netskope_ipsec_tunnels": resourceIpsecTunnels(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"netskope_publishers":    dataSourcePublishers(),
			"netskope_privateapps":   dataSourcePrivateApps(),
			"netskope_ipsec_pops":    dataSourceIpsecPops(),
			"netskope_ipsec_tunnels": dataSourceIpsecTunnels(),
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	//Client Config
	config := nsgo.Config{
		BaseURL:  d.Get("baseurl").(string),
		ApiToken: d.Get("apitoken").(string),
	}
	nsclient := nsgo.NewRetryClient(config)
	return nsclient, nil
}
