package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNetworkProfile_basic(t *testing.T) {
	testutil.SkipUnlessEnvSet(t, "TF_RUN_NETWORK_PROFILES", "network profiles require a tenant with Cloud Firewall licensing enabled")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_network_profile.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_network_profile"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance test network profile"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttr(resourceName, "values.1", "10.0.0.1-10.0.0.254"),
					resource.TestCheckResourceAttrSet(resourceName, "id"),
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

func TestAccNetworkProfile_update(t *testing.T) {
	testutil.SkipUnlessEnvSet(t, "TF_RUN_NETWORK_PROFILES", "network profiles require a tenant with Cloud Firewall licensing enabled")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_network_profile.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_network_profile"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "values.0", "192.168.1.0/24"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated by acceptance test"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.1", "10.10.0.0/16"),
				),
			},
			{
				// Rename + shrink values (2→1) + drop description from config.
				// Verifies rename does not force replacement, values can shrink,
				// and removing a Computed+Optional field causes no perpetual drift.
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName + "-renamed"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-renamed"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "values.0", "10.10.0.0/16"),
				),
			},
		},
	})
}

func TestAccNetworkProfileDataSource_basic(t *testing.T) {
	testutil.SkipUnlessEnvSet(t, "TF_RUN_NETWORK_PROFILES", "network profiles require a tenant with Cloud Firewall licensing enabled")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_network_profile.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_network_profile"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "values.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "values.0", "192.168.1.0/24"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccNetworkProfileListDataSource_basic(t *testing.T) {
	testutil.SkipUnlessEnvSet(t, "TF_RUN_NETWORK_PROFILES", "network profiles require a tenant with Cloud Firewall licensing enabled")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_network_profile_list.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_network_profile"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "elements.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "elements.0.id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "elements.0.name"),
				),
			},
		},
	})
}
