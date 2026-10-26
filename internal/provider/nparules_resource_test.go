// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
	sdk "github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
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

// TestAccNPARules_importNewFields verifies that a rule containing all the fields
// exposed in v0.4.9 (periodic_reauth, schedule, private_app_tag_ids,
// rule_data.description) can be imported and produces a clean plan.
// Both the create and import steps share the same config (TestNameDirectory).
// Requires NETSKOPE_TEST_TIME_INTERVAL_ID and NETSKOPE_TEST_PRIVATE_APP_TAG_ID
// to be set to valid IDs on the tenant; skipped if either is absent.
func TestAccNPARules_importNewFields(t *testing.T) {
	timeIntervalID := os.Getenv("NETSKOPE_TEST_TIME_INTERVAL_ID")
	if timeIntervalID == "" {
		t.Skip("Skipping: NETSKOPE_TEST_TIME_INTERVAL_ID not set (must be a valid time interval ID on the tenant)")
	}
	privateAppTagID := os.Getenv("NETSKOPE_TEST_PRIVATE_APP_TAG_ID")
	if privateAppTagID == "" {
		t.Skip("Skipping: NETSKOPE_TEST_PRIVATE_APP_TAG_ID not set (must be a valid private app tag ID on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name":               config.StringVariable(rName),
		"time_interval_id":   config.StringVariable(timeIntervalID),
		"private_app_tag_id": config.StringVariable(privateAppTagID),
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.description", "import-test-description"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_app_tag_ids.0", privateAppTagID),
				),
			},
			{
				ResourceName:      resourceName,
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   vars,
				ImportState:       true,
				ImportStateVerify: true,
				// rule_order, group_id, description (top-level) are not returned by GET-by-ID.
				// rule_data.private_app_tags is auto-populated by Read (GET) but not by
				// the create response; ImportStateVerify sees [] vs ["acme-mfg"].
				// After a subsequent plan+apply the state converges to the Read value.
				ImportStateVerifyIgnore: []string{"rule_order", "group_id", "description", "rule_data.private_app_tags"},
			},
		},
	})
}

// TestAccNPARules_denyRule verifies that a block rule can be created and converges
// to a stable plan after one converging apply.
//
// Three steps (A8 from the test plan):
//  1. Create — API returns the .html filename; state holds filename; no private state yet.
//  2. Converging apply — plan shows "block_page.html → Default Template"; apply persists
//     private state; state now holds display name.
//  3. Plan-only — branch 2 suppresses; plan must be empty.
//
// Regression guard for https://github.com/netskopeoss/terraform-provider-netskope/issues/79
// (perpetual diff) and RC-2 (403 tenants never tested under update path).
func TestAccNPARules_denyRule(t *testing.T) {
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
			// Step 1: create. Private state is empty after Create (§4.3). Convergence
			// check refreshes state from API (returns .html filename) then plans a diff.
			// ExpectNonEmptyPlan: true acknowledges this one-time diff.
			{
				ConfigDirectory:    config.TestNameDirectory(),
				ConfigVariables:    vars,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					// The post-apply refresh plan is the documented one-time converging diff
					// (block_page.html -> display name). ExpectNonEmptyPlan alone would accept
					// ANY diff; these checks pin it to exactly one update on the rule, with the
					// template planned to the display name, and no action on anything else.
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_data").AtMapKey("match_criteria_action").AtMapKey("template"), knownvalue.StringExact("Default Template")),
						plancheck.ExpectResourceAction("netskope_npa_private_app.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_publisher.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_policy_groups.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "block"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.emit_alert", "true"),
					// Create response returns the .html filename; display name comes after convergence.
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 2: converging apply. Plan shows "block_page.html → Default Template" (branch 3).
			// Apply succeeds and private state is set. On 200 tenants AfterSuccess translates
			// the filename so state holds the display name; on 403 tenants state holds the
			// filename — either way the apply must not error.
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 3: plan must be empty — branch 2 suppresses the filename↔display-name diff
			// because private state now holds "Default Template". This is the convergence proof.
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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

func TestAccNPARules_ruleOrderBottom(t *testing.T) {
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

func TestAccNPARules_ruleOrderBefore(t *testing.T) {
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

// TestAccNPARules_withDeviceClassification verifies that a rule with
// device_classification_id works end-to-end, including:
// - The BeforeRequest hook string→int conversion
// - The plan modifier normalization on refresh
func TestAccNPARules_withDeviceClassification(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.device_classification_id.#", "1"),
				),
			},
		},
	})
}

