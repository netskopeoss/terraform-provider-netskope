// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPolicyGroupsDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_policy_groups.test"
	dataSourceName := "data.netskope_npa_policy_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPolicyGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "id",
						resourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "group_name",
						resourceName, "group_name",
					),
				),
			},
		},
	})
}

func testAccNPAPolicyGroupsDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = %q

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

data "netskope_npa_policy_groups" "test" {
  id = netskope_npa_policy_groups.test.id
}
`, testAccProviderConfig(), name)
}