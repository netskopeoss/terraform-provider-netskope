# API Audit: URL Lists

**Date:** 2026-05-19
**Tenant:** alliances.goskope.com
**Base path:** `/api/v2/policy/urllist`

## Endpoints Tested

| Method | Path | Status | Notes |
|--------|------|--------|-------|
| GET | `/policy/urllist` | 200 | Returns array of all URL lists |
| POST | `/policy/urllist` | 201 | Returns array containing the created item; pending until deploy |
| GET | `/policy/urllist/{id}` | 200 | Returns single object (not array) |
| PUT | `/policy/urllist/{id}` | 200 | Full replace; returns single object; pending until deploy |
| DELETE | `/policy/urllist/{id}` | 200 | Returns single object with modify_type "Deleted"; pending until deploy |
| PATCH | `/policy/urllist/{id}/{action}` | 200 | Action = "append" or "replace"; returns single object |
| POST | `/policy/urllist/deploy` | 200 | Applies all pending changes; returns array of deployed items |
| POST | `/policy/urllist/file` | 200 | Multipart upload (not needed for Terraform) |

## Deploy Model

This API uses a **two-phase commit** pattern:
1. CRUD operations create **pending** changes (`pending: 1`)
2. `POST /policy/urllist/deploy` applies all pending changes (`pending: 0`)

**Terraform implication:** Hooks must auto-deploy after every create/update/delete, otherwise changes remain pending and subsequent reads return stale data.

## Schema: Computed Fields (not user-settable)

| Field | Type | Notes |
|-------|------|-------|
| `id` | integer (int64) | Assigned by API on create |
| `modify_by` | string | e.g., "tf-test-service-role" |
| `modify_time` | string | ISO 8601 format: `2026-05-19T13:03:32.000Z` |
| `modify_type` | string | "Created", "Edited", "Deleted" |
| `pending` | integer | 0 = applied, 1 = pending |
| `data.json_version` | integer | Always `2`; **NOT in OAS** — must add and mark computed |

## Schema: Editable Fields

| Field | Required (create) | Notes |
|-------|-------------------|-------|
| `name` | **Yes** | 400 if missing; restricted characters: `< > / ! @ # $ % ^ & * ( ) { } ; + = , ? . \| : ' "` |
| `data` | **Yes** | 400 if missing or null |
| `data.urls` | **Yes** | Must be array of strings; 400 if missing |
| `data.type` | **Yes** | Enum: `exact`, `regex`; 400 if missing or invalid |

## OAS vs API Discrepancies

### 1. `data.json_version` missing from OAS

**API returns:** `"json_version": 2` inside `data` object on all responses.
**OAS says:** Field does not exist.

**Resolution:** Add to response schema with `x-speakeasy-terraform-ignore: true` (computed, not user-settable).

### 2. `modify_time` format incorrect in OAS

**OAS example:** `"1997-01-01 00:00:00"`
**API returns:** `"2026-05-19T13:03:32.000Z"` (ISO 8601 with timezone)

**Resolution:** Correct the example; type remains `string`.

### 3. `RequestSchema` name is too generic

**OAS defines:** `RequestSchema` used for create/update request body.
**Problem:** Name collision risk in a multi-endpoint provider.

**Resolution:** Rename to `UrllistRequest` in the extracted OAS.

### 4. POST create returns array, GET by ID returns object

**POST `/policy/urllist`:** Returns `[{...}]` (array with single item)
**GET `/policy/urllist/{id}`:** Returns `{...}` (single object)

**Resolution:** OAS already reflects this correctly. Hook will need to extract the item from the array on create.

### 5. Path parameter `id` uses `content.text/plain` instead of `schema`

**OAS defines:** `id` path parameter with `content: { text/plain: { schema: { type: number } } }` instead of the standard `schema: { type: integer }`.

**Resolution:** Use standard `schema` format in the extracted OAS. Also change type from `number` to `integer` to match actual int64 IDs.

### 6. `pending` field type

**OAS says:** `type: integer` (correct)
**API returns:** 0 or 1 (integer, confirmed)

**Resolution:** No change needed.

## Terraform Resource Design

### Resource: `netskope_urllist`

**Editable attributes:**
- `name` (Required, string)
- `data` (Required, object)
  - `urls` (Required, list of strings)
  - `type` (Required, string, enum: exact/regex)

**Computed attributes:**
- `id` (int64, used for import)
- `modify_by`, `modify_time`, `modify_type`, `pending` — all computed metadata

### Data Sources

- `netskope_urllist` — Get single URL list by ID
- `netskope_urllist_list` — Get all URL lists

### Hooks Needed

1. **AfterSuccess on Create** — POST returns array, extract single item; then auto-deploy
2. **AfterSuccess on Update** — auto-deploy after PUT
3. **AfterSuccess on Delete** — auto-deploy after DELETE
4. **BeforeRequest** — may need to strip computed fields (`json_version`) from request body
