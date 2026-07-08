# Changelog

All notable changes to the Netskope Terraform Provider will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.7] - 2026-07-08

### Fixed
- **`netskope_npa_publisher` ReadResource crash for publishers with connected apps** ([BUG-017](docs/bugs/BUG-017-publisher-connected-apps-type-mismatch.md), [#96](https://github.com/netskopeoss/terraform-provider-netskope/issues/96)) — The `GET /infrastructure/publishers/{id}` endpoint changed `connected_apps` from an array of strings to an array of objects. The SDK struct still declared `[]string`, causing `json: cannot unmarshal object into Go value of type string` on every state refresh for publishers with at least one app connected. Fixed by excluding the field from SDK deserialization (`x-speakeasy-ignore`); it was already excluded from Terraform state. **Affects all releases from v0.3.3 onwards** — upgrade to v0.4.7 to resolve.

## [0.4.6] - 2026-06-29

### Added

**AI Gateway (AIG) resources and data sources** — Full Terraform support for the Netskope AI Gateway:

- **`netskope_aig_appliance`** resource and data source — Manage AIG appliances. Fields: `name`, `host`, `ports` (http/https), `status`, `ai_provider_ids`, `mcp_server_ids`, `sku_addons`, `certificate_imported`.
- **`netskope_aig_appliance_list`** data source — List all AIG appliances.
- **`netskope_aig_appliance_capacity_list`** data source — List capacity metrics per appliance.
- **`netskope_aig_appliance_image_list`** data source — List available AIG firmware images.
- **`netskope_aig_appliance_enrollment_token`** resource — Generate an enrollment token for a registered AIG appliance. Token value and expiry are stored as computed (sensitive) attributes.
- **`netskope_aig_ai_provider`** resource and data source — Manage custom AI providers (e.g. on-prem Ollama, private LLM endpoints). Supports `openai`, `gemini`, and `claude` schemas. Name must start with `cust-`.
- **`netskope_aig_ai_provider_list`** data source — List all AI providers (predefined and custom).
- **`netskope_aig_mcp_server`** resource and data source — Manage custom MCP (Model Context Protocol) servers for AI agent workflows. Supports tool/resource/prompt filtering. Name must start with `mcp-cust-`.
- **`netskope_aig_mcp_server_list`** data source — List all MCP servers (predefined and custom).
- **`netskope_aig_rate_limit`** resource and data source — Manage AI Gateway rate limit rules for AI provider or MCP server traffic. Supports per-token-group and per-model filtering.
- **`netskope_aig_rate_limit_list`** data source — List all rate limit rules.
- **`netskope_aig_token_group`** resource and data source — Manage AIG token groups for grouping API tokens.
- **`netskope_aig_token_group_list`** data source — List all token groups.
- **`netskope_aig_token`** resource and data source — Manage AIG tokens. The token value is write-only (returned only on create) and stored sensitive in state.
- **`netskope_aig_token_list`** data source — List all tokens.

### Fixed
- **Fixed "Cannot configure source IP criteria when flag is disabled" error on `netskope_npa_rules`** — On tenants where the Egress IP feature flag is disabled, the API rejected rule create/update requests that included `b_negateNetLocation` or `b_negateSrcCountries` even when set to `false`. These flags are now stripped from the request when no source IP objects (`net_location_obj`) or source countries (`src_countries`) are configured.
- **Registered `netskope_npa_rules_order` in the provider** — The resource was implemented in v0.4.3 but accidentally omitted from provider registration, causing "Invalid resource type" errors.

## [0.4.5] - 2026-05-29

### Added
- **`netskope_urllist` resource** — Manage URL lists for Netskope policies. Supports create, read, update, delete, and import. Changes are auto-deployed via hooks.
- **`netskope_urllist` data source** — Look up a single URL list by ID.
- **`netskope_urllist_list` data source** — List all URL lists.

### Fixed
- **`netskope_npa_rules_order` now respects provider-level configuration** ([#80](https://github.com/netskopeoss/terraform-provider-netskope/issues/80)) — The resource previously read credentials directly from `NETSKOPE_SERVER_URL` and `NETSKOPE_API_KEY` environment variables, causing it to fail when credentials were supplied via the provider block (required for multi-tenant configurations). It now uses the provider-configured SDK client, consistent with all other resources. The SDK's default retry policy now applies to ordering calls — apply time may be longer if the API returns transient errors. See [Known API Issues #16](docs/KNOWN_API_ISSUES.md#16-npa-rules-ordering--eventual-consistency).

## [0.4.4] - 2026-05-18

### Fixed
- **Fixed `private_app_id = 0` in `netskope_npa_private_apps_list` data source** ([BUG-012](docs/bugs/BUG-012-list-datasource-private-app-id-zero.md)) — The list endpoint returns `app_id` but the SDK mapping read from `id` (absent in list responses). Added `app_id` → `id` normalization in the bulk app AfterSuccess hook.
- **Added `pop_names` and `options` to `netskope_gre_tunnels_list` data source** ([BUG-014](docs/bugs/BUG-014-gre-tunnel-list-missing-pop-names.md)) — Fields were present in the API response but excluded from the Terraform schema by `x-speakeasy-terraform-ignore`. Manually added to generated files and protected via `.genignore`.
- **Added `pop_names` to `netskope_ip_sec_tunnels_list` data source** ([BUG-015](docs/bugs/BUG-015-ipsec-tunnel-list-missing-pop-names.md)) — Same pattern as BUG-014. `pop_names` is Required on the resource but was missing from the list data source.
- **Added `group_name` to `netskope_npa_rules` resource and data sources** ([BUG-016](docs/bugs/BUG-016-npa-rules-group-name-not-mapped.md)) — The API returns `group_name` but it was excluded by `x-speakeasy-terraform-ignore`. Removed the annotation and regenerated. Rules now preserve policy group assignment in state.

## [0.4.3] - 2026-05-14

### Added
- **`netskope_npa_rules_order` resource** — Manages the list position of NPA policy rules.

### Fixed
- **Fixed perpetual hostname whitespace drift on `netskope_npa_private_app`** ([BUG-011](docs/bugs/BUG-011-hostname-whitespace-drift.md)) — Multi-host `private_app_hostname` values with comma-separated hosts caused perpetual plan diffs due to the API normalizing whitespace around commas (e.g., `"host1,host2"` vs `"host1, host2"`). Added `suppressHostnameWhitespaceDrift` plan modifier.

### Changed
- Added documentation for configuring multiple ports per application, including how to split CSV port strings into individual protocol entries to avoid drift. See [terraform-netskope-examples](https://github.com/netskopeoss/terraform-netskope-examples) and the [NPA Rules Guide](docs/guides/NPA_RULES_GUIDE.md#handling-csv-port-strings).

## [0.4.2] - 2026-04-22

### Added
- **Device Classification Tag resource** (`netskope_device_classification_tag`) — Full CRUD with import support. AfterSuccess hooks handle non-standard API responses (create returns `{status, data:[id]}`, update returns `{status:true}`).
- **Device Classification data sources** — `netskope_device_classification_tag` (single lookup), `netskope_device_classification_tag_list` (list all tags), `netskope_device_classification_options_list` (list classification options).
- Rule placement acceptance tests: `TestAccNPARules_ruleOrderBottom`, `TestAccNPARules_ruleOrderBefore`
- Drift detection test: `TestAccDrift_PrivateApp_MultiHostWhitespace` (BUG-011)
- Device classification acceptance tests (5 tests)

### Fixed
- Updated `hc-install` to v0.9.4 to fix expired HashiCorp GPG key
- Updated `terraform-plugin-testing` to v1.14.1 to fix GPG key expiry
- Fixed 10 dependency vulnerabilities: Go 1.26.2, grpc v1.79.3, circl v1.6.3

## [0.4.0] - 2026-03-31

### Added
- **RBAC Labels resource and data sources** ([#63](https://github.com/netskopeoss/terraform-provider-netskope/issues/63)) — New `netskope_rbac_label` resource (full CRUD with import), `netskope_rbac_label` data source, and `netskope_rbac_label_list` data source for Label Based Access Control (LBAC). Supports label hierarchy up to 4 levels via `parent_id`, color assignment, and name-based lookups.
- **`label_ids` on `netskope_npa_private_app`** — RBAC labels can now be assigned to private applications. The AfterSuccess hooks populate `label_ids` from the API's `labels` response to prevent drift.
- **Block rule fields (`emit_alert`, `template`) on `netskope_npa_rules`** — The `match_criteria_action` object now supports `emit_alert` (boolean) and `template` (string) for block actions. Note: the API has a name/filename mismatch on the `template` field requiring a `lifecycle { ignore_changes = [rule_data] }` workaround (see [KNOWN_API_ISSUES #13](docs/KNOWN_API_ISSUES.md#13-api-tokens-cannot-resolve-user-notification-templates-for-block-rules)).
- **Destination Profile resource and data sources** — New `netskope_destination_profile` resource, `netskope_destination_profile` data source, and `netskope_destination_profile_list` data source. Requires tenant licensing.
- **DNS Profile resource and data sources** — New `netskope_dns_profile_v2` resource, `netskope_dns_profile_v2` data source, and `netskope_dns_profile_v2_list` data source with category actions, custom configs, domain lists, and tunnel settings.
- **IPS Status data source** — New `netskope_ips_status` data source. Requires tenant licensing.
- RBAC label sweeper for test cleanup (`sweep_test.go`)
- `testacc-rbaclabels` Makefile target
- 7 RBAC label acceptance tests (basic, update, hierarchy, data sources, publisher integration, private app integration)
- Example configurations for all new resources and data sources

### Fixed
- **Fixed perpetual protocol ordering drift** — Protocols on private apps are now auto-sorted by type then port in AfterSuccess hooks, eliminating perpetual drift for multi-protocol apps. **Upgrade note:** users may see a one-time reorder diff on first plan after upgrading; safe to apply.
- **Fixed `bypass_uris` SQL serialization error on private app update** — Empty `bypass_uris` arrays caused the same SQL error as `paths` ([KNOWN_API_ISSUES #8](docs/KNOWN_API_ISSUES.md#8-empty-objects-cause-sql-serialization-error-on-update)). Added to the BeforeRequest hook's strip list.
- **Fixed `emit_alert`/`template` silently stripped from block rules** — The policy BeforeRequest hook's local `MatchCriteriaAction` struct and the AfterSuccess hook model were missing these fields, causing them to be dropped during unmarshal/re-marshal.
- **Fixed `device_classification_id` type mismatch** ([KNOWN_API_ISSUES #14](docs/KNOWN_API_ISSUES.md)) — The API returns strings but expects integers. OAS uses `string`; BeforeRequest hook coerces to `int` on write.

### Changed
- Updated Go version from 1.24.11 to 1.26.0
- Updated GitHub Actions: `actions/checkout` to v4, `actions/setup-go` to v5
- Removed obsolete `terraform_release.yaml` workflow (superseded by `tf_provider_release.yaml`)
- Updated README with v0.4.0 upgrade guide, block rule `lifecycle` workaround, and RBAC label examples
- Updated KNOWN_API_ISSUES: marked #11 (protocol ordering) as fixed, added #13 (block rule template mismatch), added #14 (device_classification_id type mismatch)

## [0.3.6] - 2026-02-25

### Fixed
- **Fixed concurrent rule creation failure** ([#66](https://github.com/netskopeoss/terraform-provider-netskope/issues/66), [BUG-008](docs/bugs/BUG-008-rule-creation-race-condition.md)) — Creating multiple `netskope_npa_rules` resources concurrently caused intermittent duplicate primary key errors from the API. Added a `sync.Mutex`-based serialization hook (`hookRuleCreateSerializer`) that ensures only one rule creation request is in-flight at a time. Users no longer need `-parallelism=1` or `depends_on` workarounds.
- **Fixed rule creation failure after private app creation** ([#65](https://github.com/netskopeoss/terraform-provider-netskope/issues/65), [BUG-009](docs/bugs/BUG-009-rule-after-app-eventual-consistency.md)) — Creating an NPA rule immediately after the referenced private app intermittently failed with "Private app doesn't exist" due to backend propagation delay. Added an HTTP client wrapper (`hookRuleCreateRetry`) that automatically retries with exponential backoff (up to 60s). Users no longer need `time_sleep` workarounds.
- **Made `protocols` required on `netskope_npa_private_app`** ([BUG-007](docs/bugs/BUG-007-clientless-empty-protocols.md)) — The API requires at least one protocol for all private apps (client-based and clientless), but the schema allowed omitting it. Marked `protocols` as required with `minItems: 1` in the OAS. Terraform now rejects invalid configs at plan time instead of failing with a confusing API error. **Breaking:** users who omit `protocols` will see a plan error (their configs were already broken).
- **Fixed perpetual diff on `private_app_tag_ids` in `netskope_npa_rules`** ([BUG-006](docs/bugs/BUG-006-private-app-tag-ids-drift.md)) — When using `private_app_tags` to reference apps by tag name, the API-computed `private_app_tag_ids` field caused a plan diff on every apply. Fixed by marking `privateAppTagIds` as `x-speakeasy-terraform-ignore` in the OAS and regenerating.
- **Fixed publisher import test failure on `upgrade_status`** — `ImportStateVerify` failed on tenants where POST omits `upgrade_status` but GET returns it. Added `ImportStateVerifyIgnore` for the computed field.
- **Fixed policy groups import test failure on `modify_time`** — POST returns microsecond precision while GET truncates it, causing `ImportStateVerify` mismatch. Added `ImportStateVerifyIgnore` for the computed field.

- **Made `custom_host` computed-only on `netskope_npa_private_app`** ([#62](https://github.com/netskopeoss/terraform-provider-netskope/issues/62)) — `custom_host` is read-only in the v2 API, derived from the CN of a certificate uploaded via the Netskope UI. Removed from POST/PUT request schemas to prevent users setting a value the API silently ignores.

### Added
- Rule creation serialization hook with unit tests (`hookRuleCreateSerializer.go`, `hookRuleCreateSerializer_test.go`)
- Rule creation retry hook with unit tests (`hookRuleCreateRetry.go`, `hookRuleCreateRetry_test.go`)
- Drift detection test coverage for `private_app_tags` in `TestAccDrift_PrivateApp_MultiPublisherWithTags`
- Concurrent rule creation acceptance test (`TestAccNPARules_concurrentCreate`) — verifies BUG-008 fix with 3 independent rules at parallelism 10

## [0.3.5] - 2026-02-12

### Fixed
- **Fixed config-order-dependent plan drift on `netskope_npa_private_app`** (Issues [#56](https://github.com/netskopeoss/terraform-provider-netskope/issues/56)) — Reordering `protocols`, `publishers`, or `tags` list elements in HCL (same elements, different positions) no longer produces a false diff. Added `ModifyPlan` normalization that detects when plan and state contain the same set of elements and suppresses the spurious update. ([BUG-002](docs/bugs/BUG-002-config-order-plan-drift.md))
- **Fixed config-order-dependent plan drift on `netskope_npa_rules`** — Same `ModifyPlan` normalization applied to `private_apps` and `access_method` list attributes in rule data.
- **Fixed config-order-dependent plan drift on `netskope_gre_tunnel`** — `ModifyPlan` normalization for `xff_ip_list` ordering.
- **Fixed config-order-dependent plan drift on `netskope_ip_sec_tunnel`** — `ModifyPlan` normalization for `pop_names` ordering.
- **Fixed plan drift on tunnels with minimal config** — Optional computed attributes (`notes`, `source_type`, `template`, `vendor`) on `netskope_gre_tunnel` and `netskope_ip_sec_tunnel` no longer show "known after apply" when omitted from config.
- **Fixed publisher token exposed in plain text** (Issue [#57](https://github.com/netskopeoss/terraform-provider-netskope/issues/57)) — The `token` attribute on `netskope_npa_publisher_token` is now marked `Sensitive: true`, preventing the token value from appearing in plan/apply output and CI/CD logs.
- **Improved 409 Conflict handling on `netskope_gre_tunnel` create** — Returns a clear "Resource Already Exists" error with guidance to use Terraform import.
- **Fixed `netskope_npa_rules` creation failure with `rule_order.rule_id`** ([BUG-003](docs/bugs/BUG-003-rule-order-type-mismatch.md)) — Creating a rule with `rule_order = { order = "after", rule_id = <id> }` failed because the BeforeRequest hook's `RuleOrder` struct had `rule_id` typed as `*string` instead of `*int64`, and lacked `omitempty` json tags causing nil fields to serialize as `null` which the API rejected.

### Added
- `ModifyPlan` normalization framework for list attribute drift suppression across four resources
- Unit tests for plan modifier logic (`npaprivateapp_resource_planmodify_test.go`, `nparules_resource_planmodify_test.go`)
- Unit tests for BeforeRequest hook rule_order handling (`hookMyPolicyBeforeRequest_test.go`)
- Acceptance test `TestAccDrift_PrivateApp_ReorderedConfig` — reproduces issue #56 scenario (reorder HCL between applies, expect empty plan)
- Acceptance test `TestAccNPARules_ruleOrderAfter` — regression test for BUG-003 (two rules with `rule_order.rule_id` reference)
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
