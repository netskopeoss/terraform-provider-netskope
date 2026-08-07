package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccRBACRoleConfig_basic(t *testing.T) {
	dataSourceName := "data.netskope_rbac_role_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRBACRoleConfigConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// The catalog is non-empty on any live tenant
					resource.TestCheckResourceAttrSet(dataSourceName, "api_groups.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "obfuscation.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "functions.#"),
					// api_groups entries have the expected fields
					resource.TestCheckResourceAttrSet(dataSourceName, "api_groups.0.api_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "api_groups.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "api_groups.0.display_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "api_groups.0.highest_permission"),
					// obfuscation entries have the expected fields
					resource.TestCheckResourceAttrSet(dataSourceName, "obfuscation.0.api_group_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "obfuscation.0.obfuscation_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "obfuscation.0.obfuscation_label"),
					// functions entries have the expected fields
					resource.TestCheckResourceAttrSet(dataSourceName, "functions.0.name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "functions.0.sub_functions.#"),
				),
			},
		},
	})
}

func testAccRBACRoleConfigConfig() string {
	return `data "netskope_rbac_role_config" "test" {}`
}
