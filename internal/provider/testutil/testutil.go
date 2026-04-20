package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	provider "github.com/netskopeoss/terraform-provider-netskope/internal/provider"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/operations"
	"github.com/netskopeoss/terraform-provider-netskope/internal/sdk/models/shared"
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

// CheckRuleOrder verifies that beforeResource appears before afterResource
// in the NPA rules evaluation order by querying the list API endpoint.
func CheckRuleOrder(beforeResource, afterResource string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rsA, ok := s.RootModule().Resources[beforeResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", beforeResource)
		}
		rsB, ok := s.RootModule().Resources[afterResource]
		if !ok {
			return fmt.Errorf("resource not found: %s", afterResource)
		}

		idA := rsA.Primary.ID
		idB := rsB.Primary.ID

		client := sdk.New(
			sdk.WithServerURL(os.Getenv("NETSKOPE_SERVER_URL")),
			sdk.WithSecurity(shared.Security{
				APIKey: os.Getenv("NETSKOPE_API_KEY"),
			}),
		)

		resp, err := client.ListNPARules(context.Background(), operations.ListNPARulesRequest{})
		if err != nil {
			return fmt.Errorf("failed to list NPA rules: %w", err)
		}
		if resp.NpaPolicyResponse == nil {
			return fmt.Errorf("nil response from ListNPARules")
		}

		indexA, indexB := -1, -1
		for i, rule := range resp.NpaPolicyResponse.Data {
			if rule.ID != nil {
				if *rule.ID == idA {
					indexA = i
				}
				if *rule.ID == idB {
					indexB = i
				}
			}
		}

		if indexA == -1 {
			return fmt.Errorf("rule %s (id=%s) not found in rules list", beforeResource, idA)
		}
		if indexB == -1 {
			return fmt.Errorf("rule %s (id=%s) not found in rules list", afterResource, idB)
		}
		if indexA >= indexB {
			return fmt.Errorf("expected %s (id=%s, position=%d) before %s (id=%s, position=%d)",
				beforeResource, idA, indexA, afterResource, idB, indexB)
		}

		return nil
	}
}
