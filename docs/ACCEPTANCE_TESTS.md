# Acceptance Test Reference

This document tracks test parameters, coverage, and dependencies across all
acceptance tests. Update this file when adding or modifying tests.

Last updated: 2026-04-22

---

## Running Tests

```bash
# All acceptance tests
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m

# Specific resource (e.g. GRE tunnels)
TF_ACC=1 go test -v ./internal/provider/ -run "TestAccGRETunnel_" -timeout 30m

# Drift detection tests only
TF_ACC=1 go test -v ./internal/provider/ -run "TestAccDrift_" -timeout 30m

# Specific drift test
TF_ACC=1 go test -v ./internal/provider/ -run "TestAccDrift_IPSecTunnel_Basic" -timeout 30m

# Device classification tests
TF_ACC=1 go test -v ./internal/provider/ -run "TestAccDeviceClassification" -timeout 30m

# Rule placement tests
TF_ACC=1 go test -v ./internal/provider/ -run "TestAccNPARules_ruleOrder" -timeout 30m
```

Required environment variables: `NETSKOPE_SERVER_URL`, `NETSKOPE_API_KEY`

---

## Latest Results (0.4.2 - 2026-04-20)

| Metric | Count |
|--------|-------|
| **Total Tests** | 118 |
| **Passed** | 112 |
| **Failed** | 0 |
| **Skipped** | 6 |

### Test Breakdown

| Category | Count |
|----------|-------|
| Acceptance tests (resource CRUD) | 56 |
| Acceptance tests (data sources) | 20 |
| Acceptance tests (drift detection) | 22 |
| Unit tests (plan modifier helpers) | 22 |

### Changes Since 0.3.6

#### Added in v0.4.0

| Test | File | Description |
|------|------|-------------|
| `TestAccRBACLabel_basic` | `rbaclabel_resource_test.go` | Create, verify, import |
| `TestAccRBACLabel_update` | `rbaclabel_resource_test.go` | Update name, color |
| `TestAccRBACLabel_hierarchy` | `rbaclabel_resource_test.go` | Parent/child label hierarchy |
| `TestAccRBACLabelDataSource_basic` | `rbaclabel_resource_test.go` | Single label data source |
| `TestAccRBACLabel_withPublisher` | `rbaclabel_resource_test.go` | Label assignment to publisher |
| `TestAccRBACLabel_withPrivateApp` | `rbaclabel_resource_test.go` | Label assignment to private app |
| `TestAccRBACLabelListDataSource_basic` | `rbaclabel_resource_test.go` | List all labels |
| `TestAccDNSProfileV2_basic` | `dnsprofilev2_resource_test.go` | Basic DNS profile CRUD |
| `TestAccDNSProfileV2_update` | `dnsprofilev2_resource_test.go` | Update DNS profile |
| `TestAccDNSProfileV2_withSecurityCategories` | `dnsprofilev2_resource_test.go` | Security category actions |
| `TestAccDNSProfileV2_withTunnelConfig` | `dnsprofilev2_resource_test.go` | Tunnel detection config |
| `TestAccDNSProfileV2_withCustomConfig` | `dnsprofilev2_resource_test.go` | Custom DNS server |
| `TestAccDNSProfileV2_withDomainLists` | `dnsprofilev2_resource_test.go` | Domain allow/block lists |
| `TestAccDNSProfileV2_fullConfig` | `dnsprofilev2_resource_test.go` | All options |
| `TestAccDNSProfileV2_blockAllExceptAllowList` | `dnsprofilev2_resource_test.go` | Block-all with allowlist |
| `TestAccDNSProfileV2_withTunnelAllowList` | `dnsprofilev2_resource_test.go` | Tunnel allow list |
| `TestAccDNSProfileV2DataSource_basic` | `dnsprofilev2_resource_test.go` | DNS profile data source |
| `TestAccDNSProfileV2ListDataSource_basic` | `dnsprofilev2_resource_test.go` | DNS profile list data source |
| `TestAccDNSProfileV2_import` | `dnsprofilev2_resource_test.go` | Import state |

#### Added in v0.4.2

