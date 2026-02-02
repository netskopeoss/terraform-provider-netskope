# Changelog

All notable changes to the Netskope Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.3] - 2026-01-29

### Fixed
- Fixed environment variable authentication (Issue #38) - Provider now supports native environment variables:
  - `NETSKOPE_SERVER_URL` - Netskope tenant API URL
  - `NETSKOPE_API_KEY` - Netskope REST API v2 key
- Both authentication methods now work:
  - Method 1: Native environment variables with empty provider block (recommended)
  - Method 2: Terraform variables with `TF_VAR_*` environment variables
- Fixed `netskope_npa_rules_list` data source - API response now correctly parsed as object with `data` array
- Fixed `rule_id` type mismatch in NPA rules - changed from integer to string to match API response
- Fixed drift detection issues with empty objects in API responses
- Added hooks for normalizing API response data
- Fixed private app `publishers` and `tags` field handling
- Fixed publisher status enum - added `disconnected` status (Issue #41)
- Fixed protocol field mismatch - schema expected `type` but API returns `transport` (Issue #42)
- Fixed publisher data source path in examples (`data[0]` → `data.publishers[0]`)
- **Fixed perpetual plan drift on `netskope_npa_rules`** - Hidden response-only fields (`modify_by`, `modify_time`, `modify_type`, `policy_type`, `group_name`, `classification`, `periodic_reauth`, `userType`, `version`) from Terraform schema using `x-speakeasy-terraform-ignore`. Removed `group_id` from response schema since the GET API does not return it.
- **Fixed perpetual plan drift on `netskope_ip_sec_tunnel`** - Unified `enable`/`enabled` field names between request and response schemas using `x-speakeasy-name-override`
- Fixed `user_id` field reference in NPA rules tests (field does not exist in schema; was being silently ignored)

### Added
- GRE tunnel resource and data sources (`netskope_gre_tunnel`, `netskope_gre_tunnels_list`, `netskope_grepop`, `netskope_grepo_ps_list`)
- IPSec tunnel resource and data sources (`netskope_ip_sec_tunnel`, `netskope_ip_sec_tunnels_list`, `netskope_ip_sec_pop`, `netskope_ip_sec_po_ps_list`)
- NPA Local Broker resources and data sources (`netskope_npa_local_broker`, `netskope_npa_local_broker_config`, `netskope_npa_local_broker_token`, `netskope_npa_local_brokers_list`)
- Comprehensive drift detection test suite (14 tests across all resource types)
- Acceptance test documentation (`docs/ACCEPTANCE_TESTS.md`)
- Separate examples repository: [terraform-netskope-examples](https://github.com/netskopeoss/terraform-netskope-examples)
  - Getting started guides for Terraform beginners
  - Tutorials for private apps, publishers on AWS/Azure/GCP, policy-as-code
  - Working examples for common use cases
  - Best practices and CI/CD integration guides
- Improved provider documentation with quick start guide
- Post-processing script for Speakeasy regeneration (`scripts/restore-docs.sh`)

### Changed
- Updated OpenAPI specification for NPA policy endpoints to match actual API responses
- Moved comprehensive examples to external repository (terraform-netskope-examples)
- Updated README with prominent link to examples repository
- Improved docs/index.md with quick start, guides, and tutorial links

### Removed
- Removed `examples/use-cases/` from provider repo (moved to terraform-netskope-examples)

## [0.3.2] - 2026-01-15

### Added
- Initial release of the new Netskope Terraform Provider codebase
- Netskope Private Access (NPA) support:
  - Private Applications management
  - Publisher management with registration tokens
  - Upgrade profile management
  - Publisher alerts configuration
  - Available version management
- Policy Groups management
- Policy Rules management

### Key Features
- Logical attribute names (e.g., `private_app_id` instead of just `id`)
- Removed duplicate/unused attributes from resources and tfstate
- Normalized API input/output handling:
  - Input: `private_app_name: "my app"`
  - Response: `private_app_name: "my app"` (brackets handled internally)

### Deprecated
- Version 0.2.x is no longer supported

## Migration Notice

The 0.3.x version represents a complete rewrite of the Netskope Terraform Provider. Users migrating from version 0.2.x should note:

1. **Breaking Changes**: Resource schemas have changed significantly
2. **State Migration**: Existing state files may need to be recreated
3. **New Features**: Many new resources and data sources are available

We recommend:
1. Export your current configuration
2. Destroy existing resources managed by the old provider
3. Update to the new provider
4. Re-import or recreate resources

## [0.2.6] - Deprecated

This version is no longer supported. Please upgrade to 0.3.x.
