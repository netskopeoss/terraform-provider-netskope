package provider_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/netskopeoss/terraform-provider-netskope/internal/provider/testutil"
)

func TestAccPlatformOAuth2Token_basic(t *testing.T) {
	clientID := os.Getenv("NETSKOPE_OAUTH2_CLIENT_ID")
	clientSecret := os.Getenv("NETSKOPE_OAUTH2_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		t.Skip("NETSKOPE_OAUTH2_CLIENT_ID and NETSKOPE_OAUTH2_CLIENT_SECRET must be set to run this test")
	}

	dataSourceName := "data.netskope_platform_oauth2_token.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testutil.PreCheck(t) },
		ProtoV6ProviderFactories: testutil.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				ConfigDirectory: config.TestNameDirectory(),
				ConfigVariables: config.Variables{
					"client_id":     config.StringVariable(clientID),
					"client_secret": config.StringVariable(clientSecret),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "access_token"),
					resource.TestCheckResourceAttr(dataSourceName, "token_type", "Bearer"),
					resource.TestCheckResourceAttr(dataSourceName, "grant_type", "client_credentials"),
				),
			},
		},
	})
}
