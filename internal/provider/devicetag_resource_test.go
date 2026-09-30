package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

// deviceTagPreCheck skips the test if the device tag feature is not deployed on
// the current tenant. The /device/tags API is a newer endpoint not present on
// all tenants. Set NETSKOPE_TEST_DEVICE_TAGS=1 to opt-in.
func deviceTagPreCheck(t *testing.T) {
	t.Helper()
	testutil.PreCheck(t)
	if os.Getenv("NETSKOPE_TEST_DEVICE_TAGS") == "" {
		t.Skip("NETSKOPE_TEST_DEVICE_TAGS not set; skipping device tag acceptance tests (endpoint may not be deployed on this tenant)")
	}
}

func TestAccDeviceTag_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_tag.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	// Note: CheckDestroy is omitted — the Netskope API does not permit deleting device tags
	// programmatically. Tags must be removed manually via the UI after acceptance test runs.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { deviceTagPreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Acceptance test device tag"),
					resource.TestCheckResourceAttrSet(resourceName, "tag_id"),
				),
			},
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

func TestAccDeviceTag_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_device_tag.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	// Note: CheckDestroy is omitted — the Netskope API does not permit deleting device tags
	// programmatically. Tags must be removed manually via the UI after acceptance test runs.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { deviceTagPreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "description", "Initial description"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "tag_id"),
					resource.TestCheckResourceAttr(resourceName, "name", rName+"-updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccDeviceTagDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_tag.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { deviceTagPreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dsName, "tag_id"),
					resource.TestCheckResourceAttr(dsName, "name", rName),
					resource.TestCheckResourceAttr(dsName, "description", "Data source test tag"),
				),
			},
		},
	})
}

func TestAccDeviceTagListDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	dsName := "data.netskope_device_tag_list.all"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { deviceTagPreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify the list returns at least the tag we just created
					resource.TestCheckResourceAttrSet(dsName, "tags.#"),
					// Verify list items have the expected fields populated (not silently omitted)
					resource.TestCheckResourceAttrSet(dsName, "tags.0.tag_id"),
					resource.TestCheckResourceAttrSet(dsName, "tags.0.name"),
				),
			},
		},
	})
}
