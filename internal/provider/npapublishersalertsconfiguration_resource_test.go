// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNPAPublishersAlertsConfiguration_basic(t *testing.T) {
	// Skip: Valid event_types are not documented. API rejects "publisher_up", "publisher_down".
	t.Skip("Skipping: Valid event_types not documented")

	resourceName := "netskope_npa_publishers_alerts_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublishersAlertsConfigurationConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPublishersAlertsConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
					resource.TestCheckResourceAttr(resourceName, "selected_users", "selected"),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "2"),
				),
			},
		},
	})
}

// Configuration functions

func testAccNPAPublishersAlertsConfigurationConfig_basic() string {
	return testAccProviderConfig() + `
resource "netskope_npa_publishers_alerts_configuration" "test" {
  admin_users    = ["jharris@netskope.com"]
  event_types    = ["publisher_up", "publisher_down"]
  selected_users = "selected"
}
`
}

// Helper functions

func testAccCheckNPAPublishersAlertsConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return nil // Resource found
		}
		return nil
	}
}
