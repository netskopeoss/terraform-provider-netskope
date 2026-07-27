# API Audit: Custom Categories

**Date:** 2026-07-27  
**Endpoint:** `GET|POST /api/v2/profiles/customcategories`, `GET|PATCH|DELETE /api/v2/profiles/customcategories/{id}`  
**Issue:** [netskopeoss/terraform-provider-netskope#99](https://github.com/netskopeoss/terraform-provider-netskope/issues/99)  
**RBAC apiGroup:** `objects_custom_category`  
**Service:** `swg-profile-service`

---

## Schema

All association ID fields are `type: string` in the API — **including URL list IDs** which are numeric integers in the URL list resource but passed as numeric strings here (e.g. `"22"`, not `22`).

| Field | Type | Writable | Computed | Notes |
|-------|------|----------|----------|-------|
| `id` | string (UUID) | No | Yes | Assigned by API on create |
| `name` | string | Yes | No | Required; unique per tenant; max 100 chars |
| `description` | string | Yes | No | Optional; max 200 chars |
| `included_predefined_categories` | `[]string` | Yes | No | Numeric string IDs (e.g. `"500"`) |
| `included_url_lists` | `[]string` | Yes | No | Numeric string IDs (e.g. `"22"`) |
| `excluded_url_lists` | `[]string` | Yes | No | Numeric string IDs |
| `included_destination_profiles` | `[]string` | Yes | No | UUID strings |
| `excluded_destination_profiles` | `[]string` | Yes | No | UUID strings |
| `status` | string | No | Yes | `applied`, `pending-create`, `pending-update`, `pending-delete` |
| `create_by` | string | No | Yes | |
| `create_time` | string | No | Yes | ISO 8601 |
| `modify_by` | string | No | Yes | |
| `modify_time` | string | No | Yes | ISO 8601 |

---

## Behavior

- **Create** (`POST`): Returns `201` when `interactive=false` (default) — category is immediately applied.
- **Read** (`GET /{id}`): Always returns all fields including empty arrays for unset associations.
- **Update** (`PATCH /{id}`): Partial update. Returns `200` when auto-deployed (`interactive=false` default).
- **Delete** (`DELETE /{id}`): Returns `{"status": "success"}` on immediate delete. Returns `409` if the category is referenced by a policy.
- **List** (`GET`): Response has `elements[]` + `total_count`. Default limit=10, max=500.

---

## Verified via live API (bespin.goskope.com)

| Operation | Result |
|-----------|--------|
| `POST` with name+description only | `201`, all array fields returned as `[]` |
| `GET /{id}` | `200`, all fields present |
| `PATCH /{id}` with empty arrays | `200`, applied immediately |
| `PATCH` with `included_predefined_categories: ["Financial_Services"]` | `400`, "must be a valid number" — **string names rejected, must use numeric string IDs** |
| `PATCH` with `included_url_lists: [1]` (integer) | `400`, "incorrect value type" — **must be string** |
| `DELETE /{id}` | `200`, `{"status": "success"}` |

---

## Drift Risks

- **Array ordering**: The API may not return `included_url_lists`, `included_predefined_categories`, `included_destination_profiles`, `excluded_*` in insertion order. Added `x-speakeasy-param-suppress-computed-diff: true` to all five array fields in the response schema. Monitor after first acceptance test runs — if ordering drift occurs, add a sort hook following the BUG-001 pattern.
- **`total_count`**: Ignored in Terraform schema (`x-speakeasy-terraform-ignore`).

---

## Endpoints Not Exposed

- `POST /profiles/customcategories/deploy` — not needed; `interactive=false` auto-deploys all write operations.
- `POST /profiles/customcategories/{id}/revert` — Terraform manages state directly; revert is a UI workflow.
- `GET /profiles/customcategories/{id}/versions/{version}` — read-only version history; not relevant to Terraform.
