// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGRETunnelsListDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelsListDataSourceConfig_basic(),
			},
		},
	})
}

// Configuration functions

func testAccGRETunnelsListDataSourceConfig_basic() string {
	return testAccProviderConfig() + `
data "netskope_gre_tunnels_list" "test" {
}
`
}
