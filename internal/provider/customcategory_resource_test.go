package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccCustomCategory_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_custom_category.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_custom_category"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance test custom category"),
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

func TestAccCustomCategory_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_custom_category.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_custom_category"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
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
				),
			},
		},
	})
}

func TestAccCustomCategoryDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_custom_category.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_custom_category"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "Acceptance test custom category"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
				),
			},
		},
	})
}

func TestAccCustomCategoryListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_custom_category"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netskope_custom_category_list.test", "elements.#"),
				),
			},
		},
	})
}
