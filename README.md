# Netskope Terraform Provider

[![Terraform Registry](https://img.shields.io/badge/terraform-registry-blue.svg)](https://registry.terraform.io/providers/netskopeoss/netskope/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/netskopeoss/terraform-provider-netskope)](https://goreportcard.com/report/github.com/netskopeoss/terraform-provider-netskope)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](LICENSE)

The official Terraform provider for [Netskope](https://www.netskope.com/), enabling infrastructure-as-code management of Netskope Private Access (NPA) resources.

## Features

- **Private Applications**: Create and manage private applications accessible via browser (clientless) or NPA client
- **Publishers**: Deploy and configure NPA publishers with upgrade profiles and alerting
- **Access Policies**: Define policy groups and rules for zero-trust access control
- **Full Lifecycle Management**: Create, read, update, and delete operations for all supported resources

---

### 📚 New to this provider?

Check out **[terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples)** for:

- **Getting Started** - Terraform basics, quick start guide, installation help
- **Tutorials** - Step-by-step guides for private apps, publishers on AWS/Azure/GCP, policy-as-code
- **Working Examples** - Browser apps, client apps, full deployments, cloud infrastructure
- **Best Practices** - Project structure, naming conventions, CI/CD integration

---

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for building from source)
- Netskope tenant with NPA enabled
- API token with appropriate permissions

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = ">= 0.3.3"
    }
  }
}

provider "netskope" {
  server_url = "https://your-tenant.goskope.com/api/v2"
  api_key    = var.netskope_api_key
}
```

Then run `terraform init` to download the provider.

## Authentication

The provider supports two authentication methods:

### Option 1: Environment Variables (Recommended)

Set credentials via environment variables and use an empty provider block:

```bash
export NETSKOPE_SERVER_URL="https://your-tenant.goskope.com/api/v2"
export NETSKOPE_API_KEY="your-api-token"
```

```hcl
provider "netskope" {}
```

### Option 2: Terraform Variables

Define variables and pass credentials explicitly:

```hcl
variable "netskope_server_url" {
  description = "Netskope API server URL"
  type        = string
}

variable "netskope_api_key" {
  description = "Netskope API key"
  type        = string
  sensitive   = true
}

provider "netskope" {
  server_url = var.netskope_server_url
  api_key    = var.netskope_api_key
}
```

Then set via `TF_VAR_` environment variables:

```bash
export TF_VAR_netskope_server_url="https://your-tenant.goskope.com/api/v2"
export TF_VAR_netskope_api_key="your-api-token"
```

Or use a `terraform.tfvars` file (don't commit to version control):

```hcl
netskope_server_url = "https://your-tenant.goskope.com/api/v2"
netskope_api_key    = "your-api-token"
```

## Quick Start

### Create a Browser-Accessible Private Application

```hcl
# Look up existing publishers
data "netskope_npa_publishers_list" "all" {}

# Create a private application
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

## Available Resources and Data Sources

### Resources

| Resource | Description |
|----------|-------------|
| [netskope_npa_private_app](docs/resources/npa_private_app.md) | Manage private applications |
| [netskope_npa_publisher](docs/resources/npa_publisher.md) | Manage NPA publishers |
| [netskope_npa_publisher_token](docs/resources/npa_publisher_token.md) | Generate publisher registration tokens |
| [netskope_npa_publisher_upgrade_profile](docs/resources/npa_publisher_upgrade_profile.md) | Manage publisher upgrade schedules |
| [netskope_npa_rules](docs/resources/npa_rules.md) | Manage NPA access policy rules |
| [netskope_npa_policy_groups](docs/resources/npa_policy_groups.md) | Manage policy groups |
| [netskope_npa_publishers_alerts_configuration](docs/resources/npa_publishers_alerts_configuration.md) | Configure publisher alerts |
| [netskope_npa_publishers_bulk_upgrade_request](docs/resources/npa_publishers_bulk_upgrade_request.md) | Bulk upgrade publishers |
| [netskope_npa_publishers_bulk_profile_updates](docs/resources/npa_publishers_bulk_profile_updates.md) | Bulk update publisher profiles |
| [netskope_npa_private_app_public_host](docs/resources/npa_private_app_public_host.md) | Manage public host configurations |

### Data Sources