// TestAccNPARules_withClassification verifies that a rule with
// rule_data.classification = ["unmanaged"] can be created, read (state shows
// the classification value), updated without losing classification, and
// imported. This is the live regression test for BUG-018.
func TestAccNPARules_withClassification(t *testing.T) {
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
			// Create with classification = ["unmanaged"]
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.classification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.classification.0", "unmanaged"),
				),
			},
			// Update (disable) — classification must survive
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.classification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.classification.0", "unmanaged"),
				),
			},
			// Import — must not crash and must round-trip classification
			{
				ResourceName:            resourceName,
				ConfigDirectory:         config.TestStepDirectory(),
				ConfigVariables:         vars,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_order", "rule_data", "description", "group_id"},
			},
		},
	})
}

// TestAccNPARules_withNetLocation verifies that net_location_obj accepts
// Network Location IDs and round-trips correctly.
// Requires NETSKOPE_TEST_NET_LOCATION_ID to be set to a valid ID on the tenant.
func TestAccNPARules_withNetLocation(t *testing.T) {
	netLocationID := os.Getenv("NETSKOPE_TEST_NET_LOCATION_ID")
	if netLocationID == "" {
		t.Skip("Skipping: NETSKOPE_TEST_NET_LOCATION_ID not set (must be a valid network location ID on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name":            config.StringVariable(rName),
		"net_location_id": config.StringVariable(netLocationID),
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.net_location_obj.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.net_location_obj.0", netLocationID),
				),
			},
		},
	})
}

// TestAccNPARules_withSchedule verifies that rule_data.schedule can be created
// with time_interval_obj, updated, and cleared without drift.
// schedule was previously terraform-ignored; this is the first live regression test.
// Requires NETSKOPE_TEST_TIME_INTERVAL_ID to be set to a valid time interval ID on the tenant.
func TestAccNPARules_withSchedule(t *testing.T) {
	timeIntervalID := os.Getenv("NETSKOPE_TEST_TIME_INTERVAL_ID")
	if timeIntervalID == "" {
		t.Skip("Skipping: NETSKOPE_TEST_TIME_INTERVAL_ID not set (must be a valid time interval ID on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name":             config.StringVariable(rName),
		"time_interval_id": config.StringVariable(timeIntervalID),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Create with schedule
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.0.time_interval_obj.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.0.time_interval_obj.0", timeIntervalID),
				),
			},
			// Omit schedule from config — plan modifier preserves state, no diff expected.
			// Note: schedule cannot be cleared via empty list because the hook's omitempty
			// silently drops empty slices; the API preserves the existing schedule.
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					// schedule preserved from state — still has the original entry
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.#", "1"),
				),
			},
		},
	})
}

