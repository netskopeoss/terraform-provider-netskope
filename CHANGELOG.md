# Changelog

All notable changes to the Netskope Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.5] - 2026-02-12

### Fixed
- **Fixed config-order-dependent plan drift on `netskope_npa_private_app`** (Issues [#56](https://github.com/netskopeoss/terraform-provider-netskope/issues/56)) — Reordering `protocols`, `publishers`, or `tags` list elements in HCL (same elements, different positions) no longer produces a false diff. Added `ModifyPlan` normalization that detects when plan and state contain the same set of elements and suppresses the spurious update. ([BUG-002](docs/bugs/BUG-002-config-order-plan-drift.md))
- **Fixed config-order-dependent plan drift on `netskope_npa_rules`** — Same `ModifyPlan` normalization applied to `private_apps` and `access_method` list attributes in rule data.
- **Fixed config-order-dependent plan drift on `netskope_gre_tunnel`** — `ModifyPlan` normalization for `xff_ip_list` ordering.
- **Fixed config-order-dependent plan drift on `netskope_ip_sec_tunnel`** — `ModifyPlan` normalization for `pop_names` ordering.
- **Fixed plan drift on tunnels with minimal config** — Optional computed attributes (`notes`, `source_type`, `template`, `vendor`) on `netskope_gre_tunnel` and `netskope_ip_sec_tunnel` no longer show "known after apply" when omitted from config.
- **Fixed publisher token exposed in plain text** (Issue [#57](https://github.com/netskopeoss/terraform-provider-netskope/issues/57)) — The `token` attribute on `netskope_npa_publisher_token` is now marked `Sensitive: true`, preventing the token value from appearing in plan/apply output and CI/CD logs.
- **Improved 409 Conflict handling on `netskope_gre_tunnel` create** — Returns a clear "Resource Already Exists" error with guidance to use Terraform import.

### Added
- `ModifyPlan` normalization framework for list attribute drift suppression across four resources
- Unit tests for plan modifier logic (`npaprivateapp_resource_planmodify_test.go`, `nparules_resource_planmodify_test.go`)
- Acceptance test `TestAccDrift_PrivateApp_ReorderedConfig` — reproduces issue #56 scenario (reorder HCL between applies, expect empty plan)
- Acceptance tests for unsorted config drift: `TestAccDrift_PrivateApp_UnsortedProtocols`, `TestAccDrift_PrivateApp_UnsortedAllLists`, `TestAccDrift_NPARules_UnsortedLists`
- Acceptance tests for tunnel drift: `TestAccDrift_GRETunnel_UnsortedXffIpList`, `TestAccDrift_IPSecTunnel_MinimalConfig`, `TestAccDrift_GRETunnel_MinimalConfig`
- Generic plan modifiers package (`internal/planmodifiers/`) with `UseConfigValue` and `UseHoistedValue` for all attribute types




## [0.3.4] - 2026-02-08

### Fixed
- **Fixed perpetual plan drift on `netskope_npa_private_app` publishers attribute** — The API returns `service_publisher_assignments` in non-deterministic order and sometimes includes leading whitespace in `publisher_name`. Since `publishers` is a list (order-sensitive), every plan showed an update. Fixed by sorting publishers by `publisher_id` and trimming whitespace in the AfterSuccess hooks. ([BUG-001](docs/bugs/BUG-001-publishers-perpetual-diff.md))
- **Fixed perpetual plan drift on `netskope_npa_private_app` protocols attribute** — The API returns protocols sorted by type then port internally, but this ordering was undocumented and not enforced by the provider. Users previously had to manually order protocols in their HCL to match the API sort order (KNOWN_API_ISSUES #11). Now sorted automatically in AfterSuccess hooks.
- **Fixed potential plan drift on `netskope_npa_private_app` tags attribute** — Tags are now sorted by `tag_id` in AfterSuccess hooks to ensure deterministic ordering regardless of API return order.

### Added
- Unit tests for list ordering normalization (publishers, protocols, tags), whitespace trimming, and edge cases in SDK hooks (19 tests)
- Acceptance test `TestAccDrift_PrivateApp_MultiPublisherWithTags` to verify BUG-001 fix with multiple publishers, mixed protocols, and tags

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
