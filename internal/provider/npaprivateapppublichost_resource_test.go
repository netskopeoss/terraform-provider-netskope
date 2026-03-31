package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPAPrivateAppPublicHost_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_public_host.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"name": config.StringVariable(rName),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
					resource.TestCheckResourceAttrSet(resourceName, "data.public_host"),
				),
			},
		},
	})
}

func TestAccNPAPrivateAppPublicHost_update(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_public_host.test"
	vars := config.Variables{
		"name": config.StringVariable(rName),
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
				),
			},
			{
				ConfigDirectory: config.TestStepDirectory(),
				ConfigVariables: vars,
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
				),
			},
		},
	})
}

func checkNPAPrivateAppPublicHostExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.Attributes["status"] != "success" {
			return fmt.Errorf("expected status 'success', got: %s", rs.Primary.Attributes["status"])
		}

		return nil
	}
}
