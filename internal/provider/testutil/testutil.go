package testutil

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	provider "github.com/netskopeoss/terraform-provider-netskope/internal/provider"
)

const ResourcePrefix = "tf-acc-test"

var ProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"netskope": providerserver.NewProtocol6WithError(provider.New("test")()),
}

func PreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("NETSKOPE_API_KEY"); v == "" {
		t.Fatal("NETSKOPE_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("NETSKOPE_SERVER_URL"); v == "" {
		t.Fatal("NETSKOPE_SERVER_URL must be set for acceptance tests")
	}
}

func CheckResourceExists(resourceName string, attrNames ...string) resource.TestCheckFunc {
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

func CheckResourceDestroy(resourceType string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != resourceType {
				continue
			}
		}
		return nil
	}
}

func ImportStateIdFunc(resourceName, attrName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("resource not found: %s", resourceName)
		}
		return rs.Primary.Attributes[attrName], nil
	}
}

