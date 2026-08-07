# API Audit: DNS Profiles v2

**Date:** 2026-03-06
**Tenant:** <tenant>.goskope.com
**Base path:** `/api/v2/profiles/dns`

## Endpoints Tested

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | `/profiles/dns?limit=2` | 200 | First call returned 400 (migration in progress), retry succeeded |
| POST | `/profiles/dns` | 201 | Auto-deployed (interactive=false default) |
| GET | `/profiles/dns/{id}` | 200 | Returns profile object directly (no wrapper) |
| PATCH | `/profiles/dns/{id}` | 200 | Partial update works, auto-deploys |
| DELETE | `/profiles/dns/{id}` | 200 | Returns `{"status": "success"}` |

## Schema: Computed Fields (not user-settable)

| Field | Type | Notes |
|-------|------|-------|
| `status` | string | `Applied`, `Pending-create`, `Pending-update`, `Pending-delete` |
| `create_by` | string | Email or "Netskope" for system-created profiles |
| `create_time` | string | Format: `"Fri, 06 Mar 2026 21:09:10 GMT"` |
| `modify_by` | string | Email |
| `modify_time` | string | Same format as create_time |
| `applied_time` | string | Format: `"2026-03-06 21:09:10"` — only on POST/PATCH responses, NOT on GET |

## Schema: Editable Fields

| Field | Required (create) | Notes |
|-------|-------------------|-------|
| `name` | Yes | Must be unique |
| `description` | No | |
| `log_traffic` | No | Enum: `"Blocked DNS"`, `"All DNS"`. Defaults to `"Blocked DNS"` |
| `domain_config` | No | Object — see nested fields below |
| `tunnel_config` | No | Object — see nested fields below |
| `custom_config` | No | Object — see nested fields below |
| `inheritance_groups` | No | Array of strings, max 1 group |

## OAS vs API Discrepancies

### 1. `tunnel_config.allow_list` enum casing

**OAS says:** `DNS2TCP`, `Evasive Protocol`, `Iodine`, `AnalogBit TCP-over-DNS`, `VPNoverDNS`
**API returns:** `dns2tcp`, `Evasive protocol`, `iodine`, `AnalogBit tcp-over-dns`, `vpnoverdns`

**Resolution:** Removed enums from OAS. Free-form string to avoid casing mismatches.

### 2. `security_categories[].name` casing

**OAS says:** `Security Risk - Command And Control Server`
**API returns:** `Security Risk - Command and Control server`

**Resolution:** Kept category names as free-form strings (no enum). Category names come from the API's domain categories endpoint and casing varies.

### 3. `applied_time` field missing from OAS

**API returns** `applied_time` on POST and PATCH responses but NOT on GET.

**Resolution:** Added to response schema with `x-speakeasy-terraform-ignore: true`.

### 4. `business_categories` is an unknown field

The API rejects `business_categories` entirely on this tenant:
```json
{"err_code":400,"errorMsg":"Invalid request body: {'business_categories': ['Unknown field.']}"}
```
Sending even an empty `business_categories: []` in the request body triggers the error. This is a tenant-feature-gated field.

**Resolution:** Removed `business_categories` from the OAS entirely. If a future tenant has this enabled, it can be re-added behind a feature flag or as a separate schema variant.

### 5. Error response shape

**OAS says:** `err_code` is string (`"310101"`)
**API returns:** `err_code` is integer (`400`), plus `errorMsg` field alongside `message`

**Resolution:** Fixed `err_code` to integer, added `errorMsg` field.

### 6. Delete response

**OAS says:** Nothing specified
**API returns:** `{"status": "success"}`

**Resolution:** Added to OAS.

## Default Values (from API)

When creating a profile with only `{"name": "..."}`, the API returns these defaults:

```json
{
  "log_traffic": "Blocked DNS",
  "domain_config": {
    "allow_list": [],
    "block_all_except_allow_list": false,
    "block_list": [],
    "security_categories": [],
    "sinkhole_ip": ""
  },
  "tunnel_config": {
    "allow_list": [],
    "enable": false
  },
  "custom_config": {
    "bypass_original_dns": false,
    "enable": false,
    "fallback_to_netskope_dns": true,
    "server_ip": []
  }
}
```

## Potential Terraform Drift Risks

1. **Empty arrays vs absent** — `domain_config.security_categories` is `[]` when empty. If user doesn't set it in config, Terraform may show drift on every plan. May need `suppress-computed-diff` or hook to normalize.
2. **`sinkhole_ip` defaults to `""`** — empty string vs null could cause drift.
3. **`tunnel_config.allow_list` casing** — if user provides PascalCase in HCL but API returns lowercase, perpetual diff. May need a normalization hook.

## Speakeasy Version

- **1.728.0** — pinned. Version 1.748.0 fails with `failed to make environment variable mapping: the following specified provider attributes were not found: [api_key]`. This is a Speakeasy regression, not related to our OAS.
- **json.go revert required** — both 1.728.0 and 1.748.0 introduce an `omitEmpty` change that skips empty slices. Must revert `internal/sdk/internal/utils/json.go` after every `speakeasy run` (see CLAUDE.md pitfall #2).

## Acceptance Tests

All 7 tests pass against `<tenant>.goskope.com`:

| Test | Coverage |
|------|----------|
| `TestAccDNSProfileV2_basic` | Create + import state verify |
| `TestAccDNSProfileV2_update` | Update name, description, log_traffic |
| `TestAccDNSProfileV2_withSecurityCategories` | domain_config with security categories + sinkhole_ip |
| `TestAccDNSProfileV2_withTunnelConfig` | tunnel_config enabled |
| `TestAccDNSProfileV2_withCustomConfig` | custom_config with server_ip |
| `TestAccDNSProfileV2_withDomainLists` | domain_config allow_list + block_list |
| `TestAccDNSProfileV2_import` | Import state only |

Run with:
```bash
NETSKOPE_SERVER_URL="https://<tenant>.goskope.com/api/v2" \
NETSKOPE_API_KEY="<key>" \
TF_ACC=1 go test -v ./internal/provider/... -run TestAccDNSProfileV2 -timeout 30m
```

## OAS File

Extracted and annotated: `endpoints/profiles/dns_profiles_v2.yaml`
Entity name: `DNSProfileV2`