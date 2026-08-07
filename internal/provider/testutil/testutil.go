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
	apiKey := os.Getenv("NETSKOPE_API_KEY")
	oauth2ID := os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID")
	oauth2Secret := os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET")
	if apiKey == "" && (oauth2ID == "" || oauth2Secret == "") {
		t.Fatal("Either NETSKOPE_API_KEY or both NETSKOPE_OAUTH2_CLIENT_ID and NETSKOPE_OAUTH2_CLIENT_SECRET must be set for acceptance tests")
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

// SkipUnlessEnvSet skips the test unless the given environment variable is set
// to a non-empty value. Use for opt-in tests that require a specific tenant or
// licensed feature (e.g. TF_RUN_DESTINATION_PROFILES=1 for destination profile tests).
func SkipUnlessEnvSet(t *testing.T, envVar, reason string) {
	t.Helper()
	if os.Getenv(envVar) == "" {
		t.Skipf("Skipping: %s (set %s=1 to run)", reason, envVar)
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