| Data Source | Description |
|-------------|-------------|
| [netskope_npa_private_app](docs/data-sources/npa_private_app.md) | Read private application details |
| [netskope_npa_private_apps_list](docs/data-sources/npa_private_apps_list.md) | List all private applications |
| [netskope_npa_publisher](docs/data-sources/npa_publisher.md) | Read publisher details |
| [netskope_npa_publishers_list](docs/data-sources/npa_publishers_list.md) | List all publishers |
| [netskope_npa_publisher_apps_list](docs/data-sources/npa_publisher_apps_list.md) | List apps assigned to a publisher |
| [netskope_npa_publisher_upgrade_profile](docs/data-sources/npa_publisher_upgrade_profile.md) | Read upgrade profile details |
| [netskope_npa_publisher_upgrade_profiles_list](docs/data-sources/npa_publisher_upgrade_profiles_list.md) | List all upgrade profiles |
| [netskope_npa_publishers_releases_list](docs/data-sources/npa_publishers_releases_list.md) | List available publisher releases |
| [netskope_npa_publishers_host_os_versions](docs/data-sources/npa_publishers_host_os_versions.md) | List supported OS versions |
| [netskope_npa_publishers_alerts_configuration](docs/data-sources/npa_publishers_alerts_configuration.md) | Read alerts configuration |
| [netskope_npa_rules](docs/data-sources/npa_rules.md) | Read access rule details |
| [netskope_npa_rules_list](docs/data-sources/npa_rules_list.md) | List all access rules |
| [netskope_npa_policy_groups](docs/data-sources/npa_policy_groups.md) | Read policy group details |
| [netskope_npa_policy_groups_list](docs/data-sources/npa_policy_groups_list.md) | List all policy groups |
| [netskope_npa_private_policy_in_use](docs/data-sources/npa_private_policy_in_use.md) | Check policy usage |

## Examples & Tutorials

