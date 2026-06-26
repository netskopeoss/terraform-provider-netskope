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
// publishers, mixed protocols (TCP/UDP), tags, and an NPA rule referencing
// those tags via private_app_tags.
// This is a regression test for BUG-001 (list ordering drift) and BUG-006
// (private_app_tag_ids drift). See docs/bugs/ for details.
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
					testAccCheckResourceExists("netskope_npa_rules.test"),
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

// TestAccDrift_PrivateApp_MultiHostWhitespace verifies no drift when the API
// normalizes whitespace around commas in multi-host private_app_hostname values.
// BUG-011: e.g. config sends "host1,host2", API returns "host1, host2".
func TestAccDrift_PrivateApp_MultiHostWhitespace(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppMultiHostConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
				),
			},
			{
				Config: testAccDriftPrivateAppMultiHostConfig(rName),
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

func testAccDriftPrivateAppMultiHostConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_npa_publisher" "test" {
  publisher_name = "%s-pub"
  lbrokerconnect = false
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = %q
  private_app_hostname = "192.168.1.100,192.168.1.101"
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
// multiple publishers, mixed TCP/UDP protocols, multiple tags, and an NPA rule
// that references those tags via private_app_tags.
// This is a regression test for BUG-001 (perpetual diff on list attributes)
// and BUG-006 (private_app_tag_ids drift).
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

# BUG-006: NPA rule referencing tags via private_app_tags.
# The API resolves tag names to tag IDs; before the fix the computed
# private_app_tag_ids field caused a perpetual diff.
resource "netskope_npa_policy_groups" "test" {
  group_name = "%s-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_rules" "test" {
  rule_name = "%s-rule"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.test.id

  depends_on = [netskope_npa_private_app.test]

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_app_tags = ["%s-tag1"]
    access_method    = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name, name, name, name, name)
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

// =============================================================================
// BUG-002: Config-order-dependent plan drift tests
// =============================================================================
//
// These tests use DELIBERATELY UNSORTED HCL configs — list elements are in a
// different order than the AfterSuccess hooks produce. Before the ModifyPlan
// fix, these would show perpetual "update in-place" diffs. After the fix,
// ModifyPlan detects that plan and state have the same elements (just reordered)
// and suppresses the false diff.
//
// See docs/bugs/BUG-002-config-order-plan-drift.md for details.
// =============================================================================

// TestAccDrift_PrivateApp_UnsortedProtocols verifies no drift when protocols
// are listed in reverse of the hook's sort order (port descending instead of
// ascending). This is the most common user scenario — users list the primary
// port first (443) and secondary ports after.
func TestAccDrift_PrivateApp_UnsortedProtocols(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppUnsortedProtocolsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "protocols.#", "3"),
				),
			},
			{
				Config: testAccDriftPrivateAppUnsortedProtocolsConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_PrivateApp_UnsortedAllLists verifies no drift when ALL list
// attributes (protocols, publishers, tags) are in non-sorted order simultaneously.
// This is the worst case — every list triggers the ModifyPlan normalization.
func TestAccDrift_PrivateApp_UnsortedAllLists(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppUnsortedAllListsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "protocols.#", "3"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "publishers.#", "2"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "tags.#", "2"),
				),
			},
			{
				Config: testAccDriftPrivateAppUnsortedAllListsConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_NPARules_UnsortedLists verifies no drift when NPA rule list
// attributes (private_apps, access_method) are in non-API order.
func TestAccDrift_NPARules_UnsortedLists(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftNPARulesUnsortedListsConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.test"),
				),
			},
			{
				Config: testAccDriftNPARulesUnsortedListsConfig(rName),
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
// BUG-002: Unsorted configuration functions
// =============================================================================

// testAccDriftPrivateAppUnsortedProtocolsConfig creates a private app with
// protocols in REVERSE of hook sort order: port descending (443, 22) and UDP
// before TCP. The hook sorts to [tcp:22, tcp:443, udp:53]. Without ModifyPlan,
// this causes drift on every plan.
func testAccDriftPrivateAppUnsortedProtocolsConfig(name string) string {
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

  # DELIBERATELY UNSORTED: udp before tcp, ports descending.
  # Hook sorts to: [tcp:22, tcp:443, udp:53]
  protocols = [
    {
      port     = "53"
      protocol = "udp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "22"
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

// testAccDriftPrivateAppUnsortedAllListsConfig creates a private app with ALL
// list attributes in non-sorted order: protocols reversed, publishers reversed
// (test2 before test1), and tags in reverse alphabetical order.
func testAccDriftPrivateAppUnsortedAllListsConfig(name string) string {
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

  # DELIBERATELY UNSORTED: udp before tcp, ports descending.
  protocols = [
    {
      port     = "53"
      protocol = "udp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "22"
      protocol = "tcp"
    }
  ]

  # DELIBERATELY UNSORTED: test2 before test1 (higher ID first).
  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test2.publisher_id)
      publisher_name = netskope_npa_publisher.test2.publisher_name
    },
    {
      publisher_id   = tostring(netskope_npa_publisher.test1.publisher_id)
      publisher_name = netskope_npa_publisher.test1.publisher_name
    }
  ]

  # DELIBERATELY UNSORTED: tag2 before tag1 (reverse alphabetical).
  tags = [
    {
      tag_name = "%s-tag2"
    },
    {
      tag_name = "%s-tag1"
    }
  ]
}
`, testAccProviderConfig(), name, name, name, name, name)
}

// testAccDriftNPARulesUnsortedListsConfig creates a rule with multiple
// private_apps and access_method values in non-sorted order.
func testAccDriftNPARulesUnsortedListsConfig(name string) string {
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

resource "netskope_npa_private_app" "test1" {
  private_app_name     = "%s-app1"
  private_app_hostname = "192.168.1.101"

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

resource "netskope_npa_private_app" "test2" {
  private_app_name     = "%s-app2"
  private_app_hostname = "192.168.1.102"

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
  description = "BUG-002 drift test - unsorted lists"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    # DELIBERATELY UNSORTED: app2 before app1 (reverse alphabetical).
    private_apps = [
      netskope_npa_private_app.test2.private_app_name,
      netskope_npa_private_app.test1.private_app_name
    ]

    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name, name)
}

// TestAccDrift_PrivateApp_ReorderedConfig verifies no drift when a user
// reorders their HCL list elements between applies — the exact scenario from
// issue 56. Step 1 creates with "sorted" order, Step 2 uses the same elements
// in reversed order and expects an empty plan.
func TestAccDrift_PrivateApp_ReorderedConfig(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_private_app"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftPrivateAppReorderConfigA(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_private_app.test", "private_app_id"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "protocols.#", "3"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "publishers.#", "2"),
					resource.TestCheckResourceAttr("netskope_npa_private_app.test", "tags.#", "2"),
				),
			},
			{
				Config: testAccDriftPrivateAppReorderConfigB(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_GRETunnel_UnsortedXffIpList verifies no drift when
// xff_ip_list IPs are in reverse order from what the API might return.
func TestAccDrift_GRETunnel_UnsortedXffIpList(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_gre_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftGRETunnelUnsortedXffConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_gre_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftGRETunnelUnsortedXffConfig(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_IPSecTunnel_MinimalConfig verifies no drift when optional
// Computed attributes (encryption, notes, source_type, template, vendor)
// are omitted from config. Without UseStateForUnknown, these would show
// as "known after apply" and cascade to other Computed fields.
func TestAccDrift_IPSecTunnel_MinimalConfig(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("198.51.100.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_ip_sec_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftIPSecTunnelMinimalConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_ip_sec_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftIPSecTunnelMinimalConfig(rName, sourceIP),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_GRETunnel_MinimalConfig verifies no drift when optional
// Computed attributes (notes, source_type, template, vendor) are omitted.
func TestAccDrift_GRETunnel_MinimalConfig(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))
	sourceIP := fmt.Sprintf("203.0.113.%d", acctest.RandIntRange(1, 254))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_gre_tunnel"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftGRETunnelMinimalConfig(rName, sourceIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_gre_tunnel.test", "tunnel_id"),
				),
			},
			{
				Config: testAccDriftGRETunnelMinimalConfig(rName, sourceIP),
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
// BUG-002 Phase 2: Tunnel configuration functions
// =============================================================================

// testAccDriftGRETunnelUnsortedXffConfig creates a GRE tunnel with xff_ip_list
// in reverse order. Tests that ModifyPlan normalizes the list ordering.
func testAccDriftGRETunnelUnsortedXffConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]

  options = {
    xff = {
      xff_enabled = true
      # DELIBERATELY UNSORTED: reverse IP order.
      xff_ip_list = ["10.0.0.3", "10.0.0.2", "10.0.0.1"]
    }
  }
}
`, testAccProviderConfig(), name, sourceIP)
}

