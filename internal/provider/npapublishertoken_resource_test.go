// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccNPAPublisherToken_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_publisher_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// No CheckDestroy - token resource doesn't support delete
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherTokenConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPublisherTokenExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
					resource.TestCheckResourceAttrSet(resourceName, "token"),
				),
			},
		},
	})
}

// Configuration functions

func testAccNPAPublisherTokenConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
}

resource "netskope_npa_publisher_token" "test" {
  publisher_id = netskope_npa_publisher.test.publisher_id
}
`, testAccProviderConfig(), name)
}

// Helper functions

func testAccCheckNPAPublisherTokenExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if rs.Primary.Attributes["publisher_id"] == "" {
			return fmt.Errorf("publisher_id not set")
		}

		if rs.Primary.Attributes["token"] == "" {
			return fmt.Errorf("token not set")
		}

		return nil
	}
}
