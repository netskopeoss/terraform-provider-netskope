// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPublisherUpgradeProfile_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher_upgrade_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher_upgrade_profile"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_upgrade_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "release_type", "Beta"),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_upgrade_profile_id"),
				),
			},
		},
	})
}

func TestAccNPAPublisherUpgradeProfile_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher_upgrade_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher_upgrade_profile"),
		Steps: []resource.TestStep{
			// Create with enabled = true
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_upgrade_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
				),
			},
			// Update - disable the profile
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_upgrade_profile_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

func TestAccNPAPublisherUpgradeProfile_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher_upgrade_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher_upgrade_profile"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_basic(rName),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "publisher_upgrade_profile_id",
				ImportStateIdFunc:                    testAccImportStateIdFunc(resourceName, "publisher_upgrade_profile_id"),
				// Some computed fields may not match exactly after import
				ImportStateVerifyIgnore: []string{"next_update_time", "created_at", "updated_at"},
			},
		},
	})
}

// Configuration functions

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

func testAccNPAPublisherUpgradeProfileConfig_disabled(name string) string {
	return fmt.Sprintf(`
%s

data "netskope_npa_publishers_releases_list" "releases" {}

resource "netskope_npa_publisher_upgrade_profile" "test" {
  name         = %q
  docker_tag   = data.netskope_npa_publishers_releases_list.releases.data[0].docker_tag
  enabled      = false
  frequency    = "0 0 * * *"
  release_type = "Beta"
  timezone     = "US/Pacific"
}
`, testAccProviderConfig(), name)
}