> **📚 [terraform-netskope-examples](https://github.com/jharris-ns/terraform-netskope-examples)** - Complete tutorials, guides, and working examples

| Resource | Description |
|----------|-------------|
| [Quick Start Guide](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/getting-started) | Get up and running in minutes |
| [Browser App Example](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/examples/use-cases/browser-app) | Clientless web application access |
| [Client App Example](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/examples/use-cases/client-app) | SSH, RDP, and database access |
| [Publisher Management](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/examples/use-cases/publisher-management) | Publisher lifecycle and upgrades |
| [Policy Rules](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/examples/use-cases/policy-rules) | Access policy configuration |
| [Full Deployment](https://github.com/jharris-ns/terraform-netskope-examples/tree/main/examples/use-cases/full-deployment) | Complete NPA setup |

## Development

### Building from Source

```bash
git clone https://github.com/netskopeoss/terraform-provider-netskope.git
cd terraform-provider-netskope
go build -o terraform-provider-netskope
```

### Testing Locally

1. Build the provider:
   ```bash
   go build -o terraform-provider-netskope
   ```

2. Create a `~/.terraformrc` file with dev overrides:
   ```hcl
   provider_installation {
     dev_overrides {
       "netskopeoss/netskope" = "/path/to/terraform-provider-netskope"
     }
     direct {}
   }
   ```

3. Run Terraform commands (skip `terraform init`):
   ```bash
   terraform plan
   terraform apply
   ```

### Running Tests

```bash
# Unit tests
go test ./...

# Acceptance tests (requires Netskope credentials)
# See the separate terraform-provider-netskope-tests repository
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Note: This provider is generated using [Speakeasy](https://www.speakeasy.com/). Manual changes to generated files in `internal/sdk/` will be overwritten. For API-related changes, please update the OpenAPI specifications.

## License

This project is licensed under the BSD 3-Clause License - see the [LICENSE](LICENSE) file for details.

## Support

- [Documentation](https://docs.netskope.com/)
- [Issue Tracker](https://github.com/netskopeoss/terraform-provider-netskope/issues)
- [Netskope Community](https://community.netskope.com/)

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for release history.

<!-- Start Summary [summary] -->
## Summary

Netskope Terraform Provider: Combined specification to produce netskope terraform provider via speakeasy
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [Netskope Terraform Provider](#netskope-terraform-provider)
  * [Features](#features)
  * [Requirements](#requirements)
  * [Installation](#installation)
  * [Authentication](#authentication)
  * [Quick Start](#quick-start)
  * [Available Resources and Data Sources](#available-resources-and-data-sources)
  * [Examples & Tutorials](#examples-tutorials)
  * [Development](#development)
* [Unit tests](#unit-tests)
* [Acceptance tests (requires Netskope credentials)](#acceptance-tests-requires-netskope-credentials)
* [See the separate terraform-provider-netskope-tests repository](#see-the-separate-terraform-provider-netskope-tests-repository)
* [Copy the TF_REATTACH_PROVIDERS env var](#copy-the-tfreattachproviders-env-var)
* [In a new terminal](#in-a-new-terminal)

<!-- End Table of Contents [toc] -->

<!-- Start Installation [installation] -->
## Installation

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = "0.3.6"
    }
  }
}

provider "netskope" {
  # Configuration options
}
```
<!-- End Installation [installation] -->

<!-- Start Authentication [security] -->
## Authentication

This provider supports authentication configuration via environment variables and provider configuration.

The configuration precedence is:

- Provider configuration
- Environment variables

Available configuration:

| Provider Attribute | Description |
|---|---|
| `api_key` | API Key. Configurable via environment variable `NETSKOPE_API_KEY`. |
<!-- End Authentication [security] -->

<!-- Start Available Resources and Data Sources [operations] -->
## Available Resources and Data Sources

### Resources

* [netskope_npa_policy_groups](docs/resources/npa_policy_groups.md)
* [netskope_npa_private_app](docs/resources/npa_private_app.md)
* [netskope_npa_private_app_public_host](docs/resources/npa_private_app_public_host.md)
* [netskope_npa_publisher](docs/resources/npa_publisher.md)
* [netskope_npa_publishers_alerts_configuration](docs/resources/npa_publishers_alerts_configuration.md)
* [netskope_npa_publishers_bulk_profile_updates](docs/resources/npa_publishers_bulk_profile_updates.md)
* [netskope_npa_publishers_bulk_upgrade_request](docs/resources/npa_publishers_bulk_upgrade_request.md)
* [netskope_npa_publisher_token](docs/resources/npa_publisher_token.md)
* [netskope_npa_publisher_upgrade_profile](docs/resources/npa_publisher_upgrade_profile.md)
* [netskope_npa_rules](docs/resources/npa_rules.md)
### Data Sources

* [netskope_npa_policy_groups](docs/data-sources/npa_policy_groups.md)
* [netskope_npa_policy_groups_list](docs/data-sources/npa_policy_groups_list.md)
* [netskope_npa_private_app](docs/data-sources/npa_private_app.md)
* [netskope_npa_private_apps_list](docs/data-sources/npa_private_apps_list.md)
* [netskope_npa_private_policy_in_use](docs/data-sources/npa_private_policy_in_use.md)
* [netskope_npa_publisher](docs/data-sources/npa_publisher.md)
* [netskope_npa_publisher_apps_list](docs/data-sources/npa_publisher_apps_list.md)
* [netskope_npa_publishers_alerts_configuration](docs/data-sources/npa_publishers_alerts_configuration.md)
* [netskope_npa_publishers_host_os_versions](docs/data-sources/npa_publishers_host_os_versions.md)
* [netskope_npa_publishers_list](docs/data-sources/npa_publishers_list.md)
* [netskope_npa_publishers_releases_list](docs/data-sources/npa_publishers_releases_list.md)
* [netskope_npa_publisher_upgrade_profile](docs/data-sources/npa_publisher_upgrade_profile.md)
* [netskope_npa_publisher_upgrade_profiles_list](docs/data-sources/npa_publisher_upgrade_profiles_list.md)
* [netskope_npa_rules](docs/data-sources/npa_rules.md)
* [netskope_npa_rules_list](docs/data-sources/npa_rules_list.md)
<!-- End Available Resources and Data Sources [operations] -->

<!-- Start Testing the provider locally [usage] -->
## Testing the provider locally

#### Local Provider

Should you want to validate a change locally, the `--debug` flag allows you to execute the provider against a terraform instance locally.

This also allows for debuggers (e.g. delve) to be attached to the provider.

```sh
go run main.go --debug
# Copy the TF_REATTACH_PROVIDERS env var
# In a new terminal
cd examples/your-example
TF_REATTACH_PROVIDERS=... terraform init
TF_REATTACH_PROVIDERS=... terraform apply
```

#### Compiled Provider

Terraform allows you to use local provider builds by setting a `dev_overrides` block in a configuration file called `.terraformrc`. This block overrides all other configured installation methods.

1. Execute `go build` to construct a binary called `terraform-provider-netskope`
2. Ensure that the `.terraformrc` file is configured with a `dev_overrides` section such that your local copy of terraform can see the provider binary

Terraform searches for the `.terraformrc` file in your home directory and applies any configuration settings you set.

```
provider_installation {

  dev_overrides {
      "registry.terraform.io/netskopeoss/netskope" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```
<!-- End Testing the provider locally [usage] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->
