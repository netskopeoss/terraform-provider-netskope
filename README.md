# Netskope Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/netskopeoss/netskope/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/netskopeoss/terraform-provider-netskope)](https://goreportcard.com/report/github.com/netskopeoss/terraform-provider-netskope)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)

The official Terraform provider for [Netskope](https://www.netskope.com/), enabling infrastructure-as-code management of Netskope resources.

> **Examples:** See [terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples) for ready-to-use Terraform configurations covering NPA private apps, policy-as-code, device classification, RBAC labels, and more.

## Upgrading to v0.4.13

v0.4.13 is a bug-fix release for `netskope_npa_rules`. It changes how the provider handles the notification `template` field and the `rule_order` field. Read this section before upgrading if you manage **block** or **periodic_reauth** rules.

### What's fixed in v0.4.13

- **Perpetual `template` diff on Terraform Cloud and on tenants where the templates endpoint is unavailable** ([#129](https://github.com/netskopeoss/terraform-provider-netskope/issues/129), [#124](https://github.com/netskopeoss/terraform-provider-netskope/issues/124)) — Earlier releases remembered the display-name ↔ filename mapping in provider memory, which is lost whenever Terraform starts a new provider process (every TFC run, every CLI command). The provider now records the last template display name it sent in Terraform **resource private state**, which lives in the state file and survives across runs. No API call is needed to reconcile the diff.
- **Block-rule updates failing with `template field is required for block action`** ([#124](https://github.com/netskopeoss/terraform-provider-netskope/issues/124)) — Updates always send the display name; the field is never stripped.
- **Every rule update moved the rule to the bottom of the policy** ([#129](https://github.com/netskopeoss/terraform-provider-netskope/issues/129)) — `rule_order` is no longer sent on update. See [`rule_order` is create-time only](#rule_order-is-create-time-only).
- **`periodic_reauth` template changes were not detected** ([#118](https://github.com/netskopeoss/terraform-provider-netskope/issues/118)) — Changing `template` from one display name to another now produces a plan diff and is applied.
- **`periodic_reauth` templates sent as filenames were rejected (`Undefined template`)** — 0.4.11/0.4.12 translated the display name to a filename before sending. The rules API only accepts the display name for `periodic_reauth` (and stores it verbatim), so the translation is removed and the provider no longer calls `/api/v2/templates/usernotifications`. See [Periodic re-authentication rules](#periodic-re-authentication-rules-use-the-template-display-name).

### One-time converging diff after upgrade

On the first `terraform plan` after upgrading, **every block rule** shows a diff of the form:

```
~ rule_data = {
    ~ match_criteria_action = {
        ~ template = "block_page.html" -> "Default Template"
      }
  }
```

This is expected. Apply it once. The apply sends the display name (which the API accepts), records it in private state, and all subsequent plans are empty. The same one-time diff appears after `terraform import` and after creating a new block rule, because the API's create response returns the filename rather than the name you supplied. `periodic_reauth` rules are not affected: the API stores and returns the display name as given, so there is nothing to converge. The one exception is a `periodic_reauth` rule created in the Netskope UI and then imported — the UI stores a filename, so that rule shows the same one-time diff. If you still have `lifecycle { ignore_changes = [rule_data] }` from the v0.4.4–v0.4.11 workarounds, remove it; it is no longer needed and it would hide real changes.

### `rule_order` is create-time only

`rule_order` positions a rule when it is **created**. From v0.4.13 it is not sent on updates, because sending it re-applied the create-time position (`{ order = "bottom" }`) on every change and silently moved block rules below allow rules. Changing `rule_order` in config after creation produces a plan diff that applies without effect on the live position. Use [`netskope_npa_rules_order`](docs/resources/npa_rules_order.md) to manage order after creation.

### Periodic re-authentication rules: use the template display name

`periodic_reauth` rules take a notification `template` like block rules do. Set it to the **display name** shown in the Netskope UI, exactly as for block rules:

```hcl
match_criteria_action = {
  action_name = "periodic_reauth"
  template    = "Default Periodic Authentication Template"   # display name
}
periodic_reauth = {
  reauth_interval      = "8"
  reauth_interval_unit = "hours"
}
```

This works on every tenant and needs no additional API access. Verified against two tenants (one where `/api/v2/templates/usernotifications` returns 403 to all tokens, one where it returns 200): the NPA rules API stores a `periodic_reauth` template display name verbatim on create and update, returns it unchanged on read, and **rejects a `.html` filename** (`Undefined template: 3.html`). So, unlike block rules, there is no display-name/filename diff for `periodic_reauth` rules: state always equals config and no converging apply is needed.

**What changed in 0.4.13.** Releases 0.4.11 and 0.4.12 tried to translate the display name into a filename before sending it, using the templates catalogue endpoint. That endpoint is unavailable on some tenants, and on tenants where it is available the resulting filename is rejected by the rules API — which is why #118 kept recurring. 0.4.13 removes that translation and no longer calls `/api/v2/templates/usernotifications` at all. If an earlier release left you with `template = "10.html"` in configuration, change it back to the display name; the plan will show the change and the apply will succeed.

**Importing a rule created in the Netskope UI.** The UI stores reauth templates by filename (a UI-created rule reads back as `"3.html"`), whereas the API stores names. Importing such a rule and configuring the display name shows a one-time `"3.html" -> "<display name>"` diff; applying it rewrites the rule through the API with the display name. If you would rather not change how the rule is stored, leave it unmanaged.

**Known limitation.** An API-created `periodic_reauth` rule may show no template selected in the Netskope UI's rule editor, because the UI matches templates by filename and the API stored a name. We have confirmed the API accepts and returns the name; whether the backend enforces it identically to a UI-created rule is being confirmed with the Netskope API team. If you observe reauth prompts without the expected template, please report it on the provider repository with the tenant type.

## Release history

Release notes for every version are in [CHANGELOG.md](CHANGELOG.md). Behaviour notes that still
apply to current versions are kept below.

### One-Time Protocol Reorder Diff

In v0.3.x, private apps with multiple protocols could show perpetual drift if protocols were not specified in the exact order the API returned them. v0.4.0 fixes this by automatically sorting protocols in API responses.

**On your first `terraform plan` after upgrading**, you may see a one-time diff that reorders protocols. This is cosmetic — no actual infrastructure change occurs. Run `terraform apply` once to normalize the state. Subsequent plans will be clean.

```
# Example one-time diff (safe to apply):
~ protocols = [
    ~ { port = "443" -> "22", protocol = "tcp" },
    ~ { port = "22" -> "443", protocol = "tcp" },
  ]
```

### Block Rules

NPA block rules can be managed via Terraform using `match_criteria_action` with `action_name = "block"`. Use the template **display name** (e.g. `"Default Template"`) in your config — the provider automatically handles the API's inconsistency of returning file names on read.

```hcl
resource "netskope_npa_rules" "block_example" {
  rule_name = "Block Unauthorized Access"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.default.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "block"
      template    = "Default Template"
      emit_alert  = true
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }
}
```

> **Note:** Block rules can be created directly from Terraform using the template display name (since v0.4.9). The first `terraform plan` after creation or import shows a one-time `"block_page.html" -> "Default Template"` diff; apply it once and subsequent plans are empty. See [One-time converging diff](#one-time-converging-diff-after-upgrade).

> **Allow rules are not affected** — only rules that use the `template` field are subject to the behaviour described here.

### Template Name Resolution

Both `block` and `periodic_reauth` rules use a `template` field in `match_criteria_action`. Always set it to the **display name** shown in the Netskope UI. The provider does not call any template catalogue endpoint; it relies on the rules API, which behaves as follows (verified on two tenants, 2026-10-21):

| Action | Display name on write | Filename on write | GET returns | Plan behaviour |
|---|---|---|---|---|
| `block` | accepted; API resolves it to the filename | rejected (`Undefined template`) | the filename, e.g. `block_page.html` | one converging `"block_page.html" -> "<name>"` update after create or import, then empty plans; reconciled via Terraform private state, identical on CLI and Terraform Cloud |
| `periodic_reauth` | accepted; stored verbatim | rejected (`Undefined template`) | the display name | state equals config; no converging step |

Known limitation (block rules): a template changed out-of-band in the Netskope UI to a *different* template is not detected as drift while the configured display name is unchanged. Changing the display name in config is always detected.

### RBAC Labels

Labels enable Label Based Access Control (LBAC) for managing object-level permissions. You can create labels, build hierarchies, and assign them to publishers, private apps, destination profiles, local brokers, and policy groups.

```hcl
# Create a label
resource "netskope_rbac_label" "engineering" {
  name  = "Engineering"
  color = "#0294C9"
}

# Create a child label (hierarchy up to 4 levels)
resource "netskope_rbac_label" "backend" {
  name      = "Backend"
  parent_id = netskope_rbac_label.engineering.label_id
  color     = "#FF5733"
}

# Look up an existing label by name
data "netskope_rbac_label_list" "all" {}

locals {
  ops_label = one([
    for label in data.netskope_rbac_label_list.all.labels
    : label if label.name == "Operations"
  ])
}

# Assign labels to resources
resource "netskope_npa_private_app" "app" {
  private_app_name     = "My App"
  private_app_hostname = "app.internal.example.com"
  label_ids            = [netskope_rbac_label.backend.label_id]
  # ...
}

resource "netskope_npa_publisher" "pub" {
  publisher_name = "my-publisher"
  label_ids      = [local.ops_label.label_id]
}
```

## Features

### AI Gateway (AIG)
- **Appliances** — Manage AIG appliance deployments, enrollment tokens, capacity, and firmware images
- **AI Providers** — Configure custom AI provider backends (OpenAI, Gemini, or Claude-compatible endpoints)
- **MCP Servers** — Manage custom MCP (Model Context Protocol) servers for AI agent tool/resource access
- **Rate Limits** — Define per-appliance rate limit rules for AI provider and MCP server traffic
- **Token Groups & Tokens** — Manage API token groups and individual tokens for AIG authentication

### Netskope Private Access (NPA)
- **Private Applications** — Create and manage private applications accessible via browser (clientless) or NPA client, with label assignment via `label_ids`
- **Publishers** — Deploy and configure NPA publishers with upgrade profiles, alerting, and bulk operations
- **Local Brokers** — Manage NPA local brokers and their configurations
- **Access Policies** — Define policy groups and rules for zero-trust access control, including block rules with notification templates and rule ordering

### Steering
- **GRE Tunnels** — Manage GRE tunnel configurations, PoPs, and XFF options
- **IPSec Tunnels** — Manage IPSec tunnel configurations, PoPs, encryption, and rekey/reauth options

### Security Profiles
- **DNS Profiles** — Manage DNS security profiles with category actions, custom configs, and tunnel settings
- **Destination Profiles** — Manage destination profiles with label assignment

### Platform
- **Device Classification Tags** — Create and manage device classification tags for device posture enforcement in NPA rules
- **RBAC Labels** — Create and manage labels for Label Based Access Control (LBAC), with hierarchy support up to 4 levels
- **URL Lists** — Manage URL lists used in Netskope steering policies

- **Full Lifecycle Management** — Create, read, update, delete, and import for all supported resources

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Netskope tenant with API v2 access
- REST API v2 token with appropriate permissions

## Installation

```hcl
terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = "~> 0.4.13"
    }
  }
}
```

Then run `terraform init`.

## Authentication

The provider supports two authentication methods.

### Option 1: API key (default)

```bash
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export NETSKOPE_API_KEY="your-api-token"
```

```hcl
provider "netskope" {}
```

Or via the provider block:

```hcl
provider "netskope" {
  server_url = "https://your-tenant.goskope.com/api/v2"
  api_key    = var.netskope_api_key
}
```

| Provider Attribute | Environment Variable | Description |
|---|---|---|
| `server_url` | `NETSKOPE_SERVER_URL` | Netskope tenant API v2 URL |
| `api_key` | `NETSKOPE_API_KEY` | REST API v2 token |

### Option 2: OAuth2 client credentials

Set both OAuth2 environment variables instead of an API key. The provider fetches a bearer token automatically, caches it for its lifetime, and refreshes it on expiry or 401.

```bash
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export NETSKOPE_OAUTH2_CLIENT_ID="<client-id>"
export NETSKOPE_OAUTH2_CLIENT_SECRET="<client-secret>"
```

```hcl
provider "netskope" {}
```

To create an OAuth2 service account in the Netskope admin UI:

1. Enable REST API OAuth2 authentication under **Settings > Administration > Administrators & Roles > Settings > REST API Auth Settings**.
2. On that same page click **Service Account**, then choose **OAuth 2.0** as the token type to obtain a client ID and secret.

The service account must have sufficient API permissions for the resources being managed.

**Precedence:** if `NETSKOPE_API_KEY` is set, it takes priority and OAuth2 is not used.

## Quick Start

### Create a Private Application

```hcl
data "netskope_npa_publishers_list" "all" {}

resource "netskope_npa_private_app" "internal_wiki" {
  private_app_name     = "Internal Wiki"
  private_app_hostname = "wiki.internal.company.com"
  private_app_protocol = "https"
  real_host            = "192.168.10.50"

  clientless_access  = true
  is_user_portal_app = true

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(data.netskope_npa_publishers_list.all.data.publishers[0].publisher_id)
      publisher_name = data.netskope_npa_publishers_list.all.data.publishers[0].publisher_name
    }
  ]

  use_publisher_dns = true
}
```

### Create an Access Policy

```hcl
resource "netskope_npa_rules" "allow_wiki_access" {
  rule_name   = "Allow Wiki Access"
  enabled     = "1"
  description = "Allow authenticated users to access the internal wiki"

  rule_data = {
    policy_type  = "private-app"
    json_version = 3

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = ["[Internal Wiki]"]
    access_method = ["Clientless"]
    user_type     = "user"
  }

  rule_order = {
    order = "top"
  }
}
```

## Resources and Data Sources

### Resources

| Resource | Description |
|---|---|
| [netskope_aig_appliance](docs/resources/aig_appliance.md) | AI Gateway appliances |
| [netskope_aig_appliance_enrollment_token](docs/resources/aig_appliance_enrollment_token.md) | AIG appliance enrollment tokens |
| [netskope_aig_ai_provider](docs/resources/aig_ai_provider.md) | Custom AI providers |
| [netskope_aig_mcp_server](docs/resources/aig_mcp_server.md) | Custom MCP servers |
| [netskope_aig_rate_limit](docs/resources/aig_rate_limit.md) | AI Gateway rate limit rules |
| [netskope_aig_token_group](docs/resources/aig_token_group.md) | AIG token groups |
| [netskope_aig_token](docs/resources/aig_token.md) | AIG tokens |
| [netskope_npa_private_app](docs/resources/npa_private_app.md) | Private applications |
| [netskope_npa_private_app_public_host](docs/resources/npa_private_app_public_host.md) | Private app public host configuration |
| [netskope_npa_publisher](docs/resources/npa_publisher.md) | NPA publishers |
| [netskope_npa_publisher_token](docs/resources/npa_publisher_token.md) | Publisher registration tokens |
| [netskope_npa_publisher_upgrade_profile](docs/resources/npa_publisher_upgrade_profile.md) | Publisher upgrade profiles |
| [netskope_npa_publishers_alerts_configuration](docs/resources/npa_publishers_alerts_configuration.md) | Publisher alert settings |
| [netskope_npa_publishers_bulk_profile_updates](docs/resources/npa_publishers_bulk_profile_updates.md) | Bulk publisher profile updates |
| [netskope_npa_publishers_bulk_upgrade_request](docs/resources/npa_publishers_bulk_upgrade_request.md) | Bulk publisher upgrades |
| [netskope_npa_policy_groups](docs/resources/npa_policy_groups.md) | Policy groups |
| [netskope_npa_rules](docs/resources/npa_rules.md) | Policy rules |
| [netskope_npa_rules_order](docs/resources/npa_rules_order.md) | Policy rule list positioning |
| [netskope_npa_local_broker](docs/resources/npa_local_broker.md) | Local brokers |
| [netskope_npa_local_broker_config](docs/resources/npa_local_broker_config.md) | Local broker configuration |
| [netskope_npa_local_broker_token](docs/resources/npa_local_broker_token.md) | Local broker registration tokens |
| [netskope_gre_tunnel](docs/resources/gre_tunnel.md) | GRE tunnels |
| [netskope_ip_sec_tunnel](docs/resources/ip_sec_tunnel.md) | IPSec tunnels |
| [netskope_rbac_label](docs/resources/rbac_label.md) | RBAC labels for Label Based Access Control |
| [netskope_rbac_role](docs/resources/rbac_role.md) | Custom RBAC admin roles |
| [netskope_custom_category](docs/resources/custom_category.md) | Custom URL categories |
| [netskope_service_object](docs/resources/service_object.md) | Service objects (port/protocol profiles) |
| [netskope_destination_profile](docs/resources/destination_profile.md) | Destination profiles |
| [netskope_dns_profile_v2](docs/resources/dns_profile_v2.md) | DNS security profiles |
| [netskope_device_classification_tag](docs/resources/device_classification_tag.md) | Device classification tags |
| [netskope_device_classification_rule](docs/resources/device_classification_rule.md) | Device classification rules |
| [netskope_device_classification_on_prem_detection](docs/resources/device_classification_on_prem_detection.md) | On-prem detection labels |
| [netskope_device_classification_steering_mapping](docs/resources/device_classification_steering_mapping.md) | Steering config → on-prem label mappings |
| [netskope_device_tag](docs/resources/device_tag.md) | Device tags |
| [netskope_npa_private_app_tag](docs/resources/npa_private_app_tag.md) | NPA private app tags |
| [netskope_urllist](docs/resources/urllist.md) | URL lists |
| [netskope_network_profile](docs/resources/network_profile.md) | Network location profiles (Cloud Firewall) |

### Data Sources

| Data Source | Description |
|---|---|
| [netskope_aig_appliance](docs/data-sources/aig_appliance.md) | Look up an AIG appliance |
| [netskope_aig_appliance_list](docs/data-sources/aig_appliance_list.md) | List AIG appliances |
| [netskope_aig_appliance_capacity_list](docs/data-sources/aig_appliance_capacity_list.md) | List AIG appliance capacity metrics |
| [netskope_aig_appliance_image_list](docs/data-sources/aig_appliance_image_list.md) | List AIG firmware images |
| [netskope_aig_ai_provider](docs/data-sources/aig_ai_provider.md) | Look up an AI provider |
| [netskope_aig_ai_provider_list](docs/data-sources/aig_ai_provider_list.md) | List AI providers |
| [netskope_aig_mcp_server](docs/data-sources/aig_mcp_server.md) | Look up an MCP server |
| [netskope_aig_mcp_server_list](docs/data-sources/aig_mcp_server_list.md) | List MCP servers |
| [netskope_aig_rate_limit](docs/data-sources/aig_rate_limit.md) | Look up a rate limit rule |
| [netskope_aig_rate_limit_list](docs/data-sources/aig_rate_limit_list.md) | List rate limit rules |
| [netskope_aig_token_group](docs/data-sources/aig_token_group.md) | Look up a token group |
| [netskope_aig_token_group_list](docs/data-sources/aig_token_group_list.md) | List token groups |
| [netskope_aig_token](docs/data-sources/aig_token.md) | Look up a token |
| [netskope_aig_token_list](docs/data-sources/aig_token_list.md) | List tokens |
| [netskope_npa_private_app](docs/data-sources/npa_private_app.md) | Look up a private app |
| [netskope_npa_private_apps_list](docs/data-sources/npa_private_apps_list.md) | List private apps |
| [netskope_npa_private_policy_in_use](docs/data-sources/npa_private_policy_in_use.md) | Check policy usage |
| [netskope_npa_publisher](docs/data-sources/npa_publisher.md) | Look up a publisher |
| [netskope_npa_publishers_list](docs/data-sources/npa_publishers_list.md) | List publishers |
| [netskope_npa_publisher_apps_list](docs/data-sources/npa_publisher_apps_list.md) | List apps on a publisher |
| [netskope_npa_publishers_alerts_configuration](docs/data-sources/npa_publishers_alerts_configuration.md) | Publisher alert settings |
| [netskope_npa_publishers_host_os_versions](docs/data-sources/npa_publishers_host_os_versions.md) | Publisher host OS versions |
| [netskope_npa_publishers_releases_list](docs/data-sources/npa_publishers_releases_list.md) | Available publisher releases |
| [netskope_npa_publisher_upgrade_profile](docs/data-sources/npa_publisher_upgrade_profile.md) | Look up an upgrade profile |
| [netskope_npa_publisher_upgrade_profiles_list](docs/data-sources/npa_publisher_upgrade_profiles_list.md) | List upgrade profiles |
| [netskope_npa_policy_groups](docs/data-sources/npa_policy_groups.md) | Look up a policy group |
| [netskope_npa_policy_groups_list](docs/data-sources/npa_policy_groups_list.md) | List policy groups |
| [netskope_npa_rules](docs/data-sources/npa_rules.md) | Look up a policy rule |
| [netskope_npa_rules_list](docs/data-sources/npa_rules_list.md) | List policy rules |
| [netskope_npa_local_broker](docs/data-sources/npa_local_broker.md) | Look up a local broker |
| [netskope_npa_local_broker_config](docs/data-sources/npa_local_broker_config.md) | Local broker configuration |
| [netskope_npa_local_brokers_list](docs/data-sources/npa_local_brokers_list.md) | List local brokers |
| [netskope_gre_tunnel](docs/data-sources/gre_tunnel.md) | Look up a GRE tunnel |
| [netskope_gre_tunnels_list](docs/data-sources/gre_tunnels_list.md) | List GRE tunnels |
| [netskope_grepop](docs/data-sources/grepop.md) | Look up a GRE PoP |
| [netskope_grepo_ps_list](docs/data-sources/grepo_ps_list.md) | List GRE PoPs |
| [netskope_ip_sec_tunnel](docs/data-sources/ip_sec_tunnel.md) | Look up an IPSec tunnel |
| [netskope_ip_sec_tunnels_list](docs/data-sources/ip_sec_tunnels_list.md) | List IPSec tunnels |
| [netskope_ip_sec_pop](docs/data-sources/ip_sec_pop.md) | Look up an IPSec PoP |
| [netskope_ip_sec_po_ps_list](docs/data-sources/ip_sec_po_ps_list.md) | List IPSec PoPs |
| [netskope_rbac_label](docs/data-sources/rbac_label.md) | Look up an RBAC label by ID |
| [netskope_rbac_label_list](docs/data-sources/rbac_label_list.md) | List all RBAC labels |
| [netskope_rbac_role](docs/data-sources/rbac_role.md) | Look up a custom RBAC role |
| [netskope_rbac_role_list](docs/data-sources/rbac_role_list.md) | List all RBAC roles |
| [netskope_custom_category](docs/data-sources/custom_category.md) | Look up a custom URL category |
| [netskope_custom_category_list](docs/data-sources/custom_category_list.md) | List all custom URL categories |
| [netskope_service_object](docs/data-sources/service_object.md) | Look up a service object |
| [netskope_service_object_list](docs/data-sources/service_object_list.md) | List all service objects (custom and predefined) |
| [netskope_destination_profile](docs/data-sources/destination_profile.md) | Look up a destination profile |
| [netskope_destination_profile_list](docs/data-sources/destination_profile_list.md) | List destination profiles |
| [netskope_dns_profile_v2](docs/data-sources/dns_profile_v2.md) | Look up a DNS profile |
| [netskope_dns_profile_v2_list](docs/data-sources/dns_profile_v2_list.md) | List DNS profiles |
| [netskope_ips_status](docs/data-sources/ips_status.md) | IPS license status |
| [netskope_device_classification_tag](docs/data-sources/device_classification_tag.md) | Look up a device classification tag |
| [netskope_device_classification_tag_list](docs/data-sources/device_classification_tag_list.md) | List device classification tags |
| [netskope_device_classification_options_list](docs/data-sources/device_classification_options_list.md) | List classification options |
| [netskope_device_classification_rule](docs/data-sources/device_classification_rule.md) | Look up a device classification rule |
| [netskope_device_classification_rule_list](docs/data-sources/device_classification_rule_list.md) | List device classification rules |
| [netskope_device_classification_on_prem_detection](docs/data-sources/device_classification_on_prem_detection.md) | Look up an on-prem detection label |
| [netskope_device_classification_on_prem_detection_list](docs/data-sources/device_classification_on_prem_detection_list.md) | List on-prem detection labels |
| [netskope_device_classification_steering_mapping](docs/data-sources/device_classification_steering_mapping.md) | Look up a steering config → on-prem label mapping |
| [netskope_device_tag](docs/data-sources/device_tag.md) | Look up a device tag |
| [netskope_device_tag_list](docs/data-sources/device_tag_list.md) | List device tags |
| [netskope_npa_private_app_tag](docs/data-sources/npa_private_app_tag.md) | Look up an NPA private app tag |
| [netskope_npa_private_app_tags_list](docs/data-sources/npa_private_app_tags_list.md) | List NPA private app tags |
| [netskope_urllist](docs/data-sources/urllist.md) | Look up a URL list by ID |
| [netskope_urllist_list](docs/data-sources/urllist_list.md) | List all URL lists |
| [netskope_cci_category_list](docs/data-sources/cci_category_list.md) | List predefined CCI categories (tenant-independent IDs) |
| [netskope_rbac_role_config](docs/data-sources/rbac_role_config.md) | RBAC role configuration catalog (API group IDs, permissions, functions) |
| [netskope_platform_oauth2_token](docs/data-sources/platform_oauth2_token.md) | Exchange OAuth2 client credentials for a bearer token |
| [netskope_network_profile](docs/data-sources/network_profile.md) | Look up a network location profile |
| [netskope_network_profile_list](docs/data-sources/network_profile_list.md) | List all network location profiles |

## Examples and Tutorials

See **[terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples)** for:

- Getting started guides for Terraform beginners
- Step-by-step tutorials for private apps, publishers on AWS/Azure/GCP, policy-as-code
- Working examples for browser apps, client apps, and full deployments
- Best practices for project structure, naming conventions, and CI/CD integration

## Upgrading

See below for version-specific upgrade notes. For full details, see the [Migration Guide](docs/MIGRATION_GUIDE.md).

- **From v0.2.x**: Version 0.3.x is a complete rewrite with renamed resources and changed schemas. Existing state must be re-imported. See the [Migration Guide](docs/MIGRATION_GUIDE.md).
- **From v0.3.2**: See the [v0.3.2 to v0.3.3 upgrade section](docs/MIGRATION_GUIDE.md#upgrading-from-v032-to-v033) for schema changes.
- **From v0.3.x to v0.4.x**: See the [Migration Guide](docs/MIGRATION_GUIDE.md) for schema changes.
- **From v0.4.x to v0.4.9**: See [What's New in v0.4.9](#whats-new-in-v049) at the top of this document.

## Development

### Building from Source

```bash
git clone https://github.com/netskopeoss/terraform-provider-netskope.git
cd terraform-provider-netskope
go build -o terraform-provider-netskope
```

### Testing with a Local Build

1. Build the provider:
   ```bash
   go build -o terraform-provider-netskope
   ```

2. Add a `dev_overrides` block to `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "netskopeoss/netskope" = "/path/to/terraform-provider-netskope"
     }
     direct {}
   }
   ```

3. Run Terraform (no `terraform init` needed with dev overrides):
   ```bash
   terraform plan
   terraform apply
   ```

### Debug Mode

```bash
go run main.go --debug
# Copy the TF_REATTACH_PROVIDERS env var
# In a new terminal:
cd examples/your-example
TF_REATTACH_PROVIDERS=... terraform plan
```

### Running Tests

```bash
# Unit tests
go test ./...

# Acceptance tests (requires NETSKOPE_SERVER_URL and NETSKOPE_API_KEY)
make testacc
```

See [docs/RUNNING_TESTS.md](docs/RUNNING_TESTS.md) for full details on running acceptance tests.

## Contributing

Contributions are welcome. Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

This provider is generated using [Speakeasy](https://www.speakeasy.com/). Files in `internal/sdk/` and `internal/provider/types/` are auto-generated and should not be edited manually. For API-related changes, update the OpenAPI specifications and regenerate.

## License

BSD 3-Clause License - see [LICENSE](LICENSE).

## Support

- [Terraform Registry Documentation](https://registry.terraform.io/providers/netskopeoss/netskope/latest/docs)
- [Issue Tracker](https://github.com/netskopeoss/terraform-provider-netskope/issues)
- [Netskope Documentation](https://docs.netskope.com/)
- [Changelog](CHANGELOG.md)