// testAccDriftIPSecTunnelMinimalConfig creates an IPSec tunnel with only
// API-required fields — omitting notes, source_type, template, vendor.
// Tests that UseStateForUnknown preserves these from state.
// Note: encryption is required by the API even though the schema marks it Optional.
func testAccDriftIPSecTunnelMinimalConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_ip_sec_tunnel" "test" {
  site            = %q
  source_ip       = %q
  source_identity = "%s.example.com"
  psk             = "TestPreSharedKey123!"
  encryption      = "AES128-CBC"
  pop_names       = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP, name)
}

// =============================================================================
// Issue 56: Config reorder drift test helpers
// =============================================================================

// testAccDriftPrivateAppReorderConfigA creates a private app with list elements
// in "sorted" order: protocols tcp:22, tcp:443, udp:53; publishers test1 then
// test2; tags tag1 then tag2. Used as Step 1 (create) in the reorder test.
func testAccDriftPrivateAppReorderConfigA(name string) string {
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

  # Sorted order: tcp by port ascending, then udp
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

  # Sorted order: test1 then test2
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

  # Sorted order: tag1 then tag2
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

// testAccDriftPrivateAppReorderConfigB creates the SAME private app with
// identical elements but in REVERSED order: protocols udp:53, tcp:443, tcp:22;
// publishers test2 then test1; tags tag2 then tag1. Used as Step 2 (replan)
// in the reorder test to reproduce the issue 56 scenario.
func testAccDriftPrivateAppReorderConfigB(name string) string {
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

  # REVERSED order: udp first, tcp ports descending
  protocols = [
    {
      port     = "53"
      protocol = "udp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "22"
      protocol = "tcp"
    }
  ]

  # REVERSED order: test2 before test1
  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test2.publisher_id)
      publisher_name = netskope_npa_publisher.test2.publisher_name
    },
    {
      publisher_id   = tostring(netskope_npa_publisher.test1.publisher_id)
      publisher_name = netskope_npa_publisher.test1.publisher_name
    }
  ]

  # REVERSED order: tag2 before tag1
  tags = [
    {
      tag_name = "%s-tag2"
    },
    {
      tag_name = "%s-tag1"
    }
  ]
}
`, testAccProviderConfig(), name, name, name, name, name)
}

// testAccDriftGRETunnelMinimalConfig creates a GRE tunnel with only required
// fields — omitting notes, source_type, template, vendor, options.
// Tests that UseStateForUnknown preserves these from state.
func testAccDriftGRETunnelMinimalConfig(name, sourceIP string) string {
	return fmt.Sprintf(`
