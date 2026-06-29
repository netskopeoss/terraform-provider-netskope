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

// AIG MCP server names must start with "mcp-cust-" and be ≤19 chars.
// We use "mcp-cust-tf" (11 chars) + 7 random = 18 chars total.
func aigMcpServerTestName() string {
	return "mcp-cust-tf" + acctest.RandString(7)
}

func TestAccAIGMcpServer_basic(t *testing.T) {
	rName := aigMcpServerTestName()
	resourceName := "netskope_aig_mcp_server.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_mcp_server"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "host", "mcp-backend.example.com"),
					resource.TestCheckResourceAttr(resourceName, "port", "8080"),
					resource.TestCheckResourceAttr(resourceName, "path", "/mcp"),
					resource.TestCheckResourceAttr(resourceName, "protocol", "http"),
					resource.TestCheckResourceAttr(resourceName, "schema", "mcp-http"),
					resource.TestCheckResourceAttr(resourceName, "type", "custom"),
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

func TestAccAIGMcpServer_update(t *testing.T) {
	rName := aigMcpServerTestName()
	resourceName := "netskope_aig_mcp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_mcp_server"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
					"path": config.StringVariable("/mcp"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "path", "/mcp"),
				),
			},
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
					"path": config.StringVariable("/mcp/v2"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "path", "/mcp/v2"),
				),
			},
		},
	})
}

func TestAccAIGMcpServerDataSource_basic(t *testing.T) {
	rName := aigMcpServerTestName()
	resourceName := "netskope_aig_mcp_server.test"
	dataSourceName := "data.netskope_aig_mcp_server.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_mcp_server"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host", resourceName, "host"),
					resource.TestCheckResourceAttrPair(dataSourceName, "path", resourceName, "path"),
					resource.TestCheckResourceAttrPair(dataSourceName, "schema", resourceName, "schema"),
				),
			},
		},
	})
}

func TestAccAIGMcpServerListDataSource_basic(t *testing.T) {
	rName := aigMcpServerTestName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_mcp_server"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netskope_aig_mcp_server_list.test", "elements.#"),
				),
			},
		},
	})
}
