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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_local_broker"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftLocalBrokerConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_local_broker.test", "local_broker_id"),
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
				Config: testAccNPALocalBrokerConfigConfig_basic(rHostname),
			},
			{
				Config: testAccNPALocalBrokerConfigConfig_basic(rHostname),
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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPublisherConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_publisher.test", "publisher_id"),
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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppMultiProtocolConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
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

// TestAccDrift_PrivateApp_MultiPublisherWithTags verifies no drift with multiple
// publishers, mixed protocols (TCP/UDP), and tags.
// This is a regression test for BUG-001: the API returns publishers and tags in
// non-deterministic order, causing perpetual diffs. The fix sorts these lists
// in AfterSuccess hooks. See docs/bugs/BUG-001-publishers-perpetual-diff.md.
func TestAccDrift_PrivateApp_MultiPublisherWithTags(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppMultiPublisherWithTagsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "publishers.#", "2"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "protocols.#", "3"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "tags.#", "2"),
				),
			},
			{
				Config: testAccDriftPrivateAppMultiPublisherWithTagsConfig(rName),
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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_policy_groups"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_policy_groups.test"),
				),
			},
			{
				Config: testAccNPAPolicyGroupsConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_NPARules_Basic verifies no drift on NPA rules
func TestAccDrift_NPARules_Basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftNPARulesBasicConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.test"),
				),
			},
			{
				Config: testAccDriftNPARulesBasicConfig(rName),
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
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_publisher_upgrade_profile"),
		Steps: []resource.TestStep{
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_publisher_upgrade_profile.test", "publisher_upgrade_profile_id"),
				),
			},
			{
				Config: testAccNPAPublisherUpgradeProfileConfig_basic(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_IPSecTunnel_Basic verifies no drift on basic IPSec tunnel
func TestAccDrift_IPSecTunnel_Basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_basic(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_ip_sec_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccIPSecTunnelConfig_basic(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_IPSecTunnel_WithOptions verifies no drift with IPSec options
func TestAccDrift_IPSecTunnel_WithOptions(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccIPSecTunnelConfig_withOptions(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_ip_sec_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccIPSecTunnelConfig_withOptions(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_IPSecTunnel_AllFields verifies no drift with all configurable fields
func TestAccDrift_IPSecTunnel_AllFields(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftIPSecTunnelAllFieldsConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_ip_sec_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftIPSecTunnelAllFieldsConfig(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_GRETunnel_Basic verifies no drift on basic GRE tunnel
func TestAccDrift_GRETunnel_Basic(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_gre_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftGRETunnelBasicConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_gre_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftGRETunnelBasicConfig(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_GRETunnel_WithOptions verifies no drift with XFF options
func TestAccDrift_GRETunnel_WithOptions(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_gre_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftGRETunnelWithOptionsConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_gre_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftGRETunnelWithOptionsConfig(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_GRETunnel_AllFields verifies no drift with all configurable fields
func TestAccDrift_GRETunnel_AllFields(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_gre_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftGRETunnelAllFieldsConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_gre_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftGRETunnelAllFieldsConfig(rName, sourceIP),
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
//
// The following configs are kept because they have deliberate differences from
// the resource test configs (explicit lbrokerconnect, different protocol combos,
// explicit IP parameters, unique AllFields variants).
//
// Configs reused from resource tests:
//   testAccNPAPolicyGroupsConfig_basic (from npapolicygroups_resource_test.go)
//   testAccNPALocalBrokerConfigConfig_basic (from npalocalbrokerconfig_resource_test.go)
//   testAccNPAPublisherUpgradeProfileConfig_basic (from npapublisherupgradeprofile_resource_test.go)
//   testAccIPSecTunnelConfig_basic (from ipsectunnel_resource_test.go)
//   testAccIPSecTunnelConfig_withOptions (from ipsectunnel_resource_test.go)
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

// testAccDriftPrivateAppMultiPublisherWithTagsConfig creates a private app with
// multiple publishers, mixed TCP/UDP protocols, and multiple tags.
// This is a regression test for BUG-001 (perpetual diff on list attributes).
// The config uses sorted order to match hook output:
// - Publishers: sorted by publisher_id ascending
// - Protocols: sorted by type (tcp before udp), then port ascending
// - Tags: sorted by tag_id ascending (API assigns IDs)
func testAccDriftPrivateAppMultiPublisherWithTagsConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test1" {
  publisher_name = "%s-pub1"
  lbrokerconnect = false
}

resource "netskope_npa_publisher" "test2" {
  publisher_name = "%s-pub2"
  lbrokerconnect = false
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100"
  use_publisher_dns    = true

  # Protocols sorted: tcp by port ascending, then udp by port ascending
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
      port     = "53"
      protocol = "udp"
    }
  ]

  # Publishers will be sorted by publisher_id by the hook.
  # We reference both publishers; order in state will be normalized.
  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test1.publisher_id)
      publisher_name = netskope_npa_publisher.test1.publisher_name
    },
    {
      publisher_id   = tostring(netskope_npa_publisher.test2.publisher_id)
      publisher_name = netskope_npa_publisher.test2.publisher_name
    }
  ]

  # Tags will be sorted by tag_id by the hook.
  # API assigns tag_id on creation; order in state will be normalized.
  tags = [
    {
      tag_name = "%s-tag1"
    },
    {
      tag_name = "%s-tag2"
    }
  ]
}
`, testAccProviderConfig(), name, name, name, name, name)
}

func testAccDriftNPARulesBasicConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_policy_groups" "test" {
  group_name = "%s-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = "%s-app"
  private_app_hostname = "192.168.1.100"

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

  use_publisher_dns       = true
  trust_self_signed_certs = false
}

resource "netskope_npa_rules" "test" {
  rule_name   = %q
  description = "Drift detection test rule"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

func testAccDriftIPSecTunnelAllFieldsConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES256-CBC"
  pop_names       = ["lon1", "lon2"]
  bandwidth       = 100
  enabled         = true
  notes           = "Drift detection test"

  options = {
    rekey  = true
    reauth = true
  }
}
`, testAccProviderConfig(), name, sourceIP, name)
}

func testAccDriftGRETunnelBasicConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP)
}

func testAccDriftGRETunnelWithOptionsConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]

  options = {
    xff = {
      xff_enabled = true
      xff_ip_list = ["10.0.0.1", "10.0.0.2"]
    }
  }
}
`, testAccProviderConfig(), name, sourceIP)
}

func testAccDriftGRETunnelAllFieldsConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site        = %q
  source_ip   = %q
  source_type = "Machine"
  pop_names   = ["lon1", "lon2"]
  bandwidth   = 500
  enabled     = true
  notes       = "Drift detection test"

  options = {
    xff = {
      xff_enabled = true
      xff_ip_list = ["10.0.0.1", "10.0.0.2", "10.0.0.3"]
    }
  }
}
`, testAccProviderConfig(), name, sourceIP)
}
