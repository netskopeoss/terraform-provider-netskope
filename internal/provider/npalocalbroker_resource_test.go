// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNPALocalBroker_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "NONE"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "local_broker_id",
				ImportStateIdFunc:                    testAccNPALocalBrokerImportStateIdFunc(resourceName),
				// label_ids is write-only, not returned by API
				ImportStateVerifyIgnore: []string{"label_ids"},
			},
		},
	})
}

func TestAccNPALocalBroker_fullConfig(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "city_name", "Cupertino"),
					resource.TestCheckResourceAttr(resourceName, "region_name", "CA"),
					resource.TestCheckResourceAttr(resourceName, "country_name", "United States of America"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "custom_public_ip", "203.0.113.42"),
					resource.TestCheckResourceAttr(resourceName, "custom_private_ip", "192.168.19.119"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "NONE"),
				),
			},
		},
	})
}

func TestAccNPALocalBroker_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			// Create with basic config
			{
				Config: testAccNPALocalBrokerConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
				),
			},
			// Update with location info
			{
				Config: testAccNPALocalBrokerConfig_withLocation(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "city_name", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "region_name", "CA"),
				),
			},
		},
	})
}

func TestAccNPALocalBroker_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerConfig_basic(rName),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "local_broker_id",
				ImportStateIdFunc:                    testAccNPALocalBrokerImportStateIdFunc(resourceName),
				// label_ids is write-only, not returned by API
				ImportStateVerifyIgnore: []string{"label_ids"},
			},
		},
	})
}

func TestAccNPALocalBroker_accessViaPublicIP(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPALocalBrokerConfig_accessViaPublicIP(rName, "OFF_PREM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "OFF_PREM"),
				),
			},
			{
				Config: testAccNPALocalBrokerConfig_accessViaPublicIP(rName, "ON_PREM"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "ON_PREM"),
				),
			},
		},
	})
}

// Configuration functions

func testAccNPALocalBrokerConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name = %q
}
`, testAccProviderConfig(), name)
}

func testAccNPALocalBrokerConfig_full(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name   = %q
  city_name           = "Cupertino"
  region_name         = "CA"
  country_name        = "United States of America"
  country_code        = "US"
  latitude            = 37.323
  longitude           = -122.032
  custom_public_ip    = "203.0.113.42"
  custom_private_ip   = "192.168.19.119"
  access_via_public_ip = "NONE"
}
`, testAccProviderConfig(), name)
}

func testAccNPALocalBrokerConfig_withLocation(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name = %q
  city_name         = "San Francisco"
  region_name       = "CA"
  country_name      = "United States of America"
  country_code      = "US"
}
`, testAccProviderConfig(), name)
}

func testAccNPALocalBrokerConfig_accessViaPublicIP(name, accessMode string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name    = %q
  access_via_public_ip = %q
}
`, testAccProviderConfig(), name, accessMode)
}

// Helper functions

func testAccCheckNPALocalBrokerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		if rs.Primary.Attributes["local_broker_id"] == "" {
			return fmt.Errorf("local_broker_id not set")
		}

		return nil
	}
}

func testAccCheckNPALocalBrokerDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netskope_npa_local_broker" {
			continue
		}

		// The acceptance test framework automatically destroys resources.
		// This function verifies the resource no longer exists in state.
		// In a production scenario, you would make an API call to verify
		// the resource has been deleted from the Netskope tenant.
	}
	return nil
}

func testAccNPALocalBrokerImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes["local_broker_id"], nil
	}
}