| Test | File | Description |
|------|------|-------------|
| `TestAccDeviceClassificationTagListDataSource_basic` | `deviceclassificationtag_data_source_test.go` | List all device classification tags |
| `TestAccDeviceClassificationOptionsListDataSource_basic` | `deviceclassificationtag_data_source_test.go` | List classification options |
| `TestAccDeviceClassificationTag_basic` | `deviceclassificationtag_resource_test.go` | Tag CRUD with import |
| `TestAccDeviceClassificationTag_update` | `deviceclassificationtag_resource_test.go` | Tag name/description update |
| `TestAccNPARules_ruleOrderBottom` | `nparules_resource_test.go` | Rule placement order=bottom verification |
| `TestAccNPARules_ruleOrderBefore` | `nparules_resource_test.go` | Rule placement order=before verification |

### Skips

| Test | Reason |
|------|--------|
| `TestAccNPAPolicyGroups_update` | Provider bug: missing `new_order` parameter |
| `TestAccNPAPublishersAlertsConfiguration_basic` | Valid `event_types` not documented by API |
| `TestAccNPARules_denyRule` | API template name/filename mismatch (KNOWN_API_ISSUES #13) |
| `TestAccDestinationProfile_basic` | Pending tenant license enablement |
| `TestAccDestinationProfile_update` | Pending tenant license enablement |
| `TestAccIPSStatusDataSource_basic` | Pending tenant license enablement |

---

## Test Parameter Registry

Hardcoded values used across test configurations. When an API value changes
(e.g. a POP is decommissioned or an encryption algorithm is deprecated),
search this table to find every test that needs updating.

### Environment & Naming

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Resource prefix | `tf-acc-test` | All tests | Defined as `testAccResourcePrefix` |
| Random suffix | `acctest.RandString(8)` | All tests | Appended to prefix for uniqueness |
| Provider config | `testAccProviderConfig()` | All tests | Reads `NETSKOPE_SERVER_URL`, `NETSKOPE_API_KEY` |

### IP Addresses

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| GRE source IP range | `203.0.113.{1-254}` | GRE tunnel tests, GRE drift tests | RFC 5737 TEST-NET-3 (documentation range) |
| IPSec source IP range | `198.51.100.{1-254}` | IPSec tunnel tests, IPSec drift tests | RFC 5737 TEST-NET-2 (documentation range) |
| Private app hostname | `192.168.1.100` | Private app, rules, data source tests | RFC 1918 private range |
| Private app hostname (updated) | `192.168.1.100,192.168.1.101` | Private app update test | Comma-separated multi-host |
| Public host real_host | `192.168.1.100` | Public host basic test | |
| Public host real_host (updated) | `192.168.1.200` | Public host update test | |

### POP Configuration

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| GRE/IPSec POP names | `["lon1", "lon2"]` | All GRE and IPSec tunnel tests | London POPs; must exist in tenant |
| GRE POP ID (hex) | `0x00AD` | `grepop_data_source_test.go` | Hex ID for lon1 London POP |

### IPSec-Specific

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| PSK | `TestPreSharedKey123!` | All IPSec tunnel tests | Same value everywhere |
| Encryption (default) | `AES128-CBC` | Most IPSec tests | |
| Encryption (alt) | `AES256-CBC` | `withEncryption` test, drift AllFields | |
| Source identity format | `{name}.example.com` | All IPSec tests | FQDN format |
| Default bandwidth | `50` | IPSec resource (schema default) | Mbps |
| Updated bandwidth | `100` | IPSec update test, drift AllFields | |

### GRE-Specific

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Default bandwidth | `1000` | GRE resource (schema default) | Mbps |
| Updated bandwidth | `500` | GRE update test, drift AllFields | |
| XFF IP list | `["10.0.0.1", "10.0.0.2"]` | GRE XFF test, drift WithOptions | |
| XFF IP list (extended) | `["10.0.0.1", "10.0.0.2", "10.0.0.3"]` | GRE drift AllFields | |
| Source type tested | `"Machine"` | GRE withSourceType, drift AllFields | One of: User, Machine, IoT, Guest Wifi, Mixed |

### NPA Private App

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Protocols (basic) | TCP/443 | Most private app tests | |
| Protocols (multi) | TCP/22, TCP/443, UDP/53 | `multipleProtocols` test | Must be in ascending port order per type |
| Protocols (clientless) | TCP/80 | `clientlessAccess` test | HTTP for browser access |
| Protocols (public host) | TCP/443 | Public host basic | |
| Protocols (public host updated) | TCP/8443 | Public host update | |
| Real host (clientless) | `browser.internal.test` | `clientlessAccess` test | |
| Tag name | `tf-acc-test` | `tags` test | |
| Publisher suffix | `-publisher` | Most tests | Appended to test name |

### NPA Policy & Rules

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Default group ID | `"2"` | Policy group tests, rules tests | References the "Default" policy group |
| Group order | `"after"` | All policy group configs | Placed after Default group |
| Policy type | `"private-app"` | All rules tests | |
| Action (allow) | `"allow"` | Basic/update rules tests | |
| Action (block) | `"block"` | Deny rule test (SKIPPED) | Requires DLP/Threat Protection profile |
| Users | (not set) | All rules tests | Defaults to empty list; API does not accept `*` wildcard |
| Access method | `["Client"]` | All rules tests | |
| Enabled flag | `"1"` / `"0"` | Rules tests | String, not boolean |
| Concurrent rules | 3 independent rules | `concurrentCreate` test | Exercises BUG-008 mutex serializer |

### NPA Publisher & Upgrade Profile

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Upgrade profile ID | `"1"` | Publisher `withUpgradeProfile` test | Hardcoded; assumes profile exists |
| Cron frequency | `"0 0 * * *"` | Upgrade profile tests, drift test | Daily at midnight |
| Release type | `"Beta"` | Upgrade profile tests, drift test | |
| Timezone | `"US/Pacific"` | Upgrade profile tests, drift test | |

### NPA Local Broker

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| City (drift test) | `"San Francisco"` | Drift detection test | |
| City (data source) | `"Cupertino"` | Local broker data source test | |
| Region | `"CA"` | Local broker tests | |
| Country | `"United States of America"` | Local broker tests | |
| Country code | `"US"` | Local broker tests | |
| Latitude | `37.7749` | Drift detection test | San Francisco coordinates |
| Longitude | `-122.4194` | Drift detection test | San Francisco coordinates |
| Access via public IP | `"NONE"` | Local broker tests | |
| Hostname format | `lbr-{rand}.example.com` | Local broker config tests | |

### NPA Alerts (SKIPPED)

| Parameter | Value | Used In | Notes |
|-----------|-------|---------|-------|
| Admin email | `jharris@netskope.com` | Alerts config test (SKIPPED) | Environment-specific |
| Event types | `["publisher_up", "publisher_down"]` | Alerts config test (SKIPPED) | API rejects these values |

---

## Coverage Matrix

Fields tested per resource across test types. Columns:
- **C** = Create (basic test)
- **U** = Update test
- **I** = Import verified
- **D** = Drift detection test
- **O** = Other dedicated test

### netskope_gre_tunnel

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| site | x | x | x | x | | |
| source_ip | x | x | x | x | | |
| pop_names | x | x | | x | | Not verified on import |
| enabled | x | | | x | x | Dedicated disabled test |
| bandwidth | x | x | | x | | Default 1000, updated to 500 |
| notes | | x | | x | | |
| source_type | | | | x | x | Only "Machine" tested |
| template | | | | | | **NOT TESTED** |
| vendor | | | | | | **NOT TESTED** |
| options.xff.xff_enabled | | | | x | x | |
| options.xff.xff_ip_list | | | | x | x | |

### netskope_ip_sec_tunnel

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| site | x | x | x | x | | |
| source_ip | x | x | x | x | | |
| source_identity | x | x | x | x | | FQDN format |
| psk | x | x | | x | | Not verified on import (write-only) |
| encryption | x | x | x | x | x | AES128-CBC and AES256-CBC tested |
| pop_names | x | x | | x | | Not verified on import |
| enabled | x | | | x | x | Dedicated disabled test |
| bandwidth | x | x | | x | | Default 50, updated to 100 |
| notes | | x | | x | | |
| source_type | | | | | | **NOT TESTED** |
| template | | | | | | **NOT TESTED** |
| vendor | | | | | | **NOT TESTED** |
| options.rekey | | | | x | x | |
| options.reauth | | | | x | x | |
| options.xff.enabled | | | | | | **NOT TESTED** |
| options.xff.iplist | | | | | | **NOT TESTED** |

### netskope_npa_private_app

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| private_app_name | x | x | x | x | | |
| private_app_hostname | x | x | x | x | | |
| protocols | x | x | | x | x | Not verified on import |
| publishers | x | | | x | x | Not verified on import; updatePublishers test |
| use_publisher_dns | x | x | | x | | |
| trust_self_signed_certs | x | x | | | | |
| clientless_access | | | | | x | Dedicated test |
| is_user_portal_app | | x | | | | Set in complete config |
| allow_unauthenticated_cors | | x | | | x | |
| real_host | | | | | x | Clientless access test; not verified on import |
| private_app_protocol | | | | | x | Clientless access test |
| tags | | | | | x | Dedicated test |
| custom_host | | | | x | | Computed-only; derived from cert CN (0.3.6) |
| allow_uri_bypass | | | | x | | Drift-protected; requires URI Bypass flag |
| bypass_uris | | | | x | | Drift-protected; requires URI Bypass flag |
| uribypass_header_value | | | | x | | Drift-protected; requires URI Bypass flag |
| app_option | | | | x | | Drift-protected; clientless_access only |
| hide_app_in_portal | | | | x | | Drift-protected; requires User Portal flag |
| upgrade_insecure_requests | | | | x | | Drift-protected; requires Enterprise Browser flag |
| paths | | | | x | | Drift-protected; requires User Portal flag |

### netskope_npa_publisher

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| publisher_name | x | x | x | x | | |
| lbrokerconnect | | | | | | Checked but not explicitly set |
| publisher_upgrade_profiles_id | | | | | x | Hardcoded ID "1" |

### netskope_npa_local_broker

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| local_broker_name | x | x | x | x | | |
| city_name | | x | | x | x | |
| region_name | | x | | x | x | |
| country_name | | x | | x | x | |
| country_code | | x | | x | x | |
| latitude | | | | x | | Only in drift test |
| longitude | | | | x | | Only in drift test |
| custom_public_ip | | | | | x | Full config test |
| custom_private_ip | | | | | x | Full config test |
| access_via_public_ip | | | | x | x | "NONE" tested |

### netskope_npa_local_broker_config

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| hostname | x | x | | x | | |

### netskope_npa_policy_groups

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| group_name | x | | x | x | | Update test SKIPPED |
| group_order | x | | | x | | Not verified on import |

### netskope_npa_rules

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| rule_name | x | | x | x | | |
| description | x | | | x | | Not verified on import |
| enabled | x | x | | x | | "1"/"0" string values |
| group_id | x | | | x | | Not verified on import |
| rule_data.policy_type | x | | | x | | |
| rule_data.match_criteria_action | x | | | x | | Only "allow" tested |
| rule_data.private_apps | x | | | x | | References private app |
| rule_data.access_method | x | | | x | | "Client" only |
| rule_data.device_classification_id | | | | | | Tested indirectly via device classification tag resource |
| rule_order.order | x | | | | x | "top", "after", "before", "bottom" all tested (0.4.2) |
| rule_order.rule_id | | | | | x | BUG-003 regression test (0.4.2) |
| (concurrency) | | | | | x | 3 independent rules, exercises BUG-008 mutex (0.3.6) |

### netskope_npa_publisher_upgrade_profile

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| name | x | | x | x | | |
| enabled | x | x | | x | | true/false toggle |
| docker_tag | x | | | x | | From data source |
| frequency | x | | | x | | Cron format |
| release_type | x | | | x | | "Beta" only |
| timezone | x | | | x | | "US/Pacific" only |

### netskope_npa_private_app_public_host

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| host | x | x | | | | Update triggers replacement |
| real_host | x | x | | | | |
| clientless_access | x | | | | | |
| protocols | x | x | | | | 443 -> 8443 |

### netskope_npa_publisher_token

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| publisher_id | x | | | | | References publisher resource |

### netskope_rbac_label (v0.4.0)

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| name | x | x | x | | | |
| parent_id | | | | | x | Hierarchy test |
| color | x | x | x | | | |
| label_id | x | | x | | | Import by label_id |

### netskope_dns_profile_v2 (v0.4.0)

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| name | x | x | x | | | |
| description | x | x | | | | |
| log_traffic | x | | | | | |
| domain_config.allow_list | | | | | x | withDomainLists test |
| domain_config.block_list | | | | | x | withDomainLists test |
| domain_config.security_categories | | | | | x | withSecurityCategories test |
| custom_config | | | | | x | withCustomConfig test |
| tunnel_config | | | | | x | withTunnelConfig test |

### netskope_device_classification_tag (v0.4.2)

| Field | C | U | I | D | O | Notes |
|-------|---|---|---|---|---|-------|
| name | x | x | x | | | |
| description | x | x | | | | |
| tag_id | x | | x | | | Import by tag_id |
| priority | | | | | | Computed, server-managed |
| policy_names | | | | | | Computed, read-only |

### netskope_npa_publishers_alerts_configuration

All tests SKIPPED - valid event_types not documented by API.

---

## Data Source Coverage

| Data Source | Test Exists | Checks Performed | Notes |
|-------------|:-----------:|-----------------|-------|
| netskope_gre_tunnel | Yes | tunnel_id, site, source_ip via AttrPair | |
| netskope_gre_tunnels_list | Yes | result.# is set | |
| netskope_grepop | Yes | pop_id, pop_name, gateway | |
| netskope_grepo_ps_list | Yes | result.# is set | |
| netskope_ip_sec_tunnel | **NO** | Missing test file |
| netskope_ip_sec_tunnels_list | **NO** | Missing test file |
| netskope_ip_sec_pop | **NO** | Missing test file |
| netskope_ip_sec_po_ps_list | **NO** | Missing test file |
| netskope_npa_private_app | Yes | private_app_id, name, hostname via AttrPair |
| netskope_npa_private_apps_list | Yes | private_apps.# is set |
| netskope_npa_publisher | Yes | publisher_id, publisher_name via AttrPair |
| netskope_npa_publishers_list | Yes | data.publishers.# is set |
| netskope_npa_policy_groups | Yes | id, group_name via AttrPair |
| netskope_npa_policy_groups_list | Yes | data.# is set |
| netskope_npa_rules | Yes | id, rule_name via AttrPair |
| netskope_npa_rules_list | Yes | data.# is set |
| netskope_npa_local_broker | Yes | local_broker_id, name, access_via_public_ip |
| netskope_npa_local_brokers_list | Yes | data.# is set (also tests empty list) |
| netskope_npa_private_policy_in_use | **NO** | Missing test file |
| netskope_npa_publishers_host_os_versions | **NO** | Missing test file |
| netskope_npa_publishers_releases_list | No | Used indirectly by upgrade profile tests |
| netskope_npa_publisher_apps_list | **NO** | Missing test file |
| netskope_rbac_label | Yes | label_id, name, color via AttrPair | v0.4.0 |
| netskope_rbac_label_list | Yes | labels.#, total_count is set | v0.4.0 |
| netskope_dns_profile_v2 | Yes | name via AttrPair | v0.4.0 |
| netskope_dns_profile_v2_list | Yes | profiles.# is set | v0.4.0 |
| netskope_destination_profile | **SKIPPED** | Pending tenant license | v0.4.0 |
| netskope_destination_profile_list | **SKIPPED** | Pending tenant license | v0.4.0 |
| netskope_ips_status | **SKIPPED** | Pending tenant license | v0.4.0 |
| netskope_device_classification_tag | Yes | tag_id, name, description | v0.4.2 |
| netskope_device_classification_tag_list | Yes | tags.#, tags.0.tag_id, tags.0.name | v0.4.2 |
| netskope_device_classification_options_list | Yes | options.#, options.0.key, options.0.value | v0.4.2 |

---

## Test Dependencies

Tests that create supporting resources as part of their configuration.
Important for understanding cascading failures and cleanup.

| Test File | Creates These Dependencies |
|-----------|--------------------------|
| `npaprivateapp_resource_test.go` | Publisher |
| `npaprivateapp_data_source_test.go` | Publisher, Private App |
| `npaprivateappslist_data_source_test.go` | Publisher, Private App |
| `npaprivateapppublichost_resource_test.go` | None (standalone) |
| `npapublishertoken_resource_test.go` | Publisher |
| `nparules_resource_test.go` | Policy Group, Publisher, Private App (2nd rule for `ruleOrderAfter`; 3 independent rules for `concurrentCreate`) |
| `nparules_data_source_test.go` | Policy Group, Publisher, Private App, Rule |
| `nparuleslist_data_source_test.go` | Policy Group, Publisher, Private App, Rule |
| `gretunnel_data_source_test.go` | GRE Tunnel |
| `rbaclabel_resource_test.go` | None (standalone); `withPublisher` creates Publisher; `withPrivateApp` creates Publisher + Private App |
| `deviceclassificationtag_resource_test.go` | None (standalone) |
| `deviceclassificationtag_data_source_test.go` | None (uses existing tags on tenant) |

All other tests create only the resource under test.

---

## Skipped Tests

| Test | File | Reason |
|------|------|--------|
| `TestAccNPAPolicyGroups_update` | `npapolicygroups_resource_test.go` | Provider bug: SQLAlchemy bind parameter error, missing `new_order` parameter |
| `TestAccNPARules_denyRule` | `nparules_resource_test.go` | Block action requires DLP/Threat Protection profile configuration |
| `TestAccNPAPublishersAlertsConfiguration_basic` | `npapublishersalertsconfiguration_resource_test.go` | Valid event_types not documented; API rejects test values |

---

## Known Drift Issues

| Resource | Issue | Status |
|----------|-------|--------|
| `netskope_npa_rules` | Response-only fields caused plan drift; `group_id` not returned by GET API | **Fixed** (0.3.3) - hidden response-only fields, removed `group_id` from response schema |
| `netskope_ip_sec_tunnel` | `enable`/`enabled` field name mismatch between request/response | **Fixed** (0.3.3) - unified with `x-speakeasy-name-override` |
| `netskope_npa_private_app` | Non-deterministic publisher ordering from API | **Fixed** (0.3.4) - AfterSuccess hooks sort by publisher_id |
| `netskope_npa_private_app` | Config element reordering caused false diffs | **Fixed** (0.3.5) - ModifyPlan normalization for protocols, publishers, tags |
| `netskope_npa_rules` | Config element reordering caused false diffs | **Fixed** (0.3.5) - ModifyPlan normalization for private_apps, access_method |
| `netskope_gre_tunnel` | Config element reordering caused false diffs | **Fixed** (0.3.5) - ModifyPlan normalization for xff_ip_list |
| `netskope_ip_sec_tunnel` | Config element reordering caused false diffs | **Fixed** (0.3.5) - ModifyPlan normalization for pop_names |
| `netskope_npa_rules` | `rule_order.rule_id` type mismatch in BeforeRequest hook | **Fixed** (0.3.5) - BUG-003, changed `*string` to `*int64`, added `omitempty` |
| `netskope_npa_private_app` | Browser fields (`custom_host`, `allow_uri_bypass`, etc.) caused `(known after apply)` drift | **Fixed** (0.3.6) - `suppress-computed-diff` on all 8 browser fields |
| `netskope_npa_private_app` | `custom_host` silently ignored on write, causing config/state mismatch | **Fixed** (0.3.6) - Made computed-only (removed from request schemas) |
| `netskope_npa_rules` | `private_app_tag_ids` perpetual diff when using `private_app_tags` | **Fixed** (0.3.6) - BUG-006, `x-speakeasy-terraform-ignore` on `privateAppTagIds` |
| `netskope_npa_rules` | Tag name case sensitivity caused false diffs | **Fixed** (0.3.6) - BUG-005, case-insensitive comparison in ModifyPlan |
| `netskope_npa_rules` | Concurrent rule creation failed with duplicate primary key | **Fixed** (0.3.6) - BUG-008, provider-side mutex serializer |
| `netskope_npa_rules` | Rule creation failed after app creation (eventual consistency) | **Fixed** (0.3.6) - BUG-009, retry with exponential backoff |
| `netskope_npa_private_app` | `bypass_uris: []` on PUT caused SQL error | **Fixed** (0.3.6) - Hook strips empty `bypass_uris` from PUT requests |
| `netskope_npa_rules` | `device_classification_id` type mismatch (string vs int) | **Fixed** (0.3.6) - BeforeRequest hook coerces string→int |
| `netskope_npa_private_app` | `label_ids` not populated from API response | **Fixed** (0.4.0) - AfterSuccess hook populates from response |
| `netskope_device_classification_tag` | Create returns `{status,data:[id]}` not full object | **Fixed** (0.4.2) - AfterSuccess hook does follow-up GET |
| `netskope_device_classification_tag` | Update returns `{status:true}` not full object | **Fixed** (0.4.2) - AfterSuccess hook does follow-up GET |

---

## Import State Verify Ignore Lists

Fields excluded from import state verification because the API returns them
in a different format or does not return them at all.

| Resource | Ignored Fields | Reason |
|----------|---------------|--------|
| `netskope_gre_tunnel` | `pop_names` | API returns POP objects, not name strings |
| `netskope_ip_sec_tunnel` | `pop_names`, `psk` | POPs returned as objects; PSK is write-only |
| `netskope_npa_private_app` | `publishers`, `real_host`, `protocols` | Computed fields don't round-trip |
| `netskope_npa_policy_groups` | `group_order` | Not returned on read |
| `netskope_npa_rules` | `rule_order`, `rule_data`, `description`, `group_id` | Complex computed fields |
| `netskope_npa_publisher_upgrade_profile` | `next_update_time`, `created_at`, `updated_at` | Server-managed timestamps |
| `netskope_npa_local_broker` | `label_ids` | Write-only field |
| `netskope_rbac_label` | (none) | All fields verified on import |
| `netskope_device_classification_tag` | (none) | All fields verified on import |

---

## Drift Detection Tests

Located in `drift_detection_test.go`. Each test creates a resource, then runs
a second plan asserting `ExpectEmptyPlan()`.

| Test | Resource | Fields Exercised |
|------|----------|-----------------|
| `TestAccDrift_GRETunnel_Basic` | GRE Tunnel | site, source_ip, pop_names |
| `TestAccDrift_GRETunnel_WithOptions` | GRE Tunnel | + options.xff |
| `TestAccDrift_GRETunnel_AllFields` | GRE Tunnel | + source_type, bandwidth, enabled, notes |
| `TestAccDrift_IPSecTunnel_Basic` | IPSec Tunnel | site, source_ip, source_identity, psk, encryption, pop_names |
| `TestAccDrift_IPSecTunnel_WithOptions` | IPSec Tunnel | + options.rekey, options.reauth |
| `TestAccDrift_IPSecTunnel_AllFields` | IPSec Tunnel | + bandwidth, enabled, notes |
| `TestAccDrift_LocalBroker` | Local Broker | name, city, region, country, lat/long, access_via_public_ip |
| `TestAccDrift_LocalBrokerConfig` | Local Broker Config | hostname |
| `TestAccDrift_Publisher` | Publisher | publisher_name |
| `TestAccDrift_PrivateApp_Basic` | Private App | name, hostname, protocols, publishers |
| `TestAccDrift_PrivateApp_MultiProtocol` | Private App | + multiple TCP protocols |
| `TestAccDrift_PrivateApp_MultiPublisherWithTags` | Private App | + 2 publishers, mixed protocols, tags (0.3.4) |
| `TestAccDrift_PrivateApp_UnsortedProtocols` | Private App | Unsorted protocol list (0.3.5, BUG-002) |
| `TestAccDrift_PrivateApp_UnsortedAllLists` | Private App | Unsorted protocols, publishers, tags (0.3.5, BUG-002) |
| `TestAccDrift_PrivateApp_ReorderedConfig` | Private App | Config reorder between applies, expect empty plan (0.3.5, #56) |
| `TestAccDrift_PolicyGroup` | Policy Group | group_name, group_order |
| `TestAccDrift_NPARules_Basic` | NPA Rules | rule_name, description, enabled, group_id, rule_data |
| `TestAccDrift_NPARules_UnsortedLists` | NPA Rules | Unsorted private_apps, access_method (0.3.5, BUG-002) |
| `TestAccDrift_UpgradeProfile` | Upgrade Profile | name, enabled, docker_tag, frequency, timezone, release_type |
| `TestAccDrift_GRETunnel_UnsortedXffIpList` | GRE Tunnel | Unsorted xff_ip_list (0.3.5, BUG-002) |
| `TestAccDrift_GRETunnel_MinimalConfig` | GRE Tunnel | Minimal config, no optional computed drift (0.3.5) |
| `TestAccDrift_IPSecTunnel_MinimalConfig` | IPSec Tunnel | Minimal config, no optional computed drift (0.3.5) |

---

## Unit Tests

Plan modifier helper functions tested without API calls.

### Private App Plan Modifiers (`npaprivateapp_resource_planmodify_test.go`)

| Test | Function Tested | Purpose |
|------|----------------|---------|
| `TestProtocolKey_Basic` | `protocolKey` | TCP protocol key generation |
| `TestProtocolKey_UDP` | `protocolKey` | UDP protocol key generation |
| `TestPublisherKey_ByID` | `publisherKey` | Publisher key from ID |
| `TestPublisherKey_FallbackToName` | `publisherKey` | Fallback to name when ID missing |
| `TestPublisherKey_NullID_FallbackToName` | `publisherKey` | Null ID fallback |
| `TestPublisherKey_BothUnknown_ReturnsEmpty` | `publisherKey` | Both fields unknown |
| `TestTagKey_ByName` | `tagKey` | Tag key from name |
| `TestSortedKeysMatch_SameSetDifferentOrder` | `sortedKeysMatch` | Reordered lists match |
| `TestSortedKeysMatch_DifferentElements` | `sortedKeysMatch` | Different elements don't match |
| `TestSortedKeysMatch_DuplicateElements` | `sortedKeysMatch` | Duplicate handling |
| `TestSortedKeysMatch_DuplicateCountMismatch` | `sortedKeysMatch` | Different duplicate counts |
| `TestSortedKeysMatch_EmptyLists` | `sortedKeysMatch` | Empty lists match |
| `TestTagKeysMatch_CaseInsensitive` | `sortedKeysMatch` | Case-insensitive tag comparison (BUG-005) |
| `TestTagKeysMatch_CaseInsensitive_DifferentTags` | `sortedKeysMatch` | Different tags don't match regardless of case |

### NPA Rules Plan Modifiers (`nparules_resource_planmodify_test.go`)

| Test | Function Tested | Purpose |
|------|----------------|---------|
| `TestStringList_SameSetDifferentOrder` | `stringListSemanticEqual` | Reordered string lists match |
| `TestStringList_DifferentElements` | `stringListSemanticEqual` | Different elements don't match |
| `TestStringList_DifferentLength` | `stringListSemanticEqual` | Different lengths don't match |
| `TestStringList_EmptyLists` | `stringListSemanticEqual` | Empty lists match |
| `TestStringList_DuplicateElements` | `stringListSemanticEqual` | Duplicate handling |
| `TestStringList_DuplicateCountMismatch` | `stringListSemanticEqual` | Different duplicate counts |
| `TestInt64List_SameSetDifferentOrder` | `int64ListSemanticEqual` | Reordered int lists match |
| `TestInt64List_DifferentElements` | `int64ListSemanticEqual` | Different elements don't match |
| `TestInt64List_EmptyLists` | `int64ListSemanticEqual` | Empty lists match |
