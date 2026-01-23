// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccResourcePrefix is the prefix used for all test resources to identify
// them as acceptance test resources and enable easy cleanup.
const testAccResourcePrefix = "tf-acc-test"

// testAccProtoV6ProviderFactories creates provider factories for acceptance tests.
// This is used by all acceptance tests to configure the provider.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"netskope": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccPreCheck validates required environment variables are set before
// running acceptance tests. This function should be called in the PreCheck
// function of each acceptance test.
func testAccPreCheck(t *testing.T) {
	t.Helper()

	if v := os.Getenv("NETSKOPE_API_KEY"); v == "" {
		t.Fatal("NETSKOPE_API_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("NETSKOPE_SERVER_URL"); v == "" {
		t.Fatal("NETSKOPE_SERVER_URL must be set for acceptance tests")
	}
}

// testAccProviderConfig returns the provider configuration block for use in
// acceptance test configurations. The provider uses environment variables
// for credentials.
func testAccProviderConfig() string {
	return `
provider "netskope" {
  # Credentials from environment variables:
  # NETSKOPE_API_KEY
  # NETSKOPE_SERVER_URL
}
`
}