// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccGREPOPsListDataSource_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccGREPOPsListDataSourceConfig_basic(),
			},
		},
	})
}

// Configuration functions

func testAccGREPOPsListDataSourceConfig_basic() string {
	return testAccProviderConfig() + `
data "netskope_grepo_ps_list" "test" {
}
`
}
