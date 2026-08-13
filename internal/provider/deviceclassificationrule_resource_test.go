package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccDeviceClassificationRule_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_classification_rule.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_device_classification_rule"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "rule_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "os", "windows"),
					resource.TestCheckResourceAttrSet(resourceName, "label"),
					resource.TestCheckResourceAttrSet(resourceName, "conditions"),
					resource.TestCheckResourceAttrSet(resourceName, "rule_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "rule_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "rule_id"),
			},
		},
	})
}

func TestAccDeviceClassificationRule_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_classification_rule.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_device_classification_rule"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "rule_id"),
					resource.TestCheckResourceAttr(resourceName, "os", "windows"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "rule_id"),
					resource.TestCheckResourceAttr(resourceName, "os", "mac"),
				),
			},
		},
	})
}

func TestAccDeviceClassificationRuleDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_classification_rule.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "rule_id"),
					resource.TestCheckResourceAttr(dsName, "name", rName),
					resource.TestCheckResourceAttr(dsName, "os", "windows"),
					resource.TestCheckResourceAttrSet(dsName, "label"),
					resource.TestCheckResourceAttrSet(dsName, "conditions"),
				),
			},
		},
	})
}

func TestAccDeviceClassificationRuleListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_classification_rule_list.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "rules.#"),
					resource.TestCheckResourceAttrSet(dsName, "rules.0.rule_id"),
					resource.TestCheckResourceAttrSet(dsName, "rules.0.name"),
					resource.TestCheckResourceAttrSet(dsName, "rules.0.os"),
					resource.TestCheckResourceAttrSet(dsName, "rules.0.label"),
				),
			},
		},
	})
}