%s

resource "netskope_gre_tunnel" "test" {
  site      = %q
  source_ip = %q
  pop_names = ["lon1", "lon2"]
}
`, testAccProviderConfig(), name, sourceIP)
}

// =============================================================================
// Destination Profile drift detection tests
// =============================================================================

// TestAccDrift_DestinationProfile verifies no drift on destination profile
// resources with values list and optional description.
func TestAccDrift_DestinationProfile(t *testing.T) {
	t.Skip("Skipping: Destination profiles require a license not enabled on the CI tenant")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_destination_profile"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftDestinationProfileConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_destination_profile.test", "profile_id"),
					resource.TestCheckResourceAttr("netskope_destination_profile.test", "values.#", "3"),
				),
			},
			{
				Config: testAccDriftDestinationProfileConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_DestinationProfile_Minimal verifies no drift when only required
// fields (name, type) are set and optional Computed fields are omitted.
func TestAccDrift_DestinationProfile_Minimal(t *testing.T) {
	t.Skip("Skipping: Destination profiles require a license not enabled on the CI tenant")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_destination_profile"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftDestinationProfileMinimalConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_destination_profile.test", "profile_id"),
				),
			},
			{
				Config: testAccDriftDestinationProfileMinimalConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccDriftDestinationProfileConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_destination_profile" "test" {
  name        = %q
  description = "Drift detection test"
  type        = "insensitive"
  values      = ["example.com", "test.example.com", "api.example.com"]
}
`, testAccProviderConfig(), name)
}

