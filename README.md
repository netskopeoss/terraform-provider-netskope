# Netskope Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/netskopeoss/netskope/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/netskopeoss/terraform-provider-netskope)](https://goreportcard.com/report/github.com/netskopeoss/terraform-provider-netskope)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)

The official Terraform provider for [Netskope](https://www.netskope.com/), enabling infrastructure-as-code management of Netskope resources.

## Features

- **Private Applications** - Create and manage private applications accessible via browser (clientless) or NPA client
- **Publishers** - Deploy and configure NPA publishers with upgrade profiles, alerting, and bulk operations
- **Local Brokers** - Manage NPA local brokers and their configurations
- **Access Policies** - Define policy groups and rules for zero-trust access control
- **GRE Tunnels** - Manage GRE tunnel configurations and PoPs
- **IPSec Tunnels** - Manage IPSec tunnel configurations and PoPs
- **Full Lifecycle Management** - Create, read, update, delete, and import for all supported resources

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
      version = "~> 0.3.3"
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
| [netskope_npa_local_broker](docs/resources/npa_local_broker.md) | Local brokers |
| [netskope_npa_local_broker_config](docs/resources/npa_local_broker_config.md) | Local broker configuration |
| [netskope_npa_local_broker_token](docs/resources/npa_local_broker_token.md) | Local broker registration tokens |
| [netskope_gre_tunnel](docs/resources/gre_tunnel.md) | GRE tunnels |
| [netskope_ip_sec_tunnel](docs/resources/ip_sec_tunnel.md) | IPSec tunnels |

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

## Examples and Tutorials

See **[terraform-netskope-examples](https://github.com/netskopeoss/terraform-netskope-examples)** for:

- Getting started guides for Terraform beginners
- Step-by-step tutorials for private apps, publishers on AWS/Azure/GCP, policy-as-code
- Working examples for browser apps, client apps, and full deployments
- Best practices for project structure, naming conventions, and CI/CD integration

## Upgrading

- **From v0.2.x**: See the [Migration Guide](docs/MIGRATION_GUIDE.md) for step-by-step instructions. Version 0.3.x is a complete rewrite with renamed resources and changed schemas. Existing state must be re-imported.
- **From v0.3.2**: See the [v0.3.2 to v0.3.3 upgrade section](docs/MIGRATION_GUIDE.md#upgrading-from-v032-to-v033) for details on schema changes affecting NPA rules, policy groups, and private apps.

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
