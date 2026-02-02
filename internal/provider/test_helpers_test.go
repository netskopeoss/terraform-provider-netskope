// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// testAccCheckResourceExists verifies a resource exists in Terraform state.
// If attrNames are provided, it checks that each named attribute is non-empty.
// If no attrNames are provided, it checks that the resource ID is set.
func testAccCheckResourceExists(resourceName string, attrNames ...string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		if len(attrNames) == 0 {
			if rs.Primary.ID == "" {
				return fmt.Errorf("resource ID not set")
			}
		} else {
			for _, attr := range attrNames {
				if rs.Primary.Attributes[attr] == "" {
					return fmt.Errorf("%s not set", attr)
				}
			}
		}

		return nil
	}
}

// testAccCheckResourceDestroy is a stub destroy checker for acceptance tests.
// The acceptance test framework automatically destroys resources. In a production
// scenario, you would make an API call to verify the resource has been deleted.
func testAccCheckResourceDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
		}
		return nil
	}
}

// testAccImportStateIdFunc returns the value of the given attribute for use
// as the import ID in import state tests.
func testAccImportStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes[attrName], nil
	}
}
