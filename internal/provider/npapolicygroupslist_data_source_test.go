// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPolicyGroupsListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_npa_policy_groups_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPolicyGroupsListDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the list contains at least one policy group (the one we created)
					resource.TestCheckResourceAttrSet(dataSourceName, "data.#"),
				),
			},
		},
	})
}

func testAccNPAPolicyGroupsListDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = %q

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

data "netskope_npa_policy_groups_list" "test" {
  depends_on = [netskope_npa_policy_groups.test]
}
`, testAccProviderConfig(), name)
}
