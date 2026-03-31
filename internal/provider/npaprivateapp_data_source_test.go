package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPAPrivateAppDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	dataSourceName := "data.netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "private_app_id",
						resourceName, "private_app_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "private_app_name",
						resourceName, "private_app_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "private_app_hostname",
						resourceName, "private_app_hostname",
					),
				),
			},
		},
	})
}
