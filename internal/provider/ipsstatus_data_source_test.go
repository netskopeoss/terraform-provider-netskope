package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccIPSStatusDataSource_basic(t *testing.T) {
	dataSourceName := "data.netskope_ips_status.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "web"),
					resource.TestCheckResourceAttrSet(dataSourceName, "nonweb"),
					resource.TestCheckResourceAttrSet(dataSourceName, "npa"),
				),
			},
		},
	})
}
