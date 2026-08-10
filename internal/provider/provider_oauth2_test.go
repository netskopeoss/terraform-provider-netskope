package provider_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

// oauthPreCheck is a PreCheck for OAuth2 tests. It requires NETSKOPE_SERVER_URL and
// either NETSKOPE_API_KEY or OAuth2 credentials (after t.Setenv clears the API key,
// the OAuth2 creds satisfy testutil.PreCheck's fallback check).
func oauthPreCheck(t *testing.T) {
	t.Helper()
	testutil.SkipUnlessEnvSet(t, "TF_RUN_OAUTH2_TESTS", "OAuth2 provider auth tests require TF_RUN_OAUTH2_TESTS=1 and valid OAuth2 client credentials registered on the tenant")
	if os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID") == "" || os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET") == "" {
		t.Skip("NETSKOPE_OAUTH2_CLIENT_ID and NETSKOPE_OAUTH2_CLIENT_SECRET must be set")
	}
	testutil.PreCheck(t)
}

// TestAccOAuth2ProviderAuth_PublisherList verifies that the provider can authenticate
// using OAuth2 client credentials alone — with no NETSKOPE_API_KEY set — and
// successfully complete a real API call. This validates the oauth2TokenHook is
// correctly registered and overrides API key authentication end-to-end.
func TestAccOAuth2ProviderAuth_PublisherList(t *testing.T) {
	// Remove API key for the duration of this test so OAuth2 is the only
	// credential available to the provider.
	t.Setenv("NETSKOPE_API_KEY", "")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { oauthPreCheck(t) },
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

// TestAccOAuth2ProviderAuth_PublisherLifecycle verifies that the provider can
// authenticate using OAuth2 client credentials alone for a full resource lifecycle:
// create (POST), read (GET), update (PUT), and destroy (DELETE). This ensures write
// operations work under OAuth2, not just reads.
func TestAccOAuth2ProviderAuth_PublisherLifecycle(t *testing.T) {
	// Remove API key for the duration of this test so OAuth2 is the only
	// credential available to the provider.
	t.Setenv("NETSKOPE_API_KEY", "")

	rName := fmt.Sprintf("%s-%s", testutil.ResourcePrefix, acctest.RandString(8))
	rNameUpdated := fmt.Sprintf("%s-upd", rName)
	resourceName := "netskope_npa_publisher.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { oauthPreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		CheckDestroy:             testutil.CheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			// Create
			{
				Config: fmt.Sprintf(`
provider "netskope" {}

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
}
`, rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "publisher_id"),
				),
			},
			// Update — exercises the PUT path under OAuth2
			{
				Config: fmt.Sprintf(`
provider "netskope" {}

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
}
`, rNameUpdated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testutil.CheckResourceExists(resourceName, "publisher_id"),
					resource.TestCheckResourceAttr(resourceName, "publisher_name", rNameUpdated),
				),
			},
			// Terraform destroys the resource via CheckDestroy (DELETE under OAuth2)
		},
	})
}
