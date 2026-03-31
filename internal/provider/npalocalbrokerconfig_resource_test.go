package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPALocalBrokerConfig_basic(t *testing.T) {
	rHostname := fmt.Sprintf("lbr-%s.example.com", acctest.RandString(8))
	resourceName := "netskope_npa_local_broker_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"hostname": config.StringVariable(rHostname),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "data.hostname", rHostname),
				),
			},
		},
	})
}

func TestAccNPALocalBrokerConfig_update(t *testing.T) {
	rHostname := fmt.Sprintf("lbr-%s.example.com", acctest.RandString(8))
	rHostnameUpdated := fmt.Sprintf("lbr-%s-updated.example.com", acctest.RandString(8))
	resourceName := "netskope_npa_local_broker_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"hostname": config.StringVariable(rHostname),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"hostname": config.StringVariable(rHostnameUpdated),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostnameUpdated),
				),
			},
		},
	})
}
