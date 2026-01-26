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

func TestAccNPAPrivateAppPublicHost_basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_public_host.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppPublicHostConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
					resource.TestCheckResourceAttrSet(resourceName, "data.public_host"),
				),
			},
		},
	})
}

func TestAccNPAPrivateAppPublicHost_update(t *testing.T) {
	// Note: All attributes require replacement, so an "update" will actually
	// destroy and recreate the resource.
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	resourceName := "netskope_npa_private_app_public_host.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// No CheckDestroy - resource doesn't support delete
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPrivateAppPublicHostConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
				),
			},
			{
				// Change host triggers replacement
				Config: testAccNPAPrivateAppPublicHostConfig_updated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPrivateAppPublicHostExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "status", "success"),
				),
			},
		},
	})
}

// Configuration functions

func testAccNPAPrivateAppPublicHostConfig_basic(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_private_app_public_host" "test" {
  host             = "%s.example.internal"
  real_host        = "192.168.1.100"
  clientless_access = true

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]
}
`, testAccProviderConfig(), name)
}

func testAccNPAPrivateAppPublicHostConfig_updated(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_private_app_public_host" "test" {
  host             = "%s-updated.example.internal"
  real_host        = "192.168.1.200"
  clientless_access = true

  protocols = [
    {
      port     = "8443"
      protocol = "tcp"
    }
  ]
}
`, testAccProviderConfig(), name)
}

// Helper functions

func testAccCheckNPAPrivateAppPublicHostExists(resourceName string) resource.TestCheckFunc {
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
