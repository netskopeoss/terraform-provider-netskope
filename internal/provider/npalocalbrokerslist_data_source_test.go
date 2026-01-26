// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPALocalBrokersListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_npa_local_brokers_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokersListDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the list contains local brokers
					resource.TestCheckResourceAttrSet(dataSourceName, "data.#"),
				),
			},
		},
	})
}

func TestAccNPALocalBrokersListDataSource_empty(t *testing.T) {
	dataSourceName := "data.netskope_npa_local_brokers_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokersListDataSourceConfig_empty(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the data attribute is set (may be empty if no brokers exist)
					resource.TestCheckResourceAttrSet(dataSourceName, "data.#"),
				),
			},
		},
	})
}

func testAccNPALocalBrokersListDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name = %q
}

data "netskope_npa_local_brokers_list" "test" {
  depends_on = [netskope_npa_local_broker.test]
}
`, testAccProviderConfig(), name)
}

func testAccNPALocalBrokersListDataSourceConfig_empty() string {
	return fmt.Sprintf(`
%s

data "netskope_npa_local_brokers_list" "test" {
}
`, testAccProviderConfig())
}
