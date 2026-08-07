package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

// TestAccOAuth2ProviderAuth_PublisherList verifies that the provider can authenticate
// using OAuth2 client credentials alone — with no NETSKOPE_API_KEY set — and
// successfully complete a real API call. This validates the oauth2TokenHook is
// correctly registered and overrides API key authentication end-to-end.
func TestAccOAuth2ProviderAuth_PublisherList(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("acceptance tests skipped unless TF_ACC set")
	}
	testutil.SkipUnlessEnvSet(t, "TF_RUN_OAUTH2_TESTS", "OAuth2 provider auth tests require TF_RUN_OAUTH2_TESTS=1 and valid OAuth2 client credentials registered on the tenant")
	if os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID") == "" || os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET") == "" {
		t.Skip("NETSKOPE_OAUTH2_CLIENT_ID and NETSKOPE_OAUTH2_CLIENT_SECRET must be set")
	}

	// Remove API key for the duration of this test so OAuth2 is the only
	// credential available to the provider.
	t.Setenv("NETSKOPE_API_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// No api_key in the provider block — relies entirely on OAuth2 env vars.
				Config: `
provider "netskope" {}

data "netskope_npa_publishers_list" "test" {}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.netskope_npa_publishers_list.test", "data.publishers.#"),
				),
			},
		},
	})
}