// TestAccDrift_NPARules_BlockTemplate verifies no perpetual drift on block rules
// with template display name. The API returns the file name (e.g. "2.html") but
// the plan modifier suppresses the diff against the config display name.
// See: https://github.com/netskopeoss/terraform-provider-netskope/issues/79
func TestAccDrift_NPARules_BlockTemplate(t *testing.T) {
	t.Skip("API tokens cannot create block rules — KNOWN_API_ISSUES #13")

	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftNPARulesBlockTemplateConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.test"),
				),
			},
			{
				Config: testAccDriftNPARulesBlockTemplateConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccDrift_NPARules_DeviceClassification verifies no perpetual drift on rules
// with device_classification_id. The plan modifier normalizes the list ordering.
func TestAccDrift_NPARules_DeviceClassification(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_npa_rules"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftNPARulesDeviceClassificationConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_npa_rules.test"),
				),
			},
			{
				Config: testAccDriftNPARulesDeviceClassificationConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccDriftNPARulesBlockTemplateConfig(name string) string {
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
  description = "Block template drift test"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "block"
      emit_alert  = true
      template    = "tf-test-template"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
`, testAccProviderConfig(), name, name, name, name)
}

func testAccDriftNPARulesDeviceClassificationConfig(name string) string {
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

resource "netskope_device_classification_tag" "test" {
  name = "%s-tag"
}

resource "netskope_npa_rules" "test" {
  rule_name   = %q
  description = "Device classification drift test"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps             = [netskope_npa_private_app.test.private_app_name]
    access_method            = ["Client"]
    device_classification_id = [tostring(netskope_device_classification_tag.test.tag_id)]
  }
}
`, testAccProviderConfig(), name, name, name, name, name)
}

func testAccDriftDestinationProfileMinimalConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_destination_profile" "test" {
  name = %q
  type = "insensitive"
}
`, testAccProviderConfig(), name)
}

// TestAccDrift_AIGAppliance verifies no drift on AIG appliance resources
func TestAccDrift_AIGAppliance(t *testing.T) {
	rName := fmt.Sprintf("%s-%s", testAccResourcePrefix, acctest.RandString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckResourceDestroy("netskope_aig_appliance"),
		Steps: []resource.TestStep{
			{
				Config: testAccDriftAIGApplianceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckResourceExists("netskope_aig_appliance.test", "id"),
				),
			},
			{
				Config: testAccDriftAIGApplianceConfig(rName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccDriftAIGApplianceConfig(name string) string {
	return fmt.Sprintf(`
%s

resource "netskope_aig_appliance" "test" {
  name = %q
  host = "%s.example.com"

  ports = {
    http = {
      enable = true
      port   = 80
    }
    https = {
      enable = true
      port   = 443
    }
  }
}
`, testAccProviderConfig(), name, name)
}
