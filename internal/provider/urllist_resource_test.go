package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccUrllist_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_urllist.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_urllist"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "data.type", "exact"),
					resource.TestCheckResourceAttr(resourceName, "data.urls.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "data.urls.0", "www.example.com"),
					resource.TestCheckResourceAttr(resourceName, "data.urls.1", "www.test.com"),
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

func TestAccUrllist_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_urllist.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_urllist"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "data.type", "exact"),
					resource.TestCheckResourceAttr(resourceName, "data.urls.#", "1"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "data.urls.#", "3"),
				),
			},
		},
	})
}

func TestAccUrllistDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_urllist.test"
	dsName := "data.netskope_urllist.test"
	dsListName := "data.netskope_urllist_list.all"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_urllist"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Resource checks
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					// Data source (single) checks
					resource.TestCheckResourceAttrPair(dsName, "id", resourceName, "id"),
					resource.TestCheckResourceAttr(dsName, "name", rName),
					resource.TestCheckResourceAttr(dsName, "data.type", "exact"),
					resource.TestCheckResourceAttr(dsName, "data.urls.0", "www.datasource-test.com"),
					// List data source checks
					resource.TestCheckResourceAttrSet(dsListName, "items.#"),
				),
			},
		},
	})
}
