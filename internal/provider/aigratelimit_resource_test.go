// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

// AIG rate limit names are ≤15 chars with no prefix requirement.
// We use "tfrl" (4 chars) + 10 random = 14 chars total.
func aigRateLimitTestName() string {
	return "tfrl" + acctest.RandString(10)
}

func TestAccAIGRateLimit_basic(t *testing.T) {
	rName := aigRateLimitTestName()
	resourceName := "netskope_aig_rate_limit.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_rate_limit"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "criteria.apply_on", "ai"),
					resource.TestCheckResourceAttr(resourceName, "limit.requests", "100"),
					resource.TestCheckResourceAttr(resourceName, "limit.unit", "hour"),
				),
			},
			{
				ResourceName:      resourceName,
				ConfigDirectory:   config.TestNameDirectory(),
				ConfigVariables:   vars,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAIGRateLimit_update(t *testing.T) {
	rName := aigRateLimitTestName()
	resourceName := "netskope_aig_rate_limit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_rate_limit"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(rName),
					"requests": config.IntegerVariable(100),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "limit.requests", "100"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":     config.StringVariable(rName),
					"requests": config.IntegerVariable(200),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "limit.requests", "200"),
				),
			},
		},
	})
}

func TestAccAIGRateLimitDataSource_basic(t *testing.T) {
	rName := aigRateLimitTestName()
	resourceName := "netskope_aig_rate_limit.test"
	dataSourceName := "data.netskope_aig_rate_limit.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_rate_limit"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "criteria.apply_on", resourceName, "criteria.apply_on"),
					resource.TestCheckResourceAttrPair(dataSourceName, "limit.requests", resourceName, "limit.requests"),
					resource.TestCheckResourceAttrPair(dataSourceName, "limit.unit", resourceName, "limit.unit"),
				),
			},
		},
	})
}

func TestAccAIGRateLimitListDataSource_basic(t *testing.T) {
	rName := aigRateLimitTestName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_rate_limit"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netskope_aig_rate_limit_list.test", "elements.#"),
				),
			},
		},
	})
}
