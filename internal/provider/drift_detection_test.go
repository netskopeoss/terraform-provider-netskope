// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// =============================================================================
// DRIFT DETECTION TESTS
// =============================================================================
//
// These tests verify that resources do not exhibit "perpetual drift" - a
// condition where Terraform detects changes on every plan even when no actual
// changes have been made to the resource configuration.
//
// Common causes of drift:
// - API returns fields in different order than sent (e.g., protocols)
// - API normalizes values differently (e.g., case changes)
// - Computed fields change between API calls
// - Type mismatches between state and API response
//
// Each test:
// 1. Creates a resource
// 2. Runs a second plan to verify no changes are detected
// =============================================================================

// TestAccDrift_LocalBroker verifies no drift on local broker resources
func TestAccDrift_LocalBroker(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPALocalBrokerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftLocalBrokerConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPALocalBrokerExists("netskope_npa_local_broker.test"),
				),
			},
			{
				Config: testAccDriftLocalBrokerConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_LocalBrokerConfig verifies no drift on local broker config
func TestAccDrift_LocalBrokerConfig(t *testing.T) {
	rHostname := fmt.Sprintf("lbr-%s.example.com", acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftLocalBrokerConfigConfig(rHostname),
			},
			{
				Config: testAccDriftLocalBrokerConfigConfig(rHostname),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_Publisher verifies no drift on publisher resources
func TestAccDrift_Publisher(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPAPublisherDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPublisherConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPublisherExists("netskope_npa_publisher.test"),
				),
			},
			{
				Config: testAccDriftPublisherConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_PrivateApp_Basic verifies no drift on basic private app
func TestAccDrift_PrivateApp_Basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPAPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPrivateAppExists("netskope_npa_private_app.test"),
				),
			},
			{
				Config: testAccDriftPrivateAppBasicConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_PrivateApp_MultiProtocol verifies no drift with multiple protocols
// This is a known issue area - protocols must be sorted correctly
func TestAccDrift_PrivateApp_MultiProtocol(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPAPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppMultiProtocolConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPrivateAppExists("netskope_npa_private_app.test"),
				),
			},
			{
				Config: testAccDriftPrivateAppMultiProtocolConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_PolicyGroup verifies no drift on policy group resources
func TestAccDrift_PolicyGroup(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPAPolicyGroupsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPolicyGroupConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPolicyGroupsExists("netskope_npa_policy_groups.test"),
				),
			},
			{
				Config: testAccDriftPolicyGroupConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_UpgradeProfile verifies no drift on upgrade profile resources
func TestAccDrift_UpgradeProfile(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckNPAPublisherUpgradeProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDriftUpgradeProfileConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNPAPublisherUpgradeProfileExists("netskope_npa_publisher_upgrade_profile.test"),
				),
			},
			{
				Config: testAccDriftUpgradeProfileConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// =============================================================================
// Configuration Functions
// =============================================================================

func testAccDriftLocalBrokerConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker" "test" {
  local_broker_name    = %q
  city_name            = "San Francisco"
  region_name          = "CA"
  country_name         = "United States of America"
  country_code         = "US"
  latitude             = 37.7749
  longitude            = -122.4194
  access_via_public_ip = "NONE"
}
`, testAccProviderConfig(), name)
}

func testAccDriftLocalBrokerConfigConfig(hostname string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_local_broker_config" "test" {
  hostname = %q
}
`, testAccProviderConfig(), hostname)
}

func testAccDriftPublisherConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = %q
  lbrokerconnect = false
}
`, testAccProviderConfig(), name)
}

func testAccDriftPrivateAppBasicConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-pub"
  lbrokerconnect = false
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"
  use_publisher_dns    = true

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]
}
`, testAccProviderConfig(), name, name)
}

func testAccDriftPrivateAppMultiProtocolConfig(name string) string {
	// Protocols are ordered: TCP by port ascending, then UDP by port ascending
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-pub"
  lbrokerconnect = false
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"
  use_publisher_dns    = true

  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "3389"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]
}
`, testAccProviderConfig(), name, name)
}

func testAccDriftPolicyGroupConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = %q

  group_order = {
    group_id = "2"
    order    = "after"
  }
}
`, testAccProviderConfig(), name)
}

func testAccDriftUpgradeProfileConfig(name string) string {
	return fmt.Sprintf(`
%s

data "netskope_npa_publishers_releases_list" "releases" {}

resource "netskope_npa_publisher_upgrade_profile" "test" {
  name         = %q
  enabled      = true
  docker_tag   = data.netskope_npa_publishers_releases_list.releases.data[0].docker_tag
  frequency    = "0 0 * * *"
  timezone     = "US/Pacific"
  release_type = "Beta"
}
`, testAccProviderConfig(), name)
}
