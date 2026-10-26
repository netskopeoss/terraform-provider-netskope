package provider_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
)

// Every test here asserts the LIVE order on the tenant via checkNPARulesLiveOrder
// (ListNPARules), not just rule_ids in state, and ends with an empty-plan step in a
// fresh provider process. The order resource's own convergence loop reports failure
// only when it cannot converge; these checks catch the case where it converges to the
// wrong order or where a later refresh would disagree.

// TestAccNPARulesOrder_basic verifies that the rules_order resource
// correctly positions three rules: A before B before C.
func TestAccNPARulesOrder_basic(t *testing.T) {
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
					resource.TestCheckResourceAttrSet("netskope_npa_rules_order.test", "id"),
					resource.TestCheckResourceAttr("netskope_npa_rules_order.test", "rule_ids.#", "3"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b", "netskope_npa_rules.rule_c"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})
}

// TestAccNPARulesOrder_reorder verifies that changing the rule_ids list
// actually repositions the rules on the tenant. Step 1 creates A,B,C order,
// step 2 changes to C,A,B order, step 3 must plan empty.
func TestAccNPARulesOrder_reorder(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Step 1: Create with order A, B, C
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule_a"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_b"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_c"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b", "netskope_npa_rules.rule_c"),
				),
			},
			// Step 2: Reorder to C, A, B
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule_a"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_b"),
					testutil.CheckResourceExists("netskope_npa_rules.rule_c"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_c", "netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b"),
				),
			},
			// Step 3: same config — must plan empty in a fresh process.
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

// TestAccNPARulesOrder_removeRule is a regression test for issue #125: removing
// an ID from rule_ids must leave the remaining rules in order without having to
// re-position all of them.
// Step 1 creates A,B,C in order A,B,C. Step 2 removes B from rule_ids while the
// rule itself still exists. Step 3 also removes rule B's resource. Step 4 must
// plan empty. Live order of A then C is asserted at every step.
func TestAccNPARulesOrder_removeRule(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			// Step 1: Create with order A, B, C
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netskope_npa_rules_order.test", "rule_ids.#", "3"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.0", "netskope_npa_rules.rule_a", "id"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.1", "netskope_npa_rules.rule_b", "id"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.2", "netskope_npa_rules.rule_c", "id"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b", "netskope_npa_rules.rule_c"),
				),
			},
			// Step 2: Remove B from rule_ids; remaining order must be A, C (B unmanaged, still exists)
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists("netskope_npa_rules.rule_b"),
					resource.TestCheckResourceAttr("netskope_npa_rules_order.test", "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.0", "netskope_npa_rules.rule_a", "id"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.1", "netskope_npa_rules.rule_c", "id"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_c"),
				),
			},
			// Step 3: Remove rule B itself; order is unchanged
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("netskope_npa_rules_order.test", "rule_ids.#", "2"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.0", "netskope_npa_rules.rule_a", "id"),
					resource.TestCheckResourceAttrPair("netskope_npa_rules_order.test", "rule_ids.1", "netskope_npa_rules.rule_c", "id"),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_c"),
				),
			},
			// Step 4: same config — must plan empty.
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

// TestAccNPARulesOrder_detectsOutOfBandReorder verifies that an out-of-band reorder
// is detected on the next plan and corrected on apply (#129 P3). Steps:
//  1. Create A,B,C in managed order; assert live order A,B,C.
//  2. PreConfig PATCHes rule C to top via the SDK (simulating an out-of-band change).
//     The pre-apply plan must show netskope_npa_rules_order.test as an update.
//     After apply, live order must be A,B,C again.
//  3. Same config — plan must be empty.
func TestAccNPARulesOrder_detectsOutOfBandReorder(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}
	var ruleCID string

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
					testutil.CheckResourceExists("netskope_npa_rules.rule_c"),
					resource.TestCheckResourceAttrWith("netskope_npa_rules.rule_c", "id", func(v string) error {
						ruleCID = v
						return nil
					}),
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b", "netskope_npa_rules.rule_c"),
				),
			},
			{
				// Out-of-band: PATCH rule C to the top position, simulating what would
				// happen if an operator or another tool reordered the rules outside of
				// Terraform. The next plan must detect the drift and show an update.
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				PreConfig: func() {
					if ruleCID == "" {
						t.Fatal("rule_c ID was not captured in step 1")
					}
					client, err := newTestSDKClient()
					if err != nil {
						t.Fatalf("newTestSDKClient: %v", err)
					}
					order := shared.OrderTop
					res, err := client.UpdateNPARules(context.Background(), operations.UpdateNPARulesRequest{
						ID: ruleCID,
						NpaPolicyRequest: shared.NpaPolicyRequest{
							RuleOrder: &shared.RuleOrder{
								Order: &order,
							},
						},
					})
					if err != nil {
						t.Fatalf("out-of-band PATCH rule_c to top: %v", err)
					}
					if res.StatusCode != 200 {
						t.Fatalf("out-of-band PATCH rule_c to top: status %d", res.StatusCode)
					}
				},
				// The plan must show an update to restore the managed order.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("netskope_npa_rules_order.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPARulesLiveOrder("netskope_npa_rules.rule_a", "netskope_npa_rules.rule_b", "netskope_npa_rules.rule_c"),
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
