package provider_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccIPSecTunnelsListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	dataSourceName := "data.netskope_ip_sec_tunnels_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(rName),
					"source_ip": config.StringVariable(sourceIP),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "result.#"),
					// BUG-015: Verify pop_names is populated
					resource.TestMatchResourceAttr(dataSourceName, "result.0.pop_names.#", regexp.MustCompile(`^[1-9][0-9]*$`)),
				),
			},
		},
	})
}
