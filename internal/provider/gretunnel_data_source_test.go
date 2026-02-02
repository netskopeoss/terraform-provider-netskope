// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGRETunnelDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_gre_tunnel.test"
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "tunnel_id", resourceName, "tunnel_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "site", resourceName, "site"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_ip", resourceName, "source_ip"),
				),
			},
		},
	})
}

// Configuration functions

func testAccGRETunnelDataSourceConfig_basic(name string) string {
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
}

data "netskope_gre_tunnel" "test" {
  tunnel_id = netskope_gre_tunnel.test.tunnel_id
}
`, testAccProviderConfig(), name, randomIP)
}
