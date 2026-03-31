package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccGRETunnelDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	dataSourceName := "data.netskope_gre_tunnel.test"
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(rName),
					"source_ip": config.StringVariable(randomIP),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tunnel_id", resourceName, "tunnel_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site", resourceName, "site"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_ip", resourceName, "source_ip"),
				),
			},
		},
	})
}
