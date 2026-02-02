// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGREPOPDataSource_basic(t *testing.T) {
	dataSourceName := "data.netskope_grepop.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Use hex ID for lon1 (London)
				Config: testAccGREPOPDataSourceConfig_basic("0x00AD"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "pop_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "pop_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "gateway"),
				),
			},
		},
	})
}

// Configuration functions

func testAccGREPOPDataSourceConfig_basic(popID string) string {
	return testAccProviderConfig() + `
data "netskope_grepop" "test" {
  pop_id = "` + popID + `"
}
`
}
