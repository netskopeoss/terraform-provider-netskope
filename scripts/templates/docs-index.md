---
page_title: "Netskope Provider"
subcategory: ""
description: |-
  Terraform provider for managing Netskope Private Access (NPA) resources including private applications, publishers, policies, and access rules.
---

# Netskope Provider

The Netskope Terraform provider enables infrastructure-as-code management of Netskope Private Access (NPA) resources. Use this provider to:

- **Manage Private Applications** - Create, update, and delete private apps with full protocol and publisher configuration
- **Configure Publishers** - Set up NPA publishers, generate registration tokens, and manage upgrade profiles
- **Define Access Policies** - Create policy groups and access rules to control user access to private applications

## Quick Start

### 1. Set Environment Variables

```bash
export NETSKOPE_SERVER_URL="https://mytenant.goskope.com/api/v2"
export NETSKOPE_API_KEY="your-api-key"
```

### 2. Configure the Provider

```terraform
terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = ">= 0.3.3"
    }
  }
}

provider "netskope" {}
```

### 3. Create Your First Resource

```terraform
data "netskope_npa_publishers_list" "all" {}

resource "netskope_npa_private_app" "example" {
  private_app_name     = "my-first-app"
  private_app_hostname = "app.internal.company.com"
  private_app_protocol = "https"
  real_host            = "server.internal.company.com"

  clientless_access  = true
  use_publisher_dns  = true

  protocols = [{
    port     = "443"
    protocol = "tcp"
  }]

  publishers = [{
    publisher_id   = tostring(data.netskope_npa_publishers_list.all.data.publishers[0].publisher_id)
    publisher_name = data.netskope_npa_publishers_list.all.data.publishers[0].publisher_name
  }]
}
```

## Guides

- **[Authentication](guides/authentication.md)** - API key setup and credential management
- **[Finding Values](guides/finding-values.md)** - How to discover publisher IDs, policy groups, and other values
- **[Troubleshooting](guides/troubleshooting.md)** - Common errors and solutions

## Tutorials & Examples

For detailed tutorials and complete working examples, see:

**[terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples)**

| Tutorial | Description |
|----------|-------------|
| [Quick Start](https://github.com/jharris-ns/terraform-netskope-examples/blob/main/getting-started/quick-start.md) | Create your first private app |
| [Private App Inventory](https://github.com/jharris-ns/terraform-netskope-examples/blob/main/tutorials/private-app-inventory.md) | Manage apps at scale |
| [Publisher on AWS](https://github.com/jharris-ns/terraform-netskope-examples/blob/main/tutorials/publisher-aws.md) | Deploy publishers in AWS |
| [Policy as Code](https://github.com/jharris-ns/terraform-netskope-examples/blob/main/tutorials/policy-as-code.md) | Manage access rules |

## Resources

| Resource | Description |
|----------|-------------|
| [netskope_npa_private_app](resources/npa_private_app.md) | Manage private applications |
| [netskope_npa_publisher](resources/npa_publisher.md) | Manage NPA publishers |
| [netskope_npa_publisher_token](resources/npa_publisher_token.md) | Generate publisher registration tokens |
| [netskope_npa_publisher_upgrade_profile](resources/npa_publisher_upgrade_profile.md) | Manage publisher upgrade profiles |
| [netskope_npa_policy_groups](resources/npa_policy_groups.md) | Manage policy groups |
| [netskope_npa_rules](resources/npa_rules.md) | Manage NPA access rules |

## Data Sources

| Data Source | Description |
|-------------|-------------|
| [netskope_npa_private_apps_list](data-sources/npa_private_apps_list.md) | List all private applications |
| [netskope_npa_private_app](data-sources/npa_private_app.md) | Get a specific private application |
| [netskope_npa_publishers_list](data-sources/npa_publishers_list.md) | List all publishers |
| [netskope_npa_publisher](data-sources/npa_publisher.md) | Get a specific publisher |
| [netskope_npa_policy_groups_list](data-sources/npa_policy_groups_list.md) | List policy groups |
| [netskope_npa_rules_list](data-sources/npa_rules_list.md) | List NPA rules |

## Authentication

The provider supports two authentication methods:

### Method 1: Environment Variables (Recommended)

```bash
export NETSKOPE_SERVER_URL="https://mytenant.goskope.com/api/v2"
export NETSKOPE_API_KEY="your-api-key"
```

```terraform
provider "netskope" {}
```

### Method 2: Terraform Variables

```bash
export TF_VAR_netskope_server_url="https://mytenant.goskope.com/api/v2"
export TF_VAR_netskope_api_key="your-api-key"
```

```terraform
variable "netskope_server_url" { type = string }
variable "netskope_api_key" { type = string; sensitive = true }

provider "netskope" {
  server_url = var.netskope_server_url
  api_key    = var.netskope_api_key
}
```

## Schema

### Optional

- `server_url` (String) - Netskope tenant API URL. Format: `https://{tenant}.goskope.com/api/v2`. Can also be set via `NETSKOPE_SERVER_URL` environment variable.
- `api_key` (String, Sensitive) - Netskope REST API v2 key. Can also be set via `NETSKOPE_API_KEY` environment variable.

## Creating an API Key

1. Log in to your Netskope admin console
2. Navigate to **Settings** > **Tools** > **REST API v2**
3. Click **New Token**
4. Select the required endpoint permissions:
   - `/api/v2/steering/apps/private` - Private applications
   - `/api/v2/infrastructure/publishers` - Publishers
   - `/api/v2/policy/npa` - Policies and rules
5. Save and copy the API key immediately

## Additional Resources

- [Netskope Documentation](https://docs.netskope.com)
- [GitHub Repository](https://github.com/netskopeoss/terraform-provider-netskope)
- [Report Issues](https://github.com/netskopeoss/terraform-provider-netskope/issues)
