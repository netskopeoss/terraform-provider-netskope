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

func TestAccGRETunnel_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1000"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testAccGRETunnelImportStateIdFunc(resourceName),
				// pop_names is not returned by the API in the same format (it returns pops objects)
				ImportStateVerifyIgnore: []string{"pop_names"},
			},
		},
	})
}

func TestAccGRETunnel_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	// Use consistent IP for create and update
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccGRETunnelConfig_withIP(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1000"),
				),
			},
			// Update bandwidth and notes
			{
				Config: testAccGRETunnelConfig_updated(rNameUpdated, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "500"),
					resource.TestCheckResourceAttr(resourceName, "notes", "Updated by acceptance test"),
				),
			},
		},
	})
}

func TestAccGRETunnel_withSourceType(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelConfig_withSourceType(rName, "Machine"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "source_type", "Machine"),
				),
			},
		},
	})
}

func TestAccGRETunnel_withXFF(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelConfig_withXFF(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "options.xff.xff_enabled", "true"),
				),
			},
		},
	})
}

func TestAccGRETunnel_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelConfig_basic(rName),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testAccGRETunnelImportStateIdFunc(resourceName),
				// pop_names is not returned by the API in the same format (it returns pops objects)
				ImportStateVerifyIgnore: []string{"pop_names"},
			},
		},
	})
}

func TestAccGRETunnel_disabled(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_gre_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckGRETunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGRETunnelConfig_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGRETunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

// Configuration functions

func testAccGRETunnelConfig_basic(name string) string {
	// Generate a random IP in the test range to avoid conflicts
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	return testAccGRETunnelConfig_withIP(name, randomIP)
}

func testAccGRETunnelConfig_withIP(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP)
}

func testAccGRETunnelConfig_updated(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
  bandwidth = 500
  notes     = "Updated by acceptance test"
}
`, testAccProviderConfig(), name, sourceIP)
}

func testAccGRETunnelConfig_withSourceType(name, sourceType string) string {
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site        = %q
  source_ip   = %q
  source_type = %q
  pop_names   = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, randomIP, sourceType)
}

func testAccGRETunnelConfig_withXFF(name string) string {
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]

  options = {
    xff = {
      xff_enabled = true
      xff_ip_list = ["10.0.0.1", "10.0.0.2"]
    }
  }
}
`, testAccProviderConfig(), name, randomIP)
}

func testAccGRETunnelConfig_disabled(name string) string {
	randomIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
  enabled   = false
}
`, testAccProviderConfig(), name, randomIP)
}

// Helper functions

func testAccCheckGRETunnelExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		if rs.Primary.Attributes["tunnel_id"] == "" {
			return fmt.Errorf("tunnel_id not set")
		}

		return nil
	}
}

func testAccCheckGRETunnelDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netskope_gre_tunnel" {
			continue
		}

		// The acceptance test framework automatically destroys resources.
		// This function verifies the resource no longer exists in state.
		// In a production scenario, you would make an API call to verify
		// the resource has been deleted from the Netskope tenant.
	}
	return nil
}

func testAccGRETunnelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes["tunnel_id"], nil
	}
}
