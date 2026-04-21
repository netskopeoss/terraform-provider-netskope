package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccDeviceClassificationTagListDataSource_basic(t *testing.T) {
	dataSourceName := "data.netskope_device_classification_tag_list.all"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.0.tag_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "tags.0.name"),
				),
			},
		},
	})
}

func TestAccDeviceClassificationOptionsListDataSource_basic(t *testing.T) {
	dataSourceName := "data.netskope_device_classification_options_list.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "options.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "options.0.key"),
					resource.TestCheckResourceAttrSet(dataSourceName, "options.0.value"),
				),
			},
		},
	})
}
