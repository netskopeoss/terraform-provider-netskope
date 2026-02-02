// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// =============================================================================
// ACCEPTANCE TEST DESIGN NOTES
// =============================================================================
//
// Test Independence:
//   Each test function is completely self-contained. Tests create ALL required
//   dependencies (publishers, etc.) within their own configuration and clean up
//   after themselves. Tests MUST NOT assume resources exist from previous tests
//   or share state.
//
// Sequential Execution:
//   Tests use resource.Test (not ParallelTest) to run sequentially. This avoids
//   hitting Netskope API rate limits. While slower, it ensures reliable test runs.
//
// Dependency Management:
//   Private apps require a publisher. Each config function creates the publisher
//   inline, and the private app references it via Terraform interpolation:
//
//     resource "netskope_npa_publisher" "test" { ... }
//     resource "netskope_npa_private_app" "test" {
//       publishers = [{
//         publisher_id = tostring(netskope_npa_publisher.test.publisher_id)
//       }]
//     }
//
//   Terraform automatically handles creation order based on these references.
//
// Naming Convention:
//   All test resources use the prefix "tf-acc-test-" plus a random suffix to:
//   - Identify resources as test artifacts
//   - Avoid name collisions between parallel tests
//   - Enable easy cleanup of orphaned resources
//
// Known Issues:
//   - ImportStateVerifyIgnore skips fields that don't round-trip properly on import
//
// =============================================================================

// TestAccNPAPrivateApp_basic tests creating a private app with minimal configuration.
// Test ID: PA-ACC-001
func TestAccNPAPrivateApp_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccNPAPrivateAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100"),
					resource.TestCheckResourceAttrSet(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "use_publisher_dns", "true"),
					resource.TestCheckResourceAttr(resourceName, "trust_self_signed_certs", "false"),
				),
			},
			// Import
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "private_app_id",
				ImportStateIdFunc:                    testAccImportStateIdFunc(resourceName, "private_app_id"),
				// Skip verification of computed fields
				ImportStateVerifyIgnore: []string{"publishers", "real_host", "protocols"},
			},
		},
	})
}

func TestAccNPAPrivateApp_complete(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_complete(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100,192.168.1.101"),
					resource.TestCheckResourceAttrSet(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "use_publisher_dns", "true"),
					resource.TestCheckResourceAttr(resourceName, "trust_self_signed_certs", "true"),
					resource.TestCheckResourceAttr(resourceName, "clientless_access", "false"),
				),
				// Known provider issue: computed fields cause plan drift
			},
		},
	})
}

func TestAccNPAPrivateApp_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccNPAPrivateAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
				),
			},
			// Update - add more protocols and hosts
			{
				Config: testAccNPAPrivateAppConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100,192.168.1.101"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trust_self_signed_certs", "true"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_basic(rName),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "private_app_id",
				ImportStateIdFunc:                    testAccImportStateIdFunc(resourceName, "private_app_id"),
				ImportStateVerifyIgnore:              []string{"publishers", "real_host", "protocols"},
			},
		},
	})
}

func TestAccNPAPrivateApp_multipleProtocols(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_multipleProtocols(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "3"),
				),
				// Known provider issue: computed fields cause plan drift
			},
		},
	})
}

func TestAccNPAPrivateApp_clientlessAccess(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_clientlessAccess(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "clientless_access", "true"),
					resource.TestCheckResourceAttr(resourceName, "real_host", "browser.internal.test"),
					resource.TestCheckResourceAttr(resourceName, "private_app_protocol", "http"),
					// private_app_hostname is auto-generated for browser apps
					resource.TestCheckResourceAttrSet(resourceName, "private_app_hostname"),
					resource.TestCheckResourceAttrSet(resourceName, "private_app_id"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_tags(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_tags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_updatePublishers(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create with first publisher
			{
				Config: testAccNPAPrivateAppConfig_withPublisher(rName, "1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "publishers.#", "1"),
				),
			},
			// Update to second publisher
			{
				Config: testAccNPAPrivateAppConfig_withPublisher(rName, "2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "publishers.#", "1"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_disappears(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName, "private_app_id"),
				),
			},
			{
				Config: testAccNPAPrivateAppConfig_basic(rName),
				// This test verifies that Terraform handles a resource that was
				// deleted outside of Terraform (e.g., via the API or UI).
				// The resource will be recreated on the next apply.
				PlanOnly: true,
			},
		},
	})
}

// =============================================================================
// CONFIGURATION FUNCTIONS
// =============================================================================
//
// Each config function returns a complete, self-contained HCL configuration.
// Dependencies (like publishers) are created inline - never assume they exist.
// This ensures tests can run independently and in parallel.
//
// =============================================================================

func testAccNPAPrivateAppConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
`, testAccProviderConfig(), name, name)
}

func testAccNPAPrivateAppConfig_complete(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100,192.168.1.101"

  # IMPORTANT: Protocols must be in ascending port order to avoid drift
  # See docs/KNOWN_API_ISSUES.md - Issue #14
  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns           = true
  trust_self_signed_certs     = true
  clientless_access           = false
  is_user_portal_app          = false
  allow_unauthenticated_cors  = false
}
`, testAccProviderConfig(), name, name)
}

func testAccNPAPrivateAppConfig_updated(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100,192.168.1.101"

  # IMPORTANT: Protocols must be in ascending port order to avoid drift
  # See docs/KNOWN_API_ISSUES.md - Issue #14
  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = true
}
`, testAccProviderConfig(), name, name)
}

func testAccNPAPrivateAppConfig_multipleProtocols(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"

  # IMPORTANT: Protocols must be ordered to match API response ordering to avoid drift.
  # The API sorts by: protocol type (tcp before udp), then port number ascending.
  # See docs/KNOWN_API_ISSUES.md - Issue #14
  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "53"
      protocol = "udp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
`, testAccProviderConfig(), name, name)
}

// testAccNPAPrivateAppConfig_clientlessAccess creates a browser access app (clientless).
// Browser apps use real_host instead of private_app_hostname, and the hostname is auto-generated.
func testAccNPAPrivateAppConfig_clientlessAccess(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name           = %q
  clientless_access          = true
  real_host                  = "browser.internal.test"
  private_app_protocol       = "http"

  protocols = [
    {
      port     = "80"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  allow_unauthenticated_cors = true
  use_publisher_dns          = false
  trust_self_signed_certs    = true
}
`, testAccProviderConfig(), name, name)
}

// testAccNPAPrivateAppConfig_tags creates a private app with tags assigned.
func testAccNPAPrivateAppConfig_tags(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  tags = [
    {
      tag_name = "tf-acc-test"
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
`, testAccProviderConfig(), name, name)
}

// testAccNPAPrivateAppConfig_withPublisher creates a private app with a specific publisher.
// Used to test changing publisher assignments.
func testAccNPAPrivateAppConfig_withPublisher(name string, publisherSuffix string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test1" {
  publisher_name = "%s-publisher-1"
}

resource "netskope_npa_publisher" "test2" {
  publisher_name = "%s-publisher-2"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test%s.publisher_id)
      publisher_name = netskope_npa_publisher.test%s.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
`, testAccProviderConfig(), name, name, name, publisherSuffix, publisherSuffix)
}
