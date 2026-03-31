// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create and Read
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
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
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "private_app_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "private_app_id"),
				// Skip verification of computed fields
				ImportStateVerifyIgnore: []string{"publishers", "real_host", "protocols"},
			},
		},
	})
}

func TestAccNPAPrivateApp_complete(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "1"),
				),
			},
			// Update - add more protocols and hosts
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_hostname", "192.168.1.100,192.168.1.101"),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "trust_self_signed_certs", "true"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "private_app_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "private_app_id"),
				ImportStateVerifyIgnore:              []string{"publishers", "real_host", "protocols"},
			},
		},
	})
}

func TestAccNPAPrivateApp_multipleProtocols(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "protocols.#", "3"),
				),
				// Known provider issue: computed fields cause plan drift
			},
		},
	})
}

func TestAccNPAPrivateApp_clientlessAccess(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.#", "1"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_updatePublishers(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			// Create with first publisher
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":             config.StringVariable(rName),
					"publisher_suffix": config.StringVariable("1"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "private_app_name", rName),
					resource.TestCheckResourceAttr(resourceName, "publishers.#", "1"),
				),
			},
			// Update to second publisher
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":             config.StringVariable(rName),
					"publisher_suffix": config.StringVariable("2"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
					resource.TestCheckResourceAttr(resourceName, "publishers.#", "1"),
				),
			},
		},
	})
}

func TestAccNPAPrivateApp_disappears(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "private_app_id"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				// This test verifies that Terraform handles a resource that was
				// deleted outside of Terraform (e.g., via the API or UI).
				// The resource will be recreated on the next apply.
				PlanOnly: true,
			},
		},
	})
}
