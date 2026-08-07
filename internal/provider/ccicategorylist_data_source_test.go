package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccCCICategoryList_basic(t *testing.T) {
	testutil.PreCheck(t)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `data "netskope_cci_category_list" "test" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Should return at least one category
					resource.TestCheckResourceAttrSet("data.netskope_cci_category_list.test", "categories.#"),
					// Each entry has a name and a non-zero ID
					resource.TestCheckResourceAttrSet("data.netskope_cci_category_list.test", "categories.0.category_name"),
					resource.TestCheckResourceAttrSet("data.netskope_cci_category_list.test", "categories.0.category_id"),
				),
			},
		},
	})
}
