package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccNPAPublishersAlertsConfiguration_basic(t *testing.T) {
	// Skip: Valid event_types are not documented. API rejects "publisher_up", "publisher_down".
	t.Skip("Skipping: Valid event_types not documented")

	resourceName := "netskope_npa_publishers_alerts_configuration.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				Check: resource.ComposeAggregateTestCheckFunc(
					checkNPAPublishersAlertsConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
					resource.TestCheckResourceAttr(resourceName, "selected_users", "selected"),
					resource.TestCheckResourceAttr(resourceName, "event_types.#", "2"),
				),
			},
		},
	})
}

func checkNPAPublishersAlertsConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}
		return nil
	}
}
