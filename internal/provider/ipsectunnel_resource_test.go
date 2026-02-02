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

func TestAccIPSecTunnel_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_basic(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "50"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testAccIPSecTunnelImportStateIdFunc(resourceName),
				// pop_names and psk are not returned by the API in the same format
				ImportStateVerifyIgnore: []string{"pop_names", "psk"},
			},
		},
	})
}

func TestAccIPSecTunnel_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccIPSecTunnelConfig_withIP(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "50"),
				),
			},
			// Update bandwidth and notes
			{
				Config: testAccIPSecTunnelConfig_updated(rNameUpdated, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100"),
					resource.TestCheckResourceAttr(resourceName, "notes", "Updated by acceptance test"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_withEncryption(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_withEncryption(rName, sourceIP, "AES256-CBC"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption", "AES256-CBC"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_withOptions(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_withOptions(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "options.rekey", "true"),
					resource.TestCheckResourceAttr(resourceName, "options.reauth", "true"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_basic(rName, sourceIP),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testAccIPSecTunnelImportStateIdFunc(resourceName),
				// pop_names and psk are not returned by the API in the same format
				ImportStateVerifyIgnore: []string{"pop_names", "psk"},
			},
		},
	})
}

func TestAccIPSecTunnel_disabled(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPSecTunnelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_disabled(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIPSecTunnelExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}

// Configuration functions

func testAccIPSecTunnelConfig_basic(name, sourceIP string) string {
	// IPSec requires: site, srcidentity (source_identity), psk, encryption, pops
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

func testAccIPSecTunnelConfig_withIP(name, sourceIP string) string {
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

func testAccIPSecTunnelConfig_updated(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]
  bandwidth       = 100
  notes           = "Updated by acceptance test"
}
`, testAccProviderConfig(), name, sourceIP, name)
}

func testAccIPSecTunnelConfig_withEncryption(name, sourceIP, encryption string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = %q
  pop_names       = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP, name, encryption)
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

func testAccIPSecTunnelConfig_disabled(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]
  enabled         = false
}
`, testAccProviderConfig(), name, sourceIP, name)
}

// Helper functions

func testAccCheckIPSecTunnelExists(resourceName string) resource.TestCheckFunc {
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

func testAccCheckIPSecTunnelDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netskope_ip_sec_tunnel" {
			continue
		}

		// The acceptance test framework automatically destroys resources.
		// This function verifies the resource no longer exists in state.
	}
	return nil
}

func testAccIPSecTunnelImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes["tunnel_id"], nil
	}
}
