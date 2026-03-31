package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccRBACLabel_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_rbac_label.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_rbac_label"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "label_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "color", "#0294C9"),
					resource.TestCheckResourceAttrSet(resourceName, "label_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "label_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "label_id"),
			},
		},
	})
}

func TestAccRBACLabel_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-updated", rName)
	resourceName := "netskope_rbac_label.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_rbac_label"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "label_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "color", "#0294C9"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rNameUpdated),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "label_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "color", "#FF5733"),
				),
			},
		},
	})
}

func TestAccRBACLabel_hierarchy(t *testing.T) {
	parentName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	childName := fmt.Sprintf("%s-child-%s", testutil.ResourcePrefix, acctest.RandString(8))
	parentResourceName := "netskope_rbac_label.parent"
	childResourceName := "netskope_rbac_label.child"
	vars := config.Variables{
		"parent_name": config.StringVariable(parentName),
		"child_name":  config.StringVariable(childName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_rbac_label"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(parentResourceName, "label_id"),
					testutil.CheckResourceExists(childResourceName, "label_id"),
					resource.TestCheckResourceAttr(parentResourceName, "name", parentName),
					resource.TestCheckResourceAttr(childResourceName, "name", childName),
					resource.TestCheckResourceAttrPair(childResourceName, "parent_id", parentResourceName, "label_id"),
				),
			},
		},
	})
}

func TestAccRBACLabelDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_rbac_label.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_rbac_label"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "color", "#0294C9"),
					resource.TestCheckResourceAttrSet(dataSourceName, "label_id"),
				),
			},
		},
	})
}

func TestAccRBACLabel_withPublisher(t *testing.T) {
	labelName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	pubName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	labelResourceName := "netskope_rbac_label.test"
	pubResourceName := "netskope_npa_publisher.test"
	vars := config.Variables{
		"label_name":     config.StringVariable(labelName),
		"publisher_name": config.StringVariable(pubName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(labelResourceName, "label_id"),
					testutil.CheckResourceExists(pubResourceName, "publisher_id"),
					resource.TestCheckResourceAttr(labelResourceName, "name", labelName),
					resource.TestCheckResourceAttr(pubResourceName, "publisher_name", pubName),
					resource.TestCheckResourceAttr(pubResourceName, "label_ids.#", "1"),
					resource.TestCheckResourceAttrPair(pubResourceName, "label_ids.0", labelResourceName, "label_id"),
				),
			},
		},
	})
}

func TestAccRBACLabel_withPrivateApp(t *testing.T) {
	labelName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	appName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	labelResourceName := "netskope_rbac_label.test"
	appResourceName := "netskope_npa_private_app.test"
	vars := config.Variables{
		"label_name": config.StringVariable(labelName),
		"app_name":   config.StringVariable(appName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(labelResourceName, "label_id"),
					testutil.CheckResourceExists(appResourceName, "private_app_id"),
					resource.TestCheckResourceAttr(appResourceName, "label_ids.#", "1"),
					resource.TestCheckResourceAttrPair(appResourceName, "label_ids.0", labelResourceName, "label_id"),
				),
			},
		},
	})
}

func TestAccRBACLabelListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dataSourceName := "data.netskope_rbac_label_list.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_rbac_label"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "total_count"),
				),
			},
		},
	})
}
