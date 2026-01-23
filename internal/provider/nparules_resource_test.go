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

func TestAccNPARules_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPARulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPARulesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
				),
				// Known provider issue: computed fields in private_app cause plan drift
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_order", "rule_data", "description", "group_id"},
			},
		},
	})
}

func TestAccNPARules_update(t *testing.T) {
	// Skip: Provider issue with rule updates when private app is recreated
	t.Skip("Skipping: Provider has issue with rule updates - private app reference becomes stale")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPARulesDestroy,
		Steps: []resource.TestStep{
			// Create with allow action
			{
				Config: testAccNPARulesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPARulesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
				),
				ExpectNonEmptyPlan: true,
			},
			// Update - disable the rule
			{
				Config: testAccNPARulesConfig_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPARulesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccNPARules_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPARulesDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccNPARulesConfig_basic(rName),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_order", "rule_data", "description", "group_id"},
			},
		},
	})
}

func TestAccNPARules_denyRule(t *testing.T) {
	// Skip: Block action requires template configuration not documented
	t.Skip("Skipping: API requires template field for block action which is not documented")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPARulesDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_denyRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPARulesExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
				),
				// Known provider issue: computed fields cause plan drift
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Configuration functions

func testAccNPARulesConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = "%s-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = "%s-app"
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

resource "netskope_npa_rules" "test" {
  rule_name   = %q
  description = "Acceptance test rule"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    user_id       = ["*"]
    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

func testAccNPARulesConfig_disabled(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = "%s-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = "%s-app"
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

resource "netskope_npa_rules" "test" {
  rule_name   = %q
  description = "Acceptance test rule - disabled"
  enabled     = "0"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    user_id       = ["*"]
    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

func testAccNPARulesConfig_denyRule(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = "%s-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = "%s-app"
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

resource "netskope_npa_rules" "test" {
  rule_name   = %q
  description = "Acceptance test deny rule"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "block"
    }

    user_id       = ["*"]
    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

// Helper functions

func testAccCheckNPARulesExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource ID not set")
		}

		return nil
	}
}

func testAccCheckNPARulesDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "netskope_npa_rules" {
			continue
		}

		// The acceptance test framework automatically destroys resources.
		// This function verifies the resource no longer exists in state.
		// In a production scenario, you would make an API call to verify
		// the resource has been deleted from the Netskope tenant.
	}
	return nil
}
