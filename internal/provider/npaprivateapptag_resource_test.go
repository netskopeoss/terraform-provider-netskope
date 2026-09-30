package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPAPrivateAppTag_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_tag.test"
	vars := config.Variables{
		"tag_name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app_tag"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "tag_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "tag_id"),
				),
			},
			// Import by tag_id
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "tag_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "tag_id"),
			},
		},
	})
}

func TestAccNPAPrivateAppTag_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_tag.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app_tag"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"tag_name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "tag_name", rName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"tag_name": config.StringVariable(rName + "-renamed"),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "tag_name", rName+"-renamed"),
				),
			},
		},
	})
}

func TestAccNPAPrivateAppTagDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_npa_private_app_tag.test"
	vars := config.Variables{
		"tag_name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app_tag"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "tag_id"),
					resource.TestCheckResourceAttr(dsName, "tag_name", rName),
				),
			},
		},
	})
}

func TestAccNPAPrivateAppTagsListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_npa_private_app_tags_list.all"
	vars := config.Variables{
		"tag_name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_private_app_tag"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "tags.#"),
					resource.TestCheckResourceAttrSet(dsName, "tags.0.tag_id"),
					resource.TestCheckResourceAttrSet(dsName, "tags.0.tag_name"),
				),
			},
		},
	})
}