// TestAccNPARules_withPeriodicReauth verifies the periodic_reauth lifecycle end to end.
//
// The API stores a periodic_reauth template DISPLAY NAME verbatim and rejects filenames
// (both test tenants, 2026-10-21 — see docs/api-audit/npa-rules-template). The provider
// therefore sends the configured name unchanged; GET returns the same string, so state
// equals config and there is no converging step. NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE
// must be a display name that exists on the tenant.
//
// Steps: 1 create (backend holds the name); 2 change the interval — template must survive
// the update; 3 plan must be empty.
//
// Regression test for #116 (enum) and #118 (template lost/translated on write).
func TestAccNPARules_withPeriodicReauth(t *testing.T) {
	templateName := os.Getenv("NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE")
	if templateName == "" {
		t.Skip("Skipping: NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE not set (a periodic_reauth template display name on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name":     config.StringVariable(rName),
		"template": config.StringVariable(templateName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", templateName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval_unit", "hours"),
					checkNPARuleBackendTemplateEquals(resourceName, templateName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", templateName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "24"),
					checkNPARuleBackendTemplateEquals(resourceName, templateName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccNPARules_periodicReauthImport verifies import of a provider-created
// periodic_reauth rule round-trips cleanly: the API returns the display name verbatim,
// so the imported state matches config and the first plan after import is empty.
//
// Note: a rule created in the Netskope UI holds a FILENAME (e.g. "3.html") rather than
// a name, because the UI writes through a different path. Importing such a rule shows a
// one-time "3.html" -> "<name>" diff; applying it stores the display name. That case is
// documented in README "Periodic re-authentication rules" and is not exercised here.
//
// Regression test for #116 (import crash on periodic_reauth) and #118.
func TestAccNPARules_periodicReauthImport(t *testing.T) {
	templateName := os.Getenv("NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE")
	if templateName == "" {
		t.Skip("Skipping: NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE not set (a periodic_reauth template display name on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	dir := config.TestNameDirectory()
	vars := config.Variables{
		"name":     config.StringVariable(rName),
		"template": config.StringVariable(templateName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: dir,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", templateName),
					checkNPARuleBackendTemplateEquals(resourceName, templateName),
				),
			},
			{
				ResourceName:      resourceName,
				ConfigDirectory:   dir,
				ConfigVariables:   vars,
				ImportState:       true,
				ImportStateVerify: true,
				// template is verified: the API returns the display name, so it must round-trip.
				ImportStateVerifyIgnore: []string{"rule_order", "group_id", "description", "rule_data.private_app_tags"},
			},
			{
				ConfigDirectory: dir,
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccNPARules_allNewRuleDataFields tests all fields added in the OAS expansion:
// rule_data.description, users, user_groups, src_countries, private_app_tag_ids.
// Mirrors a reference policy configuration on the acceptance test tenant.
// The update step removes users/user_groups and adds a second src_country.
func TestAccNPARules_allNewRuleDataFields(t *testing.T) {
	testUser := os.Getenv("NETSKOPE_TEST_USER")
	if testUser == "" {
		t.Skip("Skipping: NETSKOPE_TEST_USER not set (must be a valid user email on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	vars := config.Variables{
		"name":      config.StringVariable(rName),
		"test_user": config.StringVariable(testUser),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Create with full set of new fields
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.description", "rule-data-description-v1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.users.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.user_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.src_countries.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.src_countries.0", "AL"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_app_tag_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_app_tag_ids.0", "1542"),
				),
			},
			// Update: change description, add second src_country.
			// Note: users/user_groups are retained — the hook uses omitempty so empty
			// arrays are dropped from PUT, preventing the API from clearing them.
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.description", "rule-data-description-v2"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.users.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.user_groups.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.src_countries.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_app_tag_ids.#", "1"),
				),
			},
		},
	})
}

// TestAccNPARules_updateFields verifies that updating rule_name, description,
// and adding private_apps works correctly through the update path.
// TestAccNPARules_withOS verifies that rule_data.os = ["Android", "iOS"] can be
// created, read (state shows the os values), updated without losing os, and
// imported correctly.
func TestAccNPARules_withOS(t *testing.T) {
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
			// Create with os = ["Android", "iOS"]
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.os.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_data.os.*", "Android"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_data.os.*", "iOS"),
				),
			},
			// Update (disable) — os must survive
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.os.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_data.os.*", "Android"),
					resource.TestCheckTypeSetElemAttr(resourceName, "rule_data.os.*", "iOS"),
				),
			},
			// Import — must not crash and must round-trip os
			{
				ResourceName:            resourceName,
				ConfigDirectory:         config.TestStepDirectory(),
				ConfigVariables:         vars,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_order", "rule_data", "description", "group_id"},
			},
		},
	})
}

func TestAccNPARules_updateFields(t *testing.T) {
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
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName+"-original"),
					resource.TestCheckResourceAttr(resourceName, "description", "Original description"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_apps.#", "1"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_apps.#", "2"),
				),
			},
		},
	})
}

