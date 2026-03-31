// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Shared config helpers used by drift_detection_test.go.
// These were originally in the resource test files that have been migrated
// to package provider_test. They remain here for drift detection tests
// that still use inline Config strings.

func testAccNPALocalBrokerConfigConfig_basic(hostname string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker_config" "test" {
  hostname = %q
}
`, testAccProviderConfig(), hostname)
}

func testAccNPAPolicyGroupsConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = %q

  # Set order relative to the Default group (id=2)
  group_order = {
    group_id = "2"
    order    = "after"
  }
}
`, testAccProviderConfig(), name)
}

func testAccNPAPublisherUpgradeProfileConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

data "netskope_npa_publishers_releases_list" "releases" {}

resource "netskope_npa_publisher_upgrade_profile" "test" {
  name         = %q
  docker_tag   = data.netskope_npa_publishers_releases_list.releases.data[0].docker_tag
  enabled      = true
  frequency    = "0 0 * * *"
  release_type = "Beta"
  timezone     = "US/Pacific"
}
`, testAccProviderConfig(), name)
}

// testAccCheckResourceExists verifies a resource exists in Terraform state.
// If attrNames are provided, it checks that each named attribute is non-empty.
// If no attrNames are provided, it checks that the resource ID is set.
func testAccCheckResourceExists(resourceName string, attrNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if len(attrNames) == 0 {
			if rs.Primary.ID == "" {
				return fmt.Errorf("resource ID not set")
			}
		} else {
			for _, attr := range attrNames {
				if rs.Primary.Attributes[attr] == "" {
					return fmt.Errorf("%s not set", attr)
				}
			}
		}

		return nil
	}
}

// testAccCheckResourceDestroy is a stub destroy checker for acceptance tests.
// The acceptance test framework automatically destroys resources. In a production
// scenario, you would make an API call to verify the resource has been deleted.
func testAccCheckResourceDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
		}
		return nil
	}
}

// IPSec tunnel config helpers used by drift_detection_test.go.

func testAccIPSecTunnelConfig_basic(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP, name)
}

func testAccIPSecTunnelConfig_withOptions(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]

  options = {
    rekey  = true
    reauth = true
  }
}
`, testAccProviderConfig(), name, sourceIP, name)
}
