package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPALocalBrokerDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	dataSourceName := "data.netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_id",
						resourceName, "local_broker_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_name",
						resourceName, "local_broker_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "access_via_public_ip",
						resourceName, "access_via_public_ip",
					),
				),
			},
		},
	})
}

func TestAccNPALocalBrokerDataSource_withLocation(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	dataSourceName := "data.netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_id",
						resourceName, "local_broker_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "city_name",
						resourceName, "city_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "region_name",
						resourceName, "region_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "country_name",
						resourceName, "country_name",
					),
				),
			},
		},
	})
}