// TestAccNPARules_blockRuleUpdate verifies in-place updates of a block rule succeed on
// an ALREADY-CONVERGED rule — the customer path in #124 (rules existed, state held the
// .html filename, then another field changed), which exercises the plan modifier's
// "state is filename, private state matches config, other attributes changed" branch.
//
// Steps:
//  1. Create. Private state is NOT forwarded from create (§4.3, Option A). The refresh
//     that follows apply returns the .html filename, so the convergence check sees a
//     diff. ExpectNonEmptyPlan: true acknowledges this one-time diff.
//  2. Same config — applies the one-time converging PATCH (Branch 3: lastSent is empty,
//     plan carries the display name). After apply, private state holds the display name
//     and the convergence check is empty.
//  3. Disable + change description — the #124 failure path. otherChanges=true causes
//     Branch 2 to return the display name (not suppress), so the PATCH carries the
//     display name and must not fail with "template field is required".
//  4. Same config as 3 — plan must be empty.
func TestAccNPARules_blockRuleUpdate(t *testing.T) {
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
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				// After Create, private state is empty (§4.3). The next plan (convergence
				// check) refreshes state from the API (which returns the .html filename) and
				// then runs a plan where Branch 3 fires — expected one-time non-empty plan.
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					// The post-apply refresh plan is the documented one-time converging diff
					// (block_page.html -> display name). ExpectNonEmptyPlan alone would accept
					// ANY diff; these checks pin it to exactly one update on the rule, with the
					// template planned to the display name, and no action on anything else.
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_data").AtMapKey("match_criteria_action").AtMapKey("template"), knownvalue.StringExact("Default Template")),
						plancheck.ExpectResourceAction("netskope_npa_private_app.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_publisher.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_policy_groups.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "block"),
					// State holds the planned display name after create: the generated Create
					// calls refreshPlan which overwrites the API filename with the plan value.
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", "Default Template"),
					checkNPARuleBackendTemplateIsFilename(resourceName),
				),
			},
			{
				// One-time converging PATCH: Branch 3 fires (lastSent is empty — private
				// state not forwarded from Create). After apply, private state holds
				// "Default Template" and subsequent plans are empty.
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
					checkNPARuleBackendTemplateIsFilename(resourceName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "0"),
					resource.TestCheckResourceAttr(resourceName, "description", "Block rule update test — step 3"),
					checkNPARuleBackendTemplateIsFilename(resourceName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccNPARules_updateWithIgnoredRuleOrder verifies that updating a rule that has
// lifecycle { ignore_changes = [rule_order] } does not move it (#129 P2).
//
// rule_b is created "after" rule_a, so the live order is [rule_a, rule_b]. Position is
// asserted via the rules API after: the create (1), the converging apply (2) — which on
// <=0.4.12 was itself an update that re-sent rule_order — and a description-only
// update of rule_a on the converged rule (3).
func TestAccNPARules_updateWithIgnoredRuleOrder(t *testing.T) {
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
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule_a"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_b"),
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_a", "rule_name", rName+"-a"),
					checkNPARulePosition("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPARulePosition("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netskope_npa_rules.rule_a", "description", "ignore_changes rule_order — step 2"),
					checkNPARulePosition("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccNPARules_updateBlockRuleKeepsPosition is the block-rule analogue of
// TestAccNPARules_updateWithIgnoredRuleOrder (#129 P2). It adds coverage for the
// case where a block rule (with template + emit_alert) has lifecycle { ignore_changes
// = [rule_order] } — the converging PATCH in step 2 and the description update in
// step 3 must not re-send rule_order and must not move the rule relative to rule_b.
//
// Steps:
//  1. Create: rule_a = block (bottom, ignore_changes), rule_b = allow (after rule_a).
//     After create, private state is empty; the convergence check fires a one-time diff.
//     ExpectNonEmptyPlan: true with PostApplyPostRefresh checks pinning the diff.
//  2. Same config — converging PATCH. Position checked after.
//  3. Change rule_a description — must not reorder. Position checked after.
//  4. Same config as 3 — plan must be empty.
func TestAccNPARules_updateBlockRuleKeepsPosition(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	ruleA := "netskope_npa_rules.rule_a"
	ruleB := "netskope_npa_rules.rule_b"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				// One-time converging diff: block_page.html -> "Default Template".
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(ruleA, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(ruleA, tfjsonpath.New("rule_data").AtMapKey("match_criteria_action").AtMapKey("template"), knownvalue.StringExact("Default Template")),
						plancheck.ExpectResourceAction(ruleB, plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_private_app.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_publisher.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_policy_groups.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(ruleA),
					testutil.CheckResourceExists(ruleB),
					resource.TestCheckResourceAttr(ruleA, "rule_data.match_criteria_action.action_name", "block"),
					resource.TestCheckResourceAttr(ruleA, "rule_data.match_criteria_action.template", "Default Template"),
					checkNPARuleBackendTemplateIsFilename(ruleA),
					checkNPARulePosition(ruleA, ruleB),
				),
			},
			{
				// Converging PATCH: applies the one-time template update.
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPARuleBackendTemplateIsFilename(ruleA),
					checkNPARulePosition(ruleA, ruleB),
				),
			},
			{
				// Description change on converged block rule — must not reorder.
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ruleA, "description", "block rule keeps position — step 3 updated"),
					checkNPARuleBackendTemplateIsFilename(ruleA),
					checkNPARulePosition(ruleA, ruleB),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// templateIsFilename matches a .html filename as returned by the API.
var templateIsFilename = regexp.MustCompile(`\.html$`)

// newTestSDKClient builds an SDK client from the same env vars the provider uses,
// for assertions that must read the tenant directly rather than Terraform state.
func newTestSDKClient() (*sdk.TerraformProviderNs, error) {
	serverURL := os.Getenv("NETSKOPE_SERVER_URL")
	if serverURL == "" {
		return nil, fmt.Errorf("NETSKOPE_SERVER_URL not set")
	}
	return sdk.New(sdk.WithServerURL(serverURL), sdk.WithSecurity(shared.Security{APIKey: os.Getenv("NETSKOPE_API_KEY")})), nil
}

// checkNPARuleBackendTemplateIsFilename GETs the rule from the API and asserts the
// stored template is a .html filename. This is the authoritative check that the
// backend holds a resolvable template — the #118 symptom was a display-name string
// stored verbatim, which Terraform state alone cannot distinguish from success.
func checkNPARuleBackendTemplateIsFilename(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[resourceName]
		if rs == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		client, err := newTestSDKClient()
		if err != nil {
			return err
		}
		res, err := client.GetNPARules(context.Background(), operations.GetNPARulesRequest{ID: rs.Primary.ID})
		if err != nil {
			return fmt.Errorf("GetNPARules %s: %w", rs.Primary.ID, err)
		}
		if res.StatusCode != 200 || res.Object == nil || res.Object.Data == nil || res.Object.Data.RuleData == nil ||
			res.Object.Data.RuleData.MatchCriteriaAction == nil || res.Object.Data.RuleData.MatchCriteriaAction.Template == nil {
			return fmt.Errorf("GetNPARules %s: status %d or template missing in response", rs.Primary.ID, res.StatusCode)
		}
		got := *res.Object.Data.RuleData.MatchCriteriaAction.Template
		if !templateIsFilename.MatchString(got) {
			return fmt.Errorf("backend template for %s is %q — not a .html filename; the API stored an unresolvable display name (#118)", resourceName, got)
		}
		return nil
	}
}

// checkNPARuleBackendTemplateEquals GETs the rule from the API and asserts the stored
// template equals want. For periodic_reauth rules the API stores the display name
// verbatim (verified on bespin and policydemo, 2026-10-21), so want is the display name.
func checkNPARuleBackendTemplateEquals(resourceName, want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs := s.RootModule().Resources[resourceName]
		if rs == nil {
			return fmt.Errorf("resource %s not found in state", resourceName)
		}
		client, err := newTestSDKClient()
		if err != nil {
			return err
		}
		res, err := client.GetNPARules(context.Background(), operations.GetNPARulesRequest{ID: rs.Primary.ID})
		if err != nil {
			return fmt.Errorf("GetNPARules %s: %w", rs.Primary.ID, err)
		}
		if res.StatusCode != 200 || res.Object == nil || res.Object.Data == nil || res.Object.Data.RuleData == nil ||
			res.Object.Data.RuleData.MatchCriteriaAction == nil || res.Object.Data.RuleData.MatchCriteriaAction.Template == nil {
			return fmt.Errorf("GetNPARules %s: status %d or template missing in response", rs.Primary.ID, res.StatusCode)
		}
		if got := *res.Object.Data.RuleData.MatchCriteriaAction.Template; got != want {
			return fmt.Errorf("backend template for %s is %q, want %q", resourceName, got, want)
		}
		return nil
	}
}

// checkNPARulesLiveOrder asserts the given rules appear in the live policy list in
// the given order (top to bottom). Generalises checkNPARulePosition for the
// netskope_npa_rules_order tests, whose state-only assertions cannot see the tenant.
func checkNPARulesLiveOrder(resourceNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client, err := newTestSDKClient()
		if err != nil {
			return err
		}
		res, err := client.ListNPARules(context.Background(), operations.ListNPARulesRequest{})
		if err != nil {
			return fmt.Errorf("ListNPARules: %w", err)
		}
		if res.StatusCode != 200 || res.NpaPolicyResponse == nil {
			return fmt.Errorf("ListNPARules returned %d", res.StatusCode)
		}
		pos := map[string]int{}
		for i, rule := range res.NpaPolicyResponse.Data {
			if rule.ID != nil {
				pos[*rule.ID] = i
			}
		}
		last := -1
		for _, name := range resourceNames {
			rs := s.RootModule().Resources[name]
			if rs == nil {
				return fmt.Errorf("resource %s not found in state", name)
			}
			p, ok := pos[rs.Primary.ID]
			if !ok {
				return fmt.Errorf("%s (id=%s) not found in live rule list", name, rs.Primary.ID)
			}
			if p <= last {
				return fmt.Errorf("live order violated at %s (id=%s, pos=%d): expected after pos %d", name, rs.Primary.ID, p, last)
			}
			last = p
		}
		return nil
	}
}

// checkNPARulePosition calls ListNPARules and verifies that the rule named ruleAbove
// appears before the rule named ruleBelow in the live policy evaluation list.
// IDs are read from the Terraform state; the API is called directly via an SDK client
// constructed from the same env vars as the provider.
//
// This is the authoritative position check for the #129 P2 regression: if rule_order
// is re-sent on update, the rule moves in the live list and this check fails.
func checkNPARulePosition(ruleAbove, ruleBelow string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsAbove := s.RootModule().Resources[ruleAbove]
		if rsAbove == nil {
			return fmt.Errorf("resource %s not found in state", ruleAbove)
		}
		rsBelow := s.RootModule().Resources[ruleBelow]
		if rsBelow == nil {
			return fmt.Errorf("resource %s not found in state", ruleBelow)
		}
		idAbove := rsAbove.Primary.ID
		idBelow := rsBelow.Primary.ID

		serverURL := os.Getenv("NETSKOPE_SERVER_URL")
		if serverURL == "" {
			return fmt.Errorf("NETSKOPE_SERVER_URL not set")
		}
		security := shared.Security{APIKey: os.Getenv("NETSKOPE_API_KEY")}
		client := sdk.New(sdk.WithServerURL(serverURL), sdk.WithSecurity(security))

		res, err := client.ListNPARules(context.Background(), operations.ListNPARulesRequest{})
		if err != nil {
			return fmt.Errorf("ListNPARules: %w", err)
		}
		if res.StatusCode != 200 || res.NpaPolicyResponse == nil {
			return fmt.Errorf("ListNPARules returned %d", res.StatusCode)
		}

		posAbove, posBelow := -1, -1
		for i, rule := range res.NpaPolicyResponse.Data {
			if rule.ID == nil {
				continue
			}
			switch *rule.ID {
			case idAbove:
				posAbove = i
			case idBelow:
				posBelow = i
			}
		}
		if posAbove == -1 {
			return fmt.Errorf("%s (id=%s) not found in live rule list", ruleAbove, idAbove)
		}
		if posBelow == -1 {
			return fmt.Errorf("%s (id=%s) not found in live rule list", ruleBelow, idBelow)
		}
		if posAbove >= posBelow {
			return fmt.Errorf("%s (id=%s, pos=%d) is not above %s (id=%s, pos=%d) in live list — rule_order may have been re-sent on update",
				ruleAbove, idAbove, posAbove, ruleBelow, idBelow, posBelow)
		}
		return nil
	}
}

