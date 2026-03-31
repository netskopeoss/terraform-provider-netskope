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

func TestAccNPAPublisher_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "lbrokerconnect", "false"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore:              []string{"upgrade_status"},
				ImportStateVerifyIdentifierAttribute: "publisher_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "publisher_id"),
			},
		},
	})
}

func TestAccNPAPublisher_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			// Create
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
				),
			},
			// Update name
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rNameUpdated),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccNPAPublisher_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
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
				ImportStateVerifyIgnore:              []string{"upgrade_status"},
				ImportStateVerifyIdentifierAttribute: "publisher_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "publisher_id"),
			},
		},
	})
}

func TestAccNPAPublisher_withUpgradeProfile(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
				),
			},
			{
				// Verify the resource can be read and is stable
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				PlanOnly:        true,
			},
		},
	})
}
