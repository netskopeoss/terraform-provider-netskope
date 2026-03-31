package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccGREPOPDataSource_basic(t *testing.T) {
	dataSourceName := "data.netskope_grepop.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Use hex ID for lon1 (London)
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"pop_id": config.StringVariable("0x00AD"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "pop_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "pop_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "gateway"),
				),
			},
		},
	})
}
