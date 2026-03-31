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

func TestAccNPARules_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ConfigDirectory:         config.TestNameDirectory(),
				ConfigVariables:         vars,
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Create with allow action
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
				),
			},
			// Update - disable the rule
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
				),
			},
		},
	})
}

func TestAccNPARules_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
			},
			{
				ResourceName:            resourceName,
				ConfigDirectory:         config.TestNameDirectory(),
				ConfigVariables:         vars,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_order", "rule_data", "description", "group_id"},
			},
		},
	})
}

func TestAccNPARules_denyRule(t *testing.T) {
	// Skip: The API has a name/filename mismatch for the template field in match_criteria_action:
	//   - Create/Update requires the template display NAME (e.g. "Default Template")
	//   - GET response returns the template FILE NAME (e.g. "block_page.html")
	// SuppressDiff on the template attribute doesn't prevent the diff because the parent
	// object (match_criteria_action) detects the change. See docs/KNOWN_API_ISSUES.md #13.
	t.Skip("Skipping: API template name/filename mismatch causes perpetual drift (KNOWN_API_ISSUES #13)")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "block"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.emit_alert", "true"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", "Default Template"),
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule1"),
					testutil.CheckResourceExists("netskope_npa_rules.rule2"),
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
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule_a"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_b"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_c"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_a", "rule_name", rName+"-rule-a"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_b", "rule_name", rName+"-rule-b"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_c", "rule_name", rName+"-rule-c"),
				),
			},
		},
	})
}
