// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPALocalBrokerDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	dataSourceName := "data.netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_id",
						resourceName, "local_broker_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_name",
						resourceName, "local_broker_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "access_via_public_ip",
						resourceName, "access_via_public_ip",
					),
				),
			},
		},
	})
}

func TestAccNPALocalBrokerDataSource_withLocation(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	dataSourceName := "data.netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerDataSourceConfig_withLocation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "local_broker_id",
						resourceName, "local_broker_id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "city_name",
						resourceName, "city_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "region_name",
						resourceName, "region_name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "country_name",
						resourceName, "country_name",
					),
				),
			},
		},
	})
}

func testAccNPALocalBrokerDataSourceConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name = %q
}

data "netskope_npa_local_broker" "test" {
  local_broker_id = netskope_npa_local_broker.test.local_broker_id
}
`, testAccProviderConfig(), name)
}

func testAccNPALocalBrokerDataSourceConfig_withLocation(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name = %q
  city_name         = "Cupertino"
  region_name       = "CA"
  country_name      = "United States of America"
  country_code      = "US"
}

data "netskope_npa_local_broker" "test" {
  local_broker_id = netskope_npa_local_broker.test.local_broker_id
}
`, testAccProviderConfig(), name)
}
