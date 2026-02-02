// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPALocalBrokerConfig_basic(t *testing.T) {
	rHostname := fmt.Sprintf("lbr-%s.example.com", acctest.RandString(8))
	resourceName := "netskope_npa_local_broker_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerConfigConfig_basic(rHostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
					resource.TestCheckResourceAttr(resourceName, "data.hostname", rHostname),
				),
			},
		},
	})
}

func TestAccNPALocalBrokerConfig_update(t *testing.T) {
	rHostname := fmt.Sprintf("lbr-%s.example.com", acctest.RandString(8))
	rHostnameUpdated := fmt.Sprintf("lbr-%s-updated.example.com", acctest.RandString(8))
	resourceName := "netskope_npa_local_broker_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccNPALocalBrokerConfigConfig_basic(rHostname),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostname),
				),
			},
			// Update hostname
			{
				Config: testAccNPALocalBrokerConfigConfig_basic(rHostnameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "hostname", rHostnameUpdated),
				),
			},
		},
	})
}

// Configuration functions

func testAccNPALocalBrokerConfigConfig_basic(hostname string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker_config" "test" {
  hostname = %q
}
`, testAccProviderConfig(), hostname)
}
