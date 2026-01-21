# Known API Issues and Workarounds

This document tracks known behavioral inconsistencies and quirks in the Netskope REST API that this Terraform provider works around or that developers should be aware of.

## Table of Contents

- [Steering API Issues](#steering-api-issues)
  - [1. App Name Automatic Encapsulation](#1-app-name-automatic-encapsulation)
  - [2. Inconsistent Name Key in Responses](#2-inconsistent-name-key-in-responses)
  - [3. Delete Verification Returns 200 OK with Error](#3-delete-verification-returns-200-ok-with-error)
  - [8. Empty Objects Cause SQL Serialization Error on Update](#8-empty-objects-cause-sql-serialization-error-on-update)
  - [9. Protocol Field Name Mismatch (type vs transport)](#9-protocol-field-name-mismatch-type-vs-transport)
  - [10. Write-Only Fields Not Returned in Response](#10-write-only-fields-not-returned-in-response)
- [Policy API Issues](#policy-api-issues)
  - [11. Policy Group Response Wrapper Mismatch](#11-policy-group-response-wrapper-mismatch)
  - [12. NPA Rules Response Wrapper Mismatch](#12-npa-rules-response-wrapper-mismatch)
- [Infrastructure API Issues](#infrastructure-api-issues)
  - [4. Missing external_id on Profile Creation](#4-missing-external_id-on-profile-creation)
  - [5. GET on Deleted Resource Returns 200 OK](#5-get-on-deleted-resource-returns-200-ok)
  - [6. id vs external_id Confusion](#6-id-vs-external_id-confusion)
  - [7. Inconsistent Publisher Field Names](#7-inconsistent-publisher-field-names)
- [Terraform Provider Implications](#terraform-provider-implications)
- [General Recommendations](#general-recommendations)

---

## Steering API Issues

### 1. App Name Automatic Encapsulation

**Endpoint:** `POST /api/v2/steering/apps/private`

**Issue:** The API automatically wraps application names with brackets `[]` upon creation.

**Example:**

```json
// Payload sent
{"app_name": "TF_test"}

// Response received
{"app_name": "[TF_test]"}  // Name is automatically wrapped
```

**Impact:**
- Subsequent queries for the application by name must use the bracketed version
- Name comparisons may fail if not accounting for the brackets
- Terraform state may show drift between configured and actual names

**Status:** Known API behavior - No fix planned

---

### 2. Inconsistent Name Key in Responses

**Endpoints:**
- `GET /api/v2/steering/apps/private/{id}` (single app)
- `GET /api/v2/steering/apps/private` (list apps)

**Issue:** Different endpoints use different JSON keys for the application name:

| Endpoint | Key Used |
|----------|----------|
| Single app GET | `name` |
| List apps GET | `app_name` |

**Example:**

```json
// Get single app - uses "name"
{"data": {"name": "MyApp", ...}}

// List apps - uses "app_name"
{"data": {"private_apps": [{"app_name": "MyApp", ...}]}}
```

**Impact:** Code processing API responses must handle both key names depending on the endpoint.

**Status:** Known API inconsistency

---

### 3. Delete Verification Returns 200 OK with Error

**Endpoint:** `GET /api/v2/steering/apps/private/{id}`

**Issue:** When querying for a deleted (non-existent) application, the API returns HTTP 200 OK with an error in the response body instead of HTTP 404 Not Found.

**Expected Behavior:**
```
HTTP/1.1 404 Not Found
```

**Actual Behavior:**
```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "message": "No private app with id '233' is found.",
  "status": "error"
}
```

**Impact:**
- Cannot rely on HTTP status codes alone to detect missing resources
- Must parse response body to detect errors
- Terraform Read operations must check response body to properly remove deleted resources from state

**Status:** Must be handled by checking `"status": "error"` in response body

---

### 8. Empty Objects Cause SQL Serialization Error on Update

**Endpoint:** `PUT /api/v2/steering/apps/private/{id}`

**Issue:** When updating a private app, if the request includes empty objects (`{}`) or empty arrays (`[]`) for certain fields (`app_option`, `paths`), the API backend fails with a SQL serialization error. The backend incorrectly converts these empty JSON values to Python bytes objects.

**Error Message:**
```
(raised as a result of Query-invoked autoflush; consider using a session.no_autoflush block if this flush is occurring prematurely)
(builtins.TypeError) Object of type bytes is not JSON serializable
[SQL: UPDATE private_apps SET modify_time=%(modify_time)s, app_option=%(app_option)s, paths=%(paths)s WHERE private_apps.app_id = %(private_apps_app_id)s]
[parameters: [{'app_option': b'{}', 'paths': b'[]', 'private_apps_app_id': 525}]]
```

**Affected Fields:**
- `app_option` - empty object `{}`
- `paths` - empty array `[]`
- `uribypass_header_value` - null/empty string

**Impact:**
- All private app update operations fail if these fields are included with empty values
- Even though the JSON payload is valid, the API backend mishandles it

**Workaround:** The Terraform provider uses a BeforeRequest hook (`hookPrivateAppRequest.go`) to strip these fields from PUT requests when they are empty.

**Status:** API backend bug - Workaround implemented in provider via SDK hook

---

### 9. Protocol Field Name Mismatch (type vs transport)

**Endpoints:**
- `POST /api/v2/steering/apps/private` (create)
- `GET /api/v2/steering/apps/private/{id}` (read)

**Issue:** The protocol field uses different names in requests vs responses:

| Operation | Field Name |
|-----------|------------|
| Request (POST/PUT) | `type` |
| Response (GET) | `transport` |

**Example:**

**Request:**
```json
{
  "protocols": [
    {"type": "tcp", "ports": ["443"]}
  ]
}
```

**Response:**
```json
{
  "protocols": [
    {"transport": "tcp", "port": "443", "id": 123, ...}
  ]
}
```

**Impact:**
- Cannot use the same data structure for requests and responses
- Terraform state mapping requires transformation between `type` and `transport`

**Workaround:** The provider uses an AfterSuccess hook (`hookMyAppAfterSuccess.go`) to map `transport` to `type` in responses, maintaining consistency with the request schema.

**Status:** Known API inconsistency - Handled via SDK hook

---

### 10. Write-Only Fields Not Returned in Response

**Endpoint:** `GET /api/v2/steering/apps/private/{id}`

**Issue:** Several fields that can be set during create/update are not returned in GET responses:

| Field | Accepted in Request | Returned in Response |
|-------|---------------------|---------------------|
| `allow_uri_bypass` | Yes | No |
| `app_option` | Yes | Empty `{}` only |
| `paths` | Yes | Empty `[]` only |

**Impact:**
- Terraform cannot track the actual state of these fields
- Causes perpetual drift in `terraform plan` output
- Even if explicitly set in configuration, the next plan shows changes

**Example:**
```hcl
# Configuration
resource "netskope_npa_private_app" "example" {
  allow_uri_bypass = false  # Set explicitly
  ...
}

# Plan output (perpetual drift)
+ allow_uri_bypass = false  # Always shows as addition
```

**Workaround:**
- Use `x-speakeasy-param-suppress-computed-diff: true` in OpenAPI spec for these fields
- Mark `app_option` as `x-speakeasy-terraform-ignore: true` to exclude from schema
- Accept minor cosmetic drift for `allow_uri_bypass` (defaults to false, functionally correct)

**Status:** API limitation - Partial workaround via Speakeasy annotations

---

## Policy API Issues

### 11. Policy Group Response Wrapper Mismatch

**Endpoints:**
- `POST /api/v2/policy/npa/policygroups` (create)
- `PATCH /api/v2/policy/npa/policygroups/{id}` (update)
- `GET /api/v2/policy/npa/policygroups/{id}` (read)

**Issue:** The API wraps responses in a `{"data": {...}, "status": "success"}` envelope.

**Example API response:**
```json
{
  "data": {
    "group_id": "145",
    "group_name": "my-group",
    "group_type": "0",
    ...
  },
  "status": "success"
}
```

**Resolution:** The OpenAPI specification has been updated to correctly define the wrapped response structure for all policy group endpoints. The SDK now expects and properly handles the `{"data": {...}, "status": "..."}` envelope natively.

**Status:** ✅ RESOLVED - OAS correctly defines wrapped response structure

---

### 12. NPA Rules Response Wrapper Mismatch

**Endpoints:**
- `POST /api/v2/policy/npa/rules` (create)
- `PATCH /api/v2/policy/npa/rules/{id}` (update)
- `GET /api/v2/policy/npa/rules/{id}` (read)
- `DELETE /api/v2/policy/npa/rules/{id}` (delete)

**Issue:** Similar to Issue #11, the API wraps all responses in a `{"data": {...}, "status": "success"}` envelope.

**Example API response:**
```json
{
  "data": {
    "rule_id": "4",
    "rule_name": "my-rule",
    "enabled": "1",
    "rule_data": {...}
  },
  "status": "success"
}
```

**Resolution:** The OpenAPI specification has been updated to correctly define the wrapped response structure for all NPA rules endpoints. The SDK now expects and properly handles the `{"data": {...}, "status": "..."}` envelope natively.

**Status:** ✅ RESOLVED - OAS correctly defines wrapped response structure

---

### 13. Response Wrapper `status` Field Excluded from Terraform Schema

**Endpoints:** All policy endpoints that return wrapped responses:
- `POST/GET/PATCH/DELETE /api/v2/policy/npa/rules`
- `POST/GET/PATCH/DELETE /api/v2/policy/npa/policygroups`

**Issue:** The API response wrapper includes a `status` field (`"success"` or `"error"`) at the envelope level, not as a property of the resource itself:

```json
{
  "data": { ...resource properties... },
  "status": "success"  // <-- This is response-level, not resource-level
}
```

The `status` field indicates whether the API call succeeded, not a property of the rule or policy group resource. Including it in the Terraform resource schema would be misleading and cause issues since it's not a persistent attribute of the resource.

**Resolution:** The `status` field is excluded from the Terraform resource schema using `x-speakeasy-terraform-ignore: true` in the OpenAPI specification. The SDK still parses the wrapped response correctly, but the `status` field is not exposed as a Terraform attribute.

**Status:** ✅ RESOLVED - `status` field excluded from Terraform schema via OAS annotation

---

## Infrastructure API Issues

### 4. Missing external_id on Profile Creation

**Endpoint:** `POST /api/v2/infrastructure/publisherupgradeprofiles`

**Issue:** The creation response only includes `id`, but all other operations (get, update, delete) require `external_id` in the URL path.

**Example:**

```json
// POST response - only returns 'id'
{"data": {"id": 29, "name": "Profile", ...}}

// But GET/PUT/DELETE use external_id in path:
// GET /api/v2/infrastructure/publisherupgradeprofiles/29
```

**Impact:** The `id` returned from POST must be used as the `external_id` for subsequent operations.

**Status:** Known API behavior - Workaround: use `id` from create response as `external_id` for other operations

---

### 5. GET on Deleted Resource Returns 200 OK

**Endpoint:** `GET /api/v2/infrastructure/publisherupgradeprofiles/{id}`

**Issue:** Similar to Issue #3, querying for a deleted profile returns HTTP 200 OK with an error body instead of HTTP 404.

**Actual Behavior:**
```
HTTP/1.1 200 OK
Content-Type: application/json

{
  "message": "No publisher upgrade profile with id 29 found",
  "status": "error"
}
```

**Status:** Must be handled by checking `"status": "error"` in response body

---

### 6. id vs external_id Confusion

**Endpoints:**
- `POST /api/v2/infrastructure/publisherupgradeprofiles` (create)
- `GET /api/v2/infrastructure/publisherupgradeprofiles/{id}` (get)

**Issue:** The same resource has two different IDs, and the API inconsistently labels them:

**POST Response (Create):**
```json
{
  "data": {
    "id": 33,  // This is actually the external_id
    "name": "TF_Profile"
  }
}
```

**GET Response (Read the same resource using id=33):**
```json
{
  "data": {
    "id": 11,          // Internal database ID
    "external_id": 33,  // The ID from POST response
    "name": "TF_Profile"
  }
}
```

**Impact:**
- The `id` from POST is actually the `external_id`
- The `id` in GET responses is a different, internal identifier
- All URL paths use `external_id`, not `id`

**Status:** Known API design - Use `id` from POST as `external_id` for subsequent operations

---

### 7. Inconsistent Publisher Field Names

**Endpoints:**
- `POST /api/v2/infrastructure/publishers` (create)
- `GET /api/v2/infrastructure/publishers` (list)

**Issue:** Different endpoints use different JSON keys for publisher ID and name:

| Operation | ID Field | Name Field |
|-----------|----------|------------|
| Create (POST) | `id` | `name` |
| List (GET) | `publisher_id` | `publisher_name` |

**Example:**

**POST Response (Create):**
```json
{
  "data": {
    "id": 405,
    "name": "test-publisher",
    "status": "enabled"
  }
}
```

**GET Response (List):**
```json
{
  "data": {
    "publishers": [
      {
        "publisher_id": 405,
        "publisher_name": "test-publisher",
        "status": "enabled"
      }
    ]
  }
}
```

**Impact:**
- Code processing create vs list responses must handle different field names
- Cannot reuse the same data structure for both operations

**Status:** Known API inconsistency

---

## Terraform Provider Implications

These API issues have specific implications for the Terraform provider:

| Issue | Terraform Impact |
|-------|------------------|
| Bracket wrapping | State drift - `plan` may show changes when none exist |
| Inconsistent name keys | Type mapping challenges in Speakeasy-generated SDK |
| 200 OK with error body | `Read` must check body to detect deleted resources |
| `id` vs `external_id` | Resource ID schema and import functionality affected |
| Publisher field inconsistency | Mapping between create and read operations |
| Empty objects cause SQL error | Update operations fail without workaround |
| Protocol type/transport mismatch | State mapping requires hook transformation |
| Write-only fields | Perpetual drift in plan output for certain fields |
| Policy group response wrapper | ✅ Resolved - OAS now correctly defines wrapped response |
| NPA rules response wrapper | ✅ Resolved - OAS now correctly defines wrapped response |

### Implemented Mitigations

1. **OpenAPI Overlays** - Modify the OpenAPI spec to normalize field names
2. **SDK Hooks** - Use `internal/sdk/internal/hooks/` to transform requests/responses:
   - `hookMyAppAfterSuccess.go` - Maps `transport` → `type` in protocol responses
   - `hookPrivateAppRequest.go` - Strips empty `app_option`, `paths` from PUT requests
   - `hookErrorStatusResponse.go` - Handles 200 OK with `"status": "error"` responses
3. **Plan Modifiers** - Suppress diffs for computed fields via `x-speakeasy-param-suppress-computed-diff`
4. **Custom Read Logic** - Check response body status in resource Read functions
5. **Terraform Ignore** - Use `x-speakeasy-terraform-ignore` to exclude problematic fields from schema

---

## General Recommendations

1. **Always Use External IDs:** For all operations after creation, use the `id` (which is actually `external_id`) returned from the POST request.

2. **Check Response Body:** Don't rely solely on HTTP status codes. Always check for `"status": "error"` in the response body.

3. **Handle Both Name Keys:** When processing application data, handle both `name` and `app_name` keys.

4. **Account for Brackets:** Remember that application names may be automatically wrapped in brackets.

5. **Normalize Field Names:** Consider normalizing publisher data between create and list responses.

6. **Avoid Empty Objects in Updates:** Strip empty `{}` and `[]` values from PUT request payloads to avoid SQL serialization errors.

7. **Map Protocol Fields:** Transform `transport` to `type` when processing protocol data from responses.

8. **Accept Minor Drift:** Some fields like `allow_uri_bypass` will show perpetual drift in `terraform plan` because the API doesn't return them. This is cosmetic and doesn't affect functionality.

---

## SDK Hooks Reference

The provider implements several SDK hooks to work around API issues:

| Hook File | Type | Purpose |
|-----------|------|---------|
| `hookErrorStatusResponse.go` | AfterSuccess | Detects 200 OK responses with `"status": "error"` and converts to proper errors |
| `hookMyAppAfterSuccess.go` | AfterSuccess | Maps protocol `transport` → `type` in private app responses |
| `hookMyBulkAppAfterSuccess.go` | AfterSuccess | Handles bulk app response transformations |
| `hookMyPolicyAfterSuccess.go` | AfterSuccess | Handles policy response transformations (privateApps bracket trimming) |
| `hookMyPolicyBeforeRequest.go` | BeforeRequest | Transforms policy request payloads (bracket wrapping) |
| `hookPrivateAppRequest.go` | BeforeRequest | Strips empty `app_option`, `paths`, `uribypass_header_value` from PUT requests |
| `hookDebugRequest.go` | BeforeRequest | Debug hook for logging HTTP requests (disabled by default) |

These hooks are registered in `internal/sdk/internal/hooks/registration.go`.