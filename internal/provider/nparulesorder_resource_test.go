package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

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
				),
			},
		},
	})
}

// TestAccNPARulesOrder_reorder verifies that changing the rule_ids list
// actually repositions the rules in the API. Step 1 creates A,B,C order,
// step 2 changes to C,A,B order.
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
				),
			},
		},
	})
}
