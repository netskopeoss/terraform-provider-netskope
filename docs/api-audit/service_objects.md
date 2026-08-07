# API Audit: Service Objects

**Date:** 2026-07-27  
**Endpoint:** `GET|POST /api/v2/profiles/serviceobjects`, `GET|PATCH|DELETE /api/v2/profiles/serviceobjects/{id}`  
**Issue:** [netskopeoss/terraform-provider-netskope#101](https://github.com/netskopeoss/terraform-provider-netskope/issues/101)  
**RBAC apiGroup:** `objects_service`  
**Service:** `qos`

---

## Schema

| Field | Type | Writable | Computed | Notes |
|-------|------|----------|----------|-------|
| `id` | string (UUID) | No | Yes | Assigned by API on create |
| `name` | string | Yes | No | Required; unique per tenant |
| `description` | string | Yes | No | Required (API rejects missing description) |
| `protocols.icmp` | boolean | Yes | No | Optional |
| `protocols.tcp` | `[]string` | Yes | No | Port numbers or ranges, e.g. `"443"`, `"8080-9090"` |
| `protocols.udp` | `[]string` | Yes | No | Port numbers or ranges |
| `protocols.tcp_udp` | `[]string` | Yes | No | Port numbers or ranges |
| `type` | string | No | Yes | `custom` or `PREDEFINED`; always `custom` for user-created objects |
| `status` | string | No | Yes | `applied`, `pending-create`, etc. |
| `create_by` | string | No | Yes | |
| `create_time` | string | No | Yes | Non-ISO format: `"Mon, 27 Jul 2026 09:09:01 GMT"` |
| `modify_by` | string | No | Yes | |
| `modify_time` | string | No | Yes | Non-ISO format |

---

## OAS Inaccuracies Fixed

| Field | Official OAS | Actual API |
|-------|-------------|-----------|
| `status` enum | `pending-create`, `pending-update`, `pending-delete`, `applied` (lowercase) | Lowercase for custom objects; uppercase `APPLIED` for predefined. Annotated as lowercase (user-created objects only) |
| `type` enum | `custom`, `predefine` | Lowercase `custom` and uppercase `PREDEFINED`. Annotated as both |
| `id` example | `"123"` (numeric string) | UUID string (`"cf353554-899a-11f1-96ca-4add69fefc7f"`) |
| Base path | `/serviceobjects` | `/profiles/serviceobjects` — the official OAS is relative to the service root, not the API root |
| List response key | `services` | `services` ✓ |
| List total key | `total` | `total` ✓ |
| `limit` max | 500 (Speakeasy default) | 150 — API rejects `limit=500` with HTTP 400 |
| `type` enum | `"custom"`, `"predefine"` | Returns `"CUSTOM"` (uppercase) for user objects, `"PREDEFINED"` for built-ins |
| `status` enum (predefined) | `"applied"` | Returns `"APPLIED"` (uppercase) for predefined objects |
| `err_code` in error response | string | integer (e.g. `400`) — OAS schema bug |

---

## Behavior

- **Create** (`POST`): Returns `201` when `interactive=false` (default, auto-deploys). Requires `name`, `description`, `protocols` (all three).
- **Read** (`GET /{id}`): Returns all set protocol fields; **unset protocol types are omitted** (not null, not empty array — absent).
- **Update** (`PATCH /{id}`): Partial update; returns `200` when auto-deployed.
- **Delete** (`DELETE /{id}`): Returns `{"status": "success"}` on immediate delete. Returns `409` if object is referenced by a policy.
- **List** (`GET`): `services[]` + `total`. Includes both custom and predefined objects.

---

## Verified via live API (<tenant>.goskope.com)

| Operation | Result |
|-----------|--------|
| `POST` with name+description+protocols.tcp | `201`, `type: custom`, `status: applied` |
| `POST` without description | `400`, "Missing data for required field" |
| `POST` with empty protocols `{}` | `400`, "At least one protocol must be provided" |
| `POST` with protocols.icmp=true only | `201`, `protocols: {icmp: true}` (no other keys) |
| `GET /{id}` | `200`, only set protocol fields returned |
| `PATCH /{id}` with tcp+udp | `200`, both returned |
| `DELETE /{id}` | `200`, `{"status": "success"}` |

---

## Drift Risks

- **Protocol array ordering**: Port arrays (`tcp`, `udp`, `tcp_udp`) may not be returned in insertion order. Added `x-speakeasy-param-suppress-computed-diff: true` on each port array. Add a sort hook if drift is observed in tests.
- **Absent vs null protocols**: Unset protocol types are absent from the API response (not `null` or `[]`). Terraform users should only set protocol fields they intend to use. Setting an empty list (e.g. `protocols.tcp = []`) would send an empty array to the API — behavior untested; likely rejected.
- **Predefined objects**: The list returns both custom and predefined objects. Predefined objects have `type: PREDEFINED` and cannot be modified (the Terraform resource only creates `custom` objects).

---

## Endpoints Not Exposed

- `POST /profiles/serviceobjects/deploy` — not needed; `interactive=false` auto-deploys all write operations.
- `GET /profiles/serviceobjects/{id}/versions/{identifier}` — read-only version history.
- `POST /profiles/serviceobjects/{id}/revert` — Terraform manages state directly.