// TestAccNPARules_blockRuleTemplateChange verifies that changing the block rule
// template display name in config is detected as a real change and applied — no
// suppression via private state when the user genuinely changes the template.
//
// Regression test for https://github.com/netskopeoss/terraform-provider-netskope/issues/118
// (second complaint: "Reauth A" → "Reauth B" planned as no-op — RC-4).
//
// Requires NETSKOPE_TEST_BLOCK_TEMPLATE (default "Default Template") and
// NETSKOPE_TEST_BLOCK_TEMPLATE_2; skipped if _2 is not set.
//
// Four steps:
//  1. Create block rule with TEMPLATE. Private state is NOT forwarded from Create (§4.3),
//     so the convergence check shows a one-time diff. ExpectNonEmptyPlan: true.
//  2. Same config — applies the converging PATCH (Branch 3). After apply, private state
//     holds TEMPLATE and subsequent plans are empty.
//  3. Change to TEMPLATE_2 — branch 3 fires (lastSent=TEMPLATE ≠ cfg=TEMPLATE_2), plan is non-empty.
//  4. Same config as step 3 — plan must be empty (branch 2 suppresses after step 3).
func TestAccNPARules_blockRuleTemplateChange(t *testing.T) {
	template1 := os.Getenv("NETSKOPE_TEST_BLOCK_TEMPLATE")
	if template1 == "" {
		template1 = "Default Template"
	}
	template2 := os.Getenv("NETSKOPE_TEST_BLOCK_TEMPLATE_2")
	if template2 == "" {
		t.Skip("Skipping: NETSKOPE_TEST_BLOCK_TEMPLATE_2 not set (must be a second distinct block template display name on the tenant)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	dir := config.TestNameDirectory()
	vars1 := config.Variables{
		"name":     config.StringVariable(rName),
		"template": config.StringVariable(template1),
	}
	vars2 := config.Variables{
		"name":     config.StringVariable(rName),
		"template": config.StringVariable(template2),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Step 1: create with template1. Private state is empty after Create (§4.3).
			// Convergence check sees diff; ExpectNonEmptyPlan: true acknowledges it.
			{
				ConfigDirectory:    dir,
				ConfigVariables:    vars1,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					// The post-apply refresh plan is the documented one-time converging diff
					// (block_page.html -> display name). ExpectNonEmptyPlan alone would accept
					// ANY diff; these checks pin it to exactly one update on the rule, with the
					// template planned to the display name, and no action on anything else.
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_data").AtMapKey("match_criteria_action").AtMapKey("template"), knownvalue.StringExact(template1)),
						plancheck.ExpectResourceAction("netskope_npa_private_app.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_publisher.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_policy_groups.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "block"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 2: converging PATCH — Branch 3 fires (lastSent empty), PATCH carries
			// template1. After apply, private state holds template1.
			{
				ConfigDirectory: dir,
				ConfigVariables: vars1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 3: change to template2. Plan must be non-empty (real change, not suppressed).
			{
				ConfigDirectory: dir,
				ConfigVariables: vars2,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", template2),
				),
			},
			// Step 4: same config as step 3 — plan must be empty (branch 2 suppresses).
			{
				ConfigDirectory: dir,
				ConfigVariables: vars2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccNPARules_importBlockRule verifies the full import lifecycle for a block rule:
//  1. Create (state holds .html filename; no private state yet)
//  2. Apply converging diff (branch 3: display name → private state set)
//  3. Import (private state erased; state holds .html filename again)
//  4. Apply converging diff again (branch 3 re-fires; private state re-set)
//  5. Plan only — must be empty (branch 2 suppresses)
//
// Regression test for https://github.com/netskopeoss/terraform-provider-netskope/issues/105
// and the import scenario in docs/bugs/BUG-021.
func TestAccNPARules_importBlockRule(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	dir := config.TestNameDirectory()
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Step 1: create block rule. Private state is empty after Create (§4.3).
			// Convergence check sees diff; ExpectNonEmptyPlan: true acknowledges it.
			{
				ConfigDirectory:    dir,
				ConfigVariables:    vars,
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					// The post-apply refresh plan is the documented one-time converging diff
					// (block_page.html -> display name). ExpectNonEmptyPlan alone would accept
					// ANY diff; these checks pin it to exactly one update on the rule, with the
					// template planned to the display name, and no action on anything else.
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
						plancheck.ExpectKnownValue(resourceName, tfjsonpath.New("rule_data").AtMapKey("match_criteria_action").AtMapKey("template"), knownvalue.StringExact("Default Template")),
						plancheck.ExpectResourceAction("netskope_npa_private_app.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_publisher.test", plancheck.ResourceActionNoop),
						plancheck.ExpectResourceAction("netskope_npa_policy_groups.test", plancheck.ResourceActionNoop),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 2: apply converging PATCH (Branch 3: lastSent empty) — persists private state.
			// After apply, private state holds the display name; convergence proven by step 5.
			{
				ConfigDirectory: dir,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 3: import — private state is erased; imported state holds the .html filename.
			// ImportStateVerify is disabled for template because the pre-import state holds the
			// display name while the post-import state holds the .html filename (no private state).
			{
				ConfigDirectory:         dir,
				ConfigVariables:         vars,
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"rule_data.match_criteria_action.template", "rule_order", "group_id", "description"},
			},
			// Step 4: apply converging diff again (documented one-time diff after import, §4.6-2).
			// Same 403/200 tenant caveat as step 2: template attribute is set but value varies.
			{
				ConfigDirectory: dir,
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
				),
			},
			// Step 5: plan must be empty — private state re-set; branch 2 suppresses.
			{
				ConfigDirectory: dir,
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccNPARules_periodicReauthTemplateChange verifies that changing the template on a
// periodic_reauth rule is detected and applied (RC-4 regression from #118) and that the
// backend holds the new name afterwards.
//
// Requires NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE and _2 (two display names).
// Steps: 1 create with T1; 2 change to T2 (update action, backend == T2); 3 plan empty.
func TestAccNPARules_periodicReauthTemplateChange(t *testing.T) {
	template1 := os.Getenv("NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE")
	if template1 == "" {
		t.Skip("Skipping: NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE not set")
	}
	template2 := os.Getenv("NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE_2")
	if template2 == "" {
		t.Skip("Skipping: NETSKOPE_TEST_PERIODIC_REAUTH_TEMPLATE_2 not set (must be a second distinct periodic_reauth template display name)")
	}

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_rules.test"
	dir := config.TestNameDirectory()
	vars1 := config.Variables{"name": config.StringVariable(rName), "template": config.StringVariable(template1)}
	vars2 := config.Variables{"name": config.StringVariable(rName), "template": config.StringVariable(template2)}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: dir,
				ConfigVariables: vars1,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", template1),
					checkNPARuleBackendTemplateEquals(resourceName, template1),
				),
			},
			{
				ConfigDirectory: dir,
				ConfigVariables: vars2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.template", template2),
					checkNPARuleBackendTemplateEquals(resourceName, template2),
				),
			},
			{
				ConfigDirectory: dir,
				ConfigVariables: vars2,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}
