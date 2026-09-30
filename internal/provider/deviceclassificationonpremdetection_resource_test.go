package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccDeviceClassificationOnPremDetection_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_classification_on_prem_detection.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_device_classification_on_prem_detection"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "onprem_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "config"),
					resource.TestCheckResourceAttrSet(resourceName, "onprem_id"),
				),
			},
			{
				ResourceName:                         resourceName,
				ConfigDirectory:                      config.TestNameDirectory(),
				ConfigVariables:                      vars,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIdentifierAttribute: "onprem_id",
				ImportStateIdFunc:                    testutil.ImportStateIdFunc(resourceName, "onprem_id"),
				ImportStateVerifyIgnore: []string{
					// steering_ids is computed and suppressed by plan modifier; populated only
					// when steering mappings are configured via the steering mapping resource.
					"steering_ids",
				},
			},
		},
	})
}

func TestAccDeviceClassificationOnPremDetection_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_classification_on_prem_detection.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_device_classification_on_prem_detection"),
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "onprem_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "onprem_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-updated"),
				),
			},
		},
	})
}

func TestAccDeviceClassificationOnPremDetectionDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_classification_on_prem_detection.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "onprem_id"),
					resource.TestCheckResourceAttr(dsName, "name", rName),
					resource.TestCheckResourceAttrSet(dsName, "config"),
				),
			},
		},
	})
}

func TestAccDeviceClassificationOnPremDetectionListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_classification_on_prem_detection_list.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "onpremdetection.#"),
					resource.TestCheckResourceAttrSet(dsName, "onpremdetection.0.onprem_id"),
					resource.TestCheckResourceAttrSet(dsName, "onpremdetection.0.name"),
					resource.TestCheckResourceAttrSet(dsName, "onpremdetection.0.config"),
				),
			},
		},
	})
}
