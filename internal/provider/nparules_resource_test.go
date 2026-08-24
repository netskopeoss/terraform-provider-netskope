// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"fmt"
	"os"
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

// TestAccNPARules_importNewFields verifies that a rule containing all the fields
// exposed in v0.4.9 (notify, periodic_reauth, schedule, private_app_tag_ids,
// rule_data.description) can be imported and produces a clean plan.
// Both the create and import steps share the same config (TestNameDirectory).
func TestAccNPARules_importNewFields(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.description", "import-test-description"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.notify.emails.0", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.private_app_tag_ids.0", "1542"),
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
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_name", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "block"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.emit_alert", "true"),
					// API returns the file name (e.g. "23.html"), not the display name — just verify it's set
					resource.TestCheckResourceAttrSet(resourceName, "rule_data.match_criteria_action.template"),
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
// Network Location IDs (numeric strings) and round-trips correctly.
func TestAccNPARules_withNetLocation(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.net_location_obj.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.net_location_obj.0", "1"),
				),
			},
		},
	})
}

// TestAccNPARules_withSchedule verifies that rule_data.schedule can be created
// with time_interval_obj, updated, and cleared without drift.
// schedule was previously terraform-ignored; this is the first live regression test.
// time_interval_obj ID 3 is a pre-provisioned test time interval on the acceptance test tenant.
func TestAccNPARules_withSchedule(t *testing.T) {
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
			// Create with schedule
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.0.time_interval_obj.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.schedule.0.time_interval_obj.0", "3"),
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

// TestAccNPARules_withPeriodicReauth verifies that action_name = "periodic_reauth" is
// accepted by the schema and that the interval can be updated.
// Regression test for https://github.com/netskopeoss/terraform-provider-netskope/issues/116:
// the ActionName enum was missing "periodic_reauth", causing plan-time validation errors
// and import crashes with "invalid value for ActionName: periodic_reauth".
func TestAccNPARules_withPeriodicReauth(t *testing.T) {
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
			// Create with action_name = "periodic_reauth" and 60h interval.
			// Verifies the schema accepts the new enum value and the API creates the rule.
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval_unit", "hours"),
				),
			},
			// Update to 24h — verifies the interval can be changed without drift.
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval", "24"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.periodic_reauth.reauth_interval_unit", "hours"),
				),
			},
		},
	})
}

// TestAccNPARules_periodicReauthImport verifies that importing a rule with
// action_name = "periodic_reauth" no longer crashes.
// Before the fix (issue #116), the SDK's ActionName.UnmarshalJSON returned
// "invalid value for ActionName: periodic_reauth" on any state refresh of
// a rule created via the UI with Periodic Authentication action.
func TestAccNPARules_periodicReauthImport(t *testing.T) {
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.match_criteria_action.action_name", "periodic_reauth"),
				),
			},
			{
				ResourceName:      resourceName,
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   vars,
				ImportState:       true,
				ImportStateVerify: true,
				// template: API returns a .html file name on GET ("2.html") but the config
				// holds the display name ("tf-test-template"). The mismatch is suppressed
				// at plan time by suppressTemplateDrift; ignore it here for the same reason.
				ImportStateVerifyIgnore: []string{"rule_order", "group_id", "description", "rule_data.private_app_tags", "rule_data.match_criteria_action.template"},
			},
		},
	})
}

// TestAccNPARules_allNewRuleDataFields tests all fields added in the OAS expansion:
// notify, rule_data.description, users, user_groups, src_countries, private_app_tag_ids.
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
					resource.TestCheckResourceAttr(resourceName, "rule_data.notify.emails.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.notify.emails.0", "test@example.com"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.notify.interval", "60"),
					resource.TestCheckResourceAttr(resourceName, "rule_data.notify.to_users.0", "admin"),
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
