# ns

<div align="left">
    <a href="https://www.speakeasy.com/?utm_source=<no value>&utm_campaign=terraform"><img src="https://custom-icon-badges.demolab.com/badge/-Built%20By%20Speakeasy-212015?style=for-the-badge&logoColor=FBE331&logo=speakeasy&labelColor=545454" /></a>
</div>

<no value>
<!-- Start SDK <no value> -->
To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.3.131"
    }
  }
}

provider "ns" {
  # Configuration options
}
```
<!-- End SDK <no value> -->

<no value>
<!-- Start SDK <no value> -->
### Testing the provider locally

Should you want to validate a change locally, the `--debug` flag allows you to execute the provider against a terraform instance locally.

This also allows for debuggers (e.g. delve) to be attached to the provider.

### Example

```sh
go run main.go --debug
# Copy the TF_REATTACH_PROVIDERS env var
# In a new terminal
cd examples/your-example
TF_REATTACH_PROVIDERS=... terraform init
TF_REATTACH_PROVIDERS=... terraform apply
```
<!-- End SDK <no value> -->

<no value>
<!-- Start SDK <no value> -->

<!-- End SDK <no value> -->

<!-- Start Installation [installation] -->
## Installation

To install this provider, copy and paste this code into your Terraform configuration. Then, run `terraform init`.

```hcl
terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.5.0"
    }
  }
}

provider "ns" {
  # Configuration options
}
```
<!-- End Installation [installation] -->

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

1. Execute `go build` to construct a binary called `terraform-provider-ns`
2. Ensure that the `.terraformrc` file is configured with a `dev_overrides` section such that your local copy of terraform can see the provider binary

Terraform searches for the `.terraformrc` file in your home directory and applies any configuration settings you set.

```
provider_installation {

  dev_overrides {
      "registry.terraform.io/netskope/ns" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```
<!-- End Testing the provider locally [usage] -->

<!-- Start Summary [summary] -->
## Summary

Netskope Terraform Provider: Combined specification to produce netskope terraform provider via speakeasy
<!-- End Summary [summary] -->

<!-- Start Table of Contents [toc] -->
## Table of Contents
<!-- $toc-max-depth=2 -->
* [ns](#ns)
  * [Installation](#installation)
  * [Testing the provider locally](#testing-the-provider-locally)
  * [Available Resources and Data Sources](#available-resources-and-data-sources)

<!-- End Table of Contents [toc] -->

<!-- Start Available Resources and Data Sources [operations] -->
## Available Resources and Data Sources

### Resources

* [ns_npa_policy_groups](docs/resources/npa_policy_groups.md)
* [ns_npa_private_app](docs/resources/npa_private_app.md)
* [ns_npa_private_app_tags](docs/resources/npa_private_app_tags.md)
* [ns_npa_publisher](docs/resources/npa_publisher.md)
* [ns_npa_publishers_alerts_configuration](docs/resources/npa_publishers_alerts_configuration.md)
* [ns_npa_publishers_bulk_profile_updates](docs/resources/npa_publishers_bulk_profile_updates.md)
* [ns_npa_publishers_bulk_upgrade_request](docs/resources/npa_publishers_bulk_upgrade_request.md)
* [ns_npa_publisher_token](docs/resources/npa_publisher_token.md)
* [ns_npa_publisher_upgrade_profile](docs/resources/npa_publisher_upgrade_profile.md)
* [ns_npa_rules](docs/resources/npa_rules.md)
### Data Sources

* [ns_npa_policy_groups](docs/data-sources/npa_policy_groups.md)
* [ns_npa_policy_groups_list](docs/data-sources/npa_policy_groups_list.md)
* [ns_npa_private_app](docs/data-sources/npa_private_app.md)
* [ns_npa_private_apps_list](docs/data-sources/npa_private_apps_list.md)
* [ns_npa_private_apps_tags_list](docs/data-sources/npa_private_apps_tags_list.md)
* [ns_npa_private_app_tags](docs/data-sources/npa_private_app_tags.md)
* [ns_npa_private_policy_in_use](docs/data-sources/npa_private_policy_in_use.md)
* [ns_npa_publisher](docs/data-sources/npa_publisher.md)
* [ns_npa_publisher_apps_list](docs/data-sources/npa_publisher_apps_list.md)
* [ns_npa_publishers_alerts_configuration](docs/data-sources/npa_publishers_alerts_configuration.md)
* [ns_npa_publishers_list](docs/data-sources/npa_publishers_list.md)
* [ns_npa_publishers_releases_list](docs/data-sources/npa_publishers_releases_list.md)
* [ns_npa_publisher_upgrade_profile](docs/data-sources/npa_publisher_upgrade_profile.md)
* [ns_npa_publisher_upgrade_profiles_list](docs/data-sources/npa_publisher_upgrade_profiles_list.md)
* [ns_npa_rules](docs/data-sources/npa_rules.md)
* [ns_npa_rules_list](docs/data-sources/npa_rules_list.md)
<!-- End Available Resources and Data Sources [operations] -->

<!-- Placeholder for Future Speakeasy SDK Sections -->

Terraform allows you to use local provider builds by setting a `dev_overrides` block in a configuration file called `.terraformrc`. This block overrides all other configured installation methods.

Terraform searches for the `.terraformrc` file in your home directory and applies any configuration settings you set.

```
provider_installation {

  dev_overrides {
      "registry.terraform.io/netskope/ns" = "<PATH>"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

Your `<PATH>` may vary depending on how your Go environment variables are configured. Execute `go env GOBIN` to set it, then set the `<PATH>` to the value returned. If nothing is returned, set it to the default location, `$HOME/go/bin`.

Note: To use the dev_overrides, please ensure you run `go build` in this folder. You must have a binary available for terraform to find.

### Contributions

While we value open-source contributions to this SDK, this library is generated programmatically. Any manual changes added to internal files will be overwritten on the next generation. 
We look forward to hearing your feedback. Feel free to open a PR or an issue with a proof of concept and we'll do our best to include it in a future release. 

### SDK Created by [Speakeasy](https://www.speakeasy.com/?utm_source=<no value>&utm_campaign=terraform)
