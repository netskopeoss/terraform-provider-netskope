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

func TestAccAIGApplianceDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_aig_appliance.test"
	dataSourceName := "data.netskope_aig_appliance.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_aig_appliance"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "id",
						resourceName, "id",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name",
						resourceName, "name",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "host",
						resourceName, "host",
					),
					resource.TestCheckResourceAttrPair(
						dataSourceName, "status",
						resourceName, "status",
					),
				),
			},
		},
	})
}
