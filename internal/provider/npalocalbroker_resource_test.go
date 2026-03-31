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

func TestAccNPALocalBroker_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "NONE"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "local_broker_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "local_broker_id"),
				// label_ids is write-only, not returned by API
				ImportStateVerifyIgnore: []string{"label_ids"},
			},
		},
	})
}

func TestAccNPALocalBroker_fullConfig(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "city_name", "Cupertino"),
					resource.TestCheckResourceAttr(resourceName, "region_name", "CA"),
					resource.TestCheckResourceAttr(resourceName, "country_name", "United States of America"),
					resource.TestCheckResourceAttr(resourceName, "country_code", "US"),
					resource.TestCheckResourceAttr(resourceName, "custom_public_ip", "203.0.113.42"),
					resource.TestCheckResourceAttr(resourceName, "custom_private_ip", "192.168.19.119"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "NONE"),
				),
			},
		},
	})
}

func TestAccNPALocalBroker_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			// Create with basic config
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
				),
			},
			// Update with location info
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "local_broker_name", rName),
					resource.TestCheckResourceAttr(resourceName, "city_name", "San Francisco"),
					resource.TestCheckResourceAttr(resourceName, "region_name", "CA"),
				),
			},
		},
	})
}

func TestAccNPALocalBroker_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "local_broker_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "local_broker_id"),
				// label_ids is write-only, not returned by API
				ImportStateVerifyIgnore: []string{"label_ids"},
			},
		},
	})
}

func TestAccNPALocalBroker_accessViaPublicIP(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_local_broker.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":        config.StringVariable(rName),
					"access_mode": config.StringVariable("OFF_PREM"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "OFF_PREM"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name":        config.StringVariable(rName),
					"access_mode": config.StringVariable("ON_PREM"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "local_broker_id"),
					resource.TestCheckResourceAttr(resourceName, "access_via_public_ip", "ON_PREM"),
				),
			},
		},
	})
}
