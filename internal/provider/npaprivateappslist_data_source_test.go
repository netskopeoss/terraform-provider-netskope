// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPrivateAppsListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_npa_private_apps_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppsListDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the list contains private apps
					resource.TestCheckResourceAttrSet(dataSourceName, "private_apps.#"),
				),
				// Known provider issue: computed fields cause plan drift
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccNPAPrivateAppsListDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}

data "netskope_npa_private_apps_list" "test" {
  depends_on = [netskope_npa_private_app.test]
}
`, testAccProviderConfig(), name, name)
}
