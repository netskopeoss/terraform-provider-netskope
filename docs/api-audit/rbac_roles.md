# API Audit: RBAC Roles

**Date:** 2026-07-27  
**Endpoint:** `GET|POST /api/v2/rbac/roles`, `GET|PUT|DELETE /api/v2/rbac/roles/{role_id}`  
**Issue:** [netskopeoss/terraform-provider-netskope#100](https://github.com/netskopeoss/terraform-provider-netskope/issues/100)  
**RBAC apiGroup:** `roles`  
**Service:** `ms-rbac`

---

## Schema

### Resource fields

| Field | Type | Writable | Computed | Notes |
|-------|------|----------|----------|-------|
| `roleId` → `role_id` | integer | No | Yes | API-assigned on create |
| `roleName` → `name` | string | Yes | No | Required; unique per tenant |
| `roleDescription` → `description` | string | Yes | No | Required |
| `apiGroups` → `api_groups` | list of objects | Yes | No | Required; `apiGroupId` (int) + `permission` enum |
| `ipAllowList` → `ip_allow_list` | object | Yes | No | Optional; `enableIpAllowList` (bool) + `ipList` ([]string) |

### Ignored fields (API-computed)

| Field | Notes |
|-------|-------|
| `version` | String `"3.0"` in all responses |
| `scopes` | Scope configuration (complex, skipped for MVP) |
| `isAliasNameTaken` | Boolean, read-only |
| `labels.assignedLabels` | Label assignments |

---

## OAS Inaccuracies Fixed

| Field | Official OAS | Actual API |
|-------|-------------|-----------|
| `version` type | `number` (inferred) | String `"3.0"` — causes `cannot unmarshal string into float64` |
| Create/update response | Full role object | `{"roleId": N}` only — hook fetches full GET after |
| `ipAllowList.ipList` (GET) | `string[]` | Array of `{ipAddress, createdAt, updatedAt}` objects — hook normalizes to strings |
| List fields | `roleName`, `roleDescription` | `name`, `description` (different field names than single GET) |
| `roleId` | string | integer |

---

## Behavior

- **Create** (`POST`): Returns `{"roleId": N}` with HTTP 201. Hook fetches `GET /rbac/roles/{N}` and returns full role.
- **Read** (`GET /{id}`): Returns full role with all fields. `ipAllowList.ipList` is array of objects (hook normalizes to strings).
- **Update** (`PUT /{id}`): Full replace — all fields required. Returns `{"roleId": N}` with HTTP 200. Hook fetches GET.
- **Delete** (`DELETE /{id}`): Returns `{"status": "success"}` on success.
- **List** (`GET`): Returns `{"roles": [...], "count": N, "version": "3.0"}`. List items use `name`/`description` not `roleName`/`roleDescription`.

---

## Verified via live API (<tenant>.goskope.com)

| Operation | Result |
|-----------|--------|
| `POST` with name+description+apiGroups | `201`, `{"roleId": N}` |
| `GET /{id}` | `200`, full role with `ipAllowList.ipList` as objects |
| `PUT /{id}` | `200`, `{"roleId": N}` |
| `DELETE /{id}` | `200`, `{"status": "success"}` |
| `GET /{id}` (404) | `{"statusCode": 404, "message": ["Invalid Role Id"], "error": "Invalid Role Id"}` |

---

## Hooks

- **`hookRBACRoleAfterSuccess.go`**:
  - After `createRBACRole`/`replaceRBACRoleDetails`: parse `roleId`, GET full role, normalize response
  - After `getRBACRole`: normalize `ipAllowList.ipList` objects → strings

---

## Drift Risks

- **apiGroups ordering**: The API may not return api_groups in insertion order. If drift observed, add sorting in hook.
- **ipAllowList**: When `enableIpAllowList=false`, API returns empty `ipList`. If not set in config, Terraform will show diff on next plan. Set `ip_allow_list = {}` or explicitly set both fields.

---

## Endpoints Not Exposed

- `scope`/`scopes`: Uses `\!in` key (problematic for HCL attribute names). Deferred.
- `labels.assignedLabels`: Deferred to future work.
- Per-apiGroup `obfuscation`, `constraints`: Deferred.
