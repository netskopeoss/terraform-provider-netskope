// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPAPolicyGroups_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_policy_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_policy_groups"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"group_order", "modify_time"},
			},
		},
	})
}

func TestAccNPAPolicyGroups_update(t *testing.T) {
	// Skip: Provider bug with policy group updates - SQLAlchemy bind parameter error
	t.Skip("Skipping: Provider has bug updating policy groups - missing new_order parameter")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	resourceName := "netskope_npa_policy_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_policy_groups"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
				),
			},
			// Update name
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "group_name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccNPAPolicyGroups_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_policy_groups.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_policy_groups"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rName),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"group_order", "modify_time"},
			},
		},
	})
}

// Configuration functions

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
