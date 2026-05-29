# Netskope Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/netskopeoss/netskope/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/netskopeoss/terraform-provider-netskope)](https://goreportcard.com/report/github.com/netskopeoss/terraform-provider-netskope)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)

The official Terraform provider for [Netskope](https://www.netskope.com/), enabling infrastructure-as-code management of Netskope resources.

> **Examples:** See [terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples) for ready-to-use Terraform configurations covering NPA private apps, policy-as-code, device classification, RBAC labels, and more.

## Upgrading to v0.4.5

### What's New in v0.4.5

- **`netskope_urllist` resource** — Manage URL lists for Netskope policies. Full CRUD with import support. Changes are auto-deployed after each apply.
- **`netskope_urllist` data source** — Look up a single URL list by ID.
- **`netskope_urllist_list` data source** — List all URL lists.
- **Fix `netskope_npa_rules_order` with provider-block credentials** — The resource previously failed when credentials were supplied via the provider block rather than environment variables, breaking multi-tenant configurations. It now uses the provider-configured client consistently with all other resources.

### What's New in v0.4.4

- **List data sources now expose all fields needed for backup/restore workflows:**
  - `netskope_npa_private_apps_list` — `private_app_id` now returns the correct app ID (was always 0)
  - `netskope_gre_tunnels_list` — added `pop_names` and `options` (XFF config)
  - `netskope_ip_sec_tunnels_list` — added `pop_names`
- **`group_name` on NPA rules** — Rules now expose the policy group name as a Computed attribute on both the resource and data sources. Previously only `group_id` (write-only) was available.
- **Block rule template drift fixed** — The `lifecycle { ignore_changes }` workaround is no longer needed. A plan modifier now automatically suppresses the false diff between template display names and file names.

### What's New in v0.4.3

- **`netskope_npa_rules_order` resource** — Manages the list position of NPA policy rules.
- **Fixed hostname whitespace drift** — Multi-host `private_app_hostname` values no longer cause perpetual plan diffs when the API normalizes whitespace around commas.

### What's New in v0.4.2

- **Device Classification Tags** — New `netskope_device_classification_tag` resource (full CRUD with import) and data sources (`netskope_device_classification_tag`, `netskope_device_classification_tag_list`, `netskope_device_classification_options_list`). Use device classification tags in NPA rules via `device_classification_id` to enforce device posture requirements.
- **Dependency security fixes** — Updated Go 1.26.2, hc-install v0.9.4, terraform-plugin-testing v1.14.1, grpc v1.79.3, circl v1.6.3.

### What's New in v0.4.0

- **RBAC Labels** — Full CRUD resource (`netskope_rbac_label`) and data sources for Label Based Access Control. Create labels, look them up by name, and assign them to resources via `label_ids`. See [RBAC Labels](#rbac-labels) for examples.
- **`label_ids` on Private Apps** — You can now assign RBAC labels to private applications directly in Terraform.
- **Block Rule Support** — NPA rules now support `emit_alert` and `template` fields in `match_criteria_action` for block actions.
- **Destination Profiles** — New resource and data sources for managing destination profiles.
- **DNS Profiles** — New resource and data sources for managing DNS security profiles.

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

> **Note:** Block rules must be created through the Netskope UI first, then imported into Terraform. API tokens cannot create block rules directly due to a template resolution limitation. See [KNOWN_API_ISSUES.md](docs/KNOWN_API_ISSUES.md#13-api-tokens-cannot-resolve-user-notification-templates-for-block-rules) for details.

> **Allow rules are not affected** — only block rules that use the `template` field need this workaround.

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
      version = "~> 0.4.5"
    }
  }
}
```

Then run `terraform init`.

## Authentication

### Option 1: Environment Variables (Recommended)

```bash
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export NETSKOPE_API_KEY="your-api-token"
```

```hcl
provider "netskope" {}
```

### Option 2: Provider Configuration

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
| [netskope_destination_profile](docs/resources/destination_profile.md) | Destination profiles |
| [netskope_dns_profile_v2](docs/resources/dns_profile_v2.md) | DNS security profiles |
| [netskope_device_classification_tag](docs/resources/device_classification_tag.md) | Device classification tags |
| [netskope_urllist](docs/resources/urllist.md) | URL lists |

### Data Sources

| Data Source | Description |
|---|---|
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
| [netskope_destination_profile](docs/data-sources/destination_profile.md) | Look up a destination profile |
| [netskope_destination_profile_list](docs/data-sources/destination_profile_list.md) | List destination profiles |
| [netskope_dns_profile_v2](docs/data-sources/dns_profile_v2.md) | Look up a DNS profile |
| [netskope_dns_profile_v2_list](docs/data-sources/dns_profile_v2_list.md) | List DNS profiles |
| [netskope_ips_status](docs/data-sources/ips_status.md) | IPS license status |
| [netskope_device_classification_tag](docs/data-sources/device_classification_tag.md) | Look up a device classification tag |
| [netskope_device_classification_tag_list](docs/data-sources/device_classification_tag_list.md) | List device classification tags |
| [netskope_device_classification_options_list](docs/data-sources/device_classification_options_list.md) | List classification options |
| [netskope_urllist](docs/data-sources/urllist.md) | Look up a URL list by ID |
| [netskope_urllist_list](docs/data-sources/urllist_list.md) | List all URL lists |

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
- **From v0.3.x to v0.4.x**: See [Upgrading to v0.4.5](#upgrading-to-v045) at the top of this document.

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

See [docs/ACCEPTANCE_TESTS.md](docs/ACCEPTANCE_TESTS.md) for full details on running acceptance tests.

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
