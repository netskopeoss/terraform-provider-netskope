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

func TestAccIPSecTunnel_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"
	vars := config.Variables{
		"name":      config.StringVariable(rName),
		"source_ip": config.StringVariable(sourceIP),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttrSet(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "50"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "tunnel_id"),
				// pop_names and psk are not returned by the API in the same format
				ImportStateVerifyIgnore: []string{"pop_names", "psk"},
			},
		},
	})
}

func TestAccIPSecTunnel_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			// Create
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(rName),
					"source_ip": config.StringVariable(sourceIP),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "50"),
				),
			},
			// Update bandwidth and notes
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name":      config.StringVariable(rNameUpdated),
					"source_ip": config.StringVariable(sourceIP),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100"),
					resource.TestCheckResourceAttr(resourceName, "notes", "Updated by acceptance test"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_withEncryption(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"
	vars := config.Variables{
		"name":       config.StringVariable(rName),
		"source_ip":  config.StringVariable(sourceIP),
		"encryption": config.StringVariable("AES256-CBC"),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "encryption", "AES256-CBC"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_withOptions(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"
	vars := config.Variables{
		"name":      config.StringVariable(rName),
		"source_ip": config.StringVariable(sourceIP),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "options.rekey", "true"),
					resource.TestCheckResourceAttr(resourceName, "options.reauth", "true"),
				),
			},
		},
	})
}

func TestAccIPSecTunnel_import(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"
	vars := config.Variables{
		"name":      config.StringVariable(rName),
		"source_ip": config.StringVariable(sourceIP),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
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
				ImportStateVerifyIdentifierAttribute: "tunnel_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "tunnel_id"),
				// pop_names and psk are not returned by the API in the same format
				ImportStateVerifyIgnore: []string{"pop_names", "psk"},
			},
		},
	})
}

func TestAccIPSecTunnel_disabled(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))
	resourceName := "netskope_ip_sec_tunnel.test"
	vars := config.Variables{
		"name":      config.StringVariable(rName),
		"source_ip": config.StringVariable(sourceIP),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tunnel_id"),
					resource.TestCheckResourceAttr(resourceName, "site", rName),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
				),
			},
		},
	})
}
