// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPublisher_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "lbrokerconnect", "false"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"upgrade_status"},
				ImportStateVerifyIdentifierAttribute: "publisher_id",
				ImportStateIdFunc:                    testAccImportStateIdFunc(resourceName, "publisher_id"),
			},
		},
	})
}

func TestAccNPAPublisher_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccNPAPublisherConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
				),
			},
			// Update name
			{
				Config: testAccNPAPublisherConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccNPAPublisher_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherConfig_basic(rName),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"upgrade_status"},
				ImportStateVerifyIdentifierAttribute: "publisher_id",
				ImportStateIdFunc:                    testAccImportStateIdFunc(resourceName, "publisher_id"),
			},
		},
	})
}

func TestAccNPAPublisher_withUpgradeProfile(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherConfig_withUpgradeProfile(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_upgrade_profiles_id", "1"),
				),
			},
		},
	})
}

// TestAccNPAPublisher_disappears verifies the resource lifecycle works correctly.
// Note: A true "disappears" test that deletes via API requires the provider's
// Read function to properly handle "resource not found" responses and remove
// the resource from state instead of returning an error. This is tracked as
// a provider enhancement.
func TestAccNPAPublisher_disappears(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
				),
			},
			{
				// Verify the resource can be read and is stable
				Config:   testAccNPAPublisherConfig_basic(rName),
				PlanOnly: true,
			},
		},
	})
}

// Configuration functions

func testAccNPAPublisherConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
}
`, testAccProviderConfig(), name)
}

func testAccNPAPublisherConfig_withUpgradeProfile(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name              = %q
  publisher_upgrade_profiles_id = 1
}
`, testAccProviderConfig(), name)
}
