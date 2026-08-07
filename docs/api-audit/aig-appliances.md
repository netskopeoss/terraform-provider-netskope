# API Audit: AIG Appliances (`/aig/appliances`)

Audited: 2026-06-05 against <tenant>.goskope.com

## Endpoints Tested

| Method | Path | OAS Status Code | Actual | Notes |
|--------|------|----------------|--------|-------|
| GET | `/aig/appliances` | 200 | 200 | Correct |
| POST | `/aig/appliances` | 201 | 201 | Correct |
| GET | `/aig/appliances/{id}` | 200 | 200 | Correct |
| PUT | `/aig/appliances/{id}` | 200 | 404 | **OAS WRONG — method is PATCH** |
| PATCH | `/aig/appliances/{id}` | — | 200 | Confirmed correct method |
| DELETE | `/aig/appliances/{id}` | 200 | 200 | Returns `{}` — correct |
| POST | `/aig/appliances/{id}/enrollmenttokens` | 200 | 200 | Correct |

## Field Audit

### Fields Correct in OAS (verified)

| Field | Type | Notes |
|-------|------|-------|
| `id` | string (uuid) | Computed, returned by all operations |
| `name` | string | Required, user-settable, editable via PATCH |
| `host` | string | Required, user-settable, editable via PATCH |
| `ports.http.enable` | boolean | Required, user-settable |
| `ports.http.port` | integer | Required, user-settable |
| `ports.https.enable` | boolean | Required, user-settable |
| `ports.https.port` | integer | Required, user-settable |
| `ai_provider_ids` | array of uuid strings | Optional, defaults to `[]` |
| `status` | string enum | Computed, read-only |
| `upgrade_profile_id` | string (uuid) | Computed — always defaulted, not user-settable |
| `ips` | array of strings | Computed, empty until after registration |
| `version` | string | Computed, empty until after registration |
| `certificate_imported` | boolean | Computed |
| `hypervisor_platform` | string | Computed, empty until after registration |
| `os_name` | string | Computed, empty until after registration |
| `uptime_day` | integer | Computed, 0 until after registration |
| `cpu_used` | number | Computed metrics |
| `cpu_avg` | array of numbers | Computed metrics |
| `memory_used` | number | Computed metrics |
| `partitions` | array of objects | Computed operational data |
| `reachability` | array of objects | Computed operational data |
| `create_time` | datetime string | Computed |
| `modify_time` | datetime string | Computed |
| `last_sync_time` | datetime string | Computed |

### Fields Missing from OAS (present in actual API responses)

| Field | Type | Value observed | Action |
|-------|------|----------------|--------|
| `mcp_server_ids` | array | `[]` | Add to schema + terraform-ignore |
| `sku_addons` | array | `[]` | Add to schema + terraform-ignore |
| `ui_onboarding_status` | integer | `0` | Add to schema + terraform-ignore |
| `enrollment_token` | string | JWT string | Present in CREATE response only, not GET. Add to schema + terraform-ignore |
| `enrollment_token_expire_time` | string (datetime) | ISO8601 string | Present in CREATE response only, not GET. Add to schema + terraform-ignore |

### OAS Inaccuracies Found

1. **Update HTTP method is PATCH, not PUT** — `PUT /aig/appliances/{id}` returns `404 "no Route matched"`. Fix: change `put:` to `patch:`.
2. **Create response schema `AigAppliance` is incomplete** — the actual 201 response includes `enrollment_token` and `enrollment_token_expire_time` at the top level. These are NOT included in the OAS schema.
3. **Three undocumented fields in all GET/PATCH responses** — `mcp_server_ids`, `sku_addons`, `ui_onboarding_status`.

### Partial PATCH behavior

Confirmed: PATCH accepts partial bodies. Sending only `{"name": "new-name"}` updates the name and leaves all other fields unchanged.

## Terraform Annotation Decisions

| Field | Decision | Reason |
|-------|----------|--------|
| `id` | Computed | UUID assigned by API, used as Terraform resource ID |
| `name` | Required | User-settable |
| `host` | Required | User-settable |
| `ports` | Required | User-settable |
| `ai_provider_ids` | Optional | User-settable, defaults to `[]` |
| `status` | Computed | Read-only API field; useful for outputs |
| `certificate_imported` | Computed | Useful status flag |
| `upgrade_profile_id` | terraform-ignore | Always auto-assigned; not user-settable at this endpoint |
| `ips`, `version`, `hypervisor_platform`, `os_name` | terraform-ignore | Populated only after registration; volatile |
| `uptime_day`, `cpu_used`, `cpu_avg`, `memory_used` | terraform-ignore | Volatile metrics; not useful in IaC state |
| `partitions`, `reachability` | terraform-ignore | Complex operational data; not useful in IaC state |
| `create_time`, `modify_time`, `last_sync_time` | terraform-ignore | Timestamps; noisy in state |
| `enrollment_token`, `enrollment_token_expire_time` | terraform-ignore | Only in create response; GET does not return them |
| `mcp_server_ids`, `sku_addons`, `ui_onboarding_status` | terraform-ignore | Undocumented internal fields |

## Resources and Data Sources Generated

| Type | Name | Description |
|------|------|-------------|
| Resource | `aig_appliance` | Full CRUD for AIG appliances |
| Resource | `aig_appliance_enrollment_token` | Generates an enrollment token for a given appliance |
| Data source | `aig_appliance` | Read a single appliance by ID |
| Data source | `aig_appliances` | List all appliances |
