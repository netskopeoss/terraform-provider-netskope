// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPublisherDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"
	dataSourceName := "data.netskope_npa_publisher.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "publisher_id",
						resourceName, "publisher_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "publisher_name",
						resourceName, "publisher_name",
					),
				),
			},
		},
	})
}

func testAccNPAPublisherDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
}

data "netskope_npa_publisher" "test" {
  publisher_id = netskope_npa_publisher.test.publisher_id
}
`, testAccProviderConfig(), name)
}