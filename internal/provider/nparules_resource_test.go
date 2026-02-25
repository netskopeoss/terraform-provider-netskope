// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNPARules_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
				),
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
	// Fixed: The BeforeRequest hook was checking for wrong operation ID "updateNPARulesById"
	// instead of "updateNPARules". This caused brackets not to be added to private app names
	// during updates, resulting in "Private app doesn't exist" errors.
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Create with allow action
			{
				Config: testAccNPARulesConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
				),
			},
			// Update - disable the rule
			{
				Config: testAccNPARulesConfig_disabled(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
				),
			},
		},
	})
}

func TestAccNPARules_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_basic(rName),
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
	// Skip: Block action requires a profile to be specified (DLP Profile or Threat Protection Profile)
	// These profiles must be configured in the tenant before block rules can be created.
	t.Skip("Skipping: Block action requires profile configuration (DLP Profile or Threat Protection Profile)")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_denyRule(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
				),
			},
		},
	})
}

// TestAccNPARules_ruleOrderAfter verifies that creating a rule with
// rule_order = { order = "after", rule_id = <id> } works correctly.
// This is a regression test for BUG-003: the BeforeRequest hook's RuleOrder
// struct had rule_id typed as *string, but the SDK serializes it as *int64.
// The type mismatch caused json.Unmarshal to fail during the hook.
func TestAccNPARules_ruleOrderAfter(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_ruleOrderAfter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.rule1"),
					testAccCheckResourceExists("netskope_npa_rules.rule2"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule1", "rule_name", rName+"-rule1"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule2", "rule_name", rName+"-rule2"),
				),
			},
		},
	})
}

// TestAccNPARules_concurrentCreate simulates the BUG-008 scenario:
// multiple independent rules created in a single apply. Without the
// mutex serializer in hookRuleCreateSerializer.go, the API returns
// "Duplicate entry for key 'PRIMARY'" because concurrent inserts
// race on rule_id generation.
func TestAccNPARules_concurrentCreate(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPARulesConfig_concurrentCreate(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.rule_a"),
					testAccCheckResourceExists("netskope_npa_rules.rule_b"),
					testAccCheckResourceExists("netskope_npa_rules.rule_c"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_a", "rule_name", rName+"-rule-a"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_b", "rule_name", rName+"-rule-b"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_c", "rule_name", rName+"-rule-c"),
				),
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

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

// testAccNPARulesConfig_ruleOrderAfter creates two rules where the second
// references the first via rule_order = { order = "after", rule_id = <id> }.
// This is a regression test for BUG-003.
func testAccNPARulesConfig_ruleOrderAfter(name string) string {
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

resource "netskope_npa_rules" "rule1" {
  rule_name   = "%s-rule1"
  description = "BUG-003 test - first rule"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}

resource "netskope_npa_rules" "rule2" {
  rule_name   = "%s-rule2"
  description = "BUG-003 test - second rule, ordered after first"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order   = "after"
    rule_id = tonumber(netskope_npa_rules.rule1.id)
  }
}
`, testAccProviderConfig(), name, name, name, name, name)
}

// testAccNPARulesConfig_concurrentCreate creates 3 independent rules that share
// the same app and policy group but have NO dependencies between them. Terraform
// will attempt to create all 3 concurrently, exercising the BUG-008 scenario.
func testAccNPARulesConfig_concurrentCreate(name string) string {
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

resource "netskope_npa_rules" "rule_a" {
  rule_name   = "%s-rule-a"
  description = "BUG-008 concurrent test - rule A"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}

resource "netskope_npa_rules" "rule_b" {
  rule_name   = "%s-rule-b"
  description = "BUG-008 concurrent test - rule B"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}

resource "netskope_npa_rules" "rule_c" {
  rule_name   = "%s-rule-c"
  description = "BUG-008 concurrent test - rule C"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name, name, name)
}
