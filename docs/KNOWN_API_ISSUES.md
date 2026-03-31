# Known API Issues and Workarounds

This document tracks known behavioral inconsistencies and quirks in the Netskope REST API that this Terraform provider works around or that developers should be aware of.

## Table of Contents

- [Steering API Issues (1-3)](#steering-api-issues-1-3)
  - [1. App Name Automatic Encapsulation](#1-app-name-automatic-encapsulation)
  - [2. Inconsistent Name Key in Responses](#2-inconsistent-name-key-in-responses)
  - [3. Delete Verification Returns 200 OK with Error](#3-delete-verification-returns-200-ok-with-error)
- [Infrastructure API Issues (4-7)](#infrastructure-api-issues-4-7)
  - [4. Missing external_id on Profile Creation](#4-missing-external_id-on-profile-creation)
  - [5. GET on Deleted Resource Returns 200 OK](#5-get-on-deleted-resource-returns-200-ok)
  - [6. id vs external_id Confusion](#6-id-vs-external_id-confusion)
  - [7. Inconsistent Publisher Field Names](#7-inconsistent-publisher-field-names)
- [Steering API Issues (8-10)](#steering-api-issues-8-10)
  - [8. Empty Objects Cause SQL Serialization Error on Update](#8-empty-objects-cause-sql-serialization-error-on-update)
  - [9. Protocol Field Name Mismatch (type vs transport)](#9-protocol-field-name-mismatch-type-vs-transport)
  - [10. Write-Only Fields Not Returned in Response](#10-write-only-fields-not-returned-in-response)
- [Steering API Issues (11)](#steering-api-issues-11)
  - [11. Protocol Ordering Causes Terraform State Drift](#11-protocol-ordering-causes-terraform-state-drift)
- [Policy API Issues (12-13)](#policy-api-issues-12-13)
  - [12. NPA Rules `group_id` is Write-Only](#12-npa-rules-group_id-is-write-only)
  - [13. API Tokens Cannot Resolve User Notification Templates for Block Rules](#13-api-tokens-cannot-resolve-user-notification-templates-for-block-rules)
- [Terraform Provider Implications](#terraform-provider-implications)
- [General Recommendations](#general-recommendations)

---

## Steering API Issues (1-3)

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

## Infrastructure API Issues (4-7)

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

## Steering API Issues (8-10)

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

## Steering API Issues (11)

### 11. Protocol Ordering Causes Terraform State Drift

**Endpoint:** `GET /api/v2/steering/apps/private/{id}`

**Issue:** When a private app is configured with multiple protocols, the API returns the protocols in a potentially different order than they were specified during creation. The Terraform provider uses a list (ordered) for the `protocols` attribute, so any difference in ordering between the configuration and the API response causes Terraform to detect "drift" and propose changes on every `terraform plan`.

**Example Configuration:**

```hcl
resource "netskope_npa_private_app" "example" {
  private_app_name     = "Multi-Protocol App"
  private_app_hostname = "app.internal.local"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "22"
      protocol = "tcp"
    }
  ]
  # ...
}
```

**API Response (Reordered):**

```json
{
  "protocols": [
    {"port": "22", "transport": "tcp", "id": 123},
    {"port": "443", "transport": "tcp", "id": 124}
  ]
}
```

**Terraform Plan Output (Perpetual Drift):**

```
# netskope_npa_private_app.example will be updated in-place
~ resource "netskope_npa_private_app" "example" {
    ~ protocols = [
        ~ {
            ~ port     = "22" -> "443"
            ~ protocol = "tcp" -> "tcp"
          },
        ~ {
            ~ port     = "443" -> "22"
            ~ protocol = "tcp" -> "tcp"
          },
      ]
  }
```

**Root Cause:** The API backend sorts protocols internally using a two-level sort:
1. **Protocol type** (alphabetically: `tcp` before `udp`)
2. **Port number** (ascending within each protocol type)

This ordering is not documented and was discovered through testing. Since Terraform lists are ordered, any difference between configuration order and response order is detected as drift.

**Impact:**
- Every `terraform plan` shows changes even when no actual changes are needed
- Running `terraform apply` repeatedly may work but clutters the audit log
- Tests using `ExpectNonEmptyPlan: false` will fail for multi-protocol apps
- Users may be confused about why their infrastructure shows pending changes

**Workaround:** **Always specify protocols in the same order the API returns them:**
1. All TCP protocols first, sorted by port ascending
2. All UDP protocols second, sorted by port ascending

**Correct Configuration (No Drift):**

```hcl
# Single protocol type - just sort by port
protocols = [
  {
    port     = "22"       # Lower port first
    protocol = "tcp"
  },
  {
    port     = "443"      # Higher port second
    protocol = "tcp"
  }
]

# Mixed protocol types - TCP first (sorted), then UDP (sorted)
protocols = [
  {
    port     = "22"       # TCP ports first, sorted ascending
    protocol = "tcp"
  },
  {
    port     = "443"
    protocol = "tcp"
  },
  {
    port     = "53"       # UDP ports second, sorted ascending
    protocol = "udp"
  },
  {
    port     = "123"
    protocol = "udp"
  }
]
```

**Incorrect Configuration (Causes Drift):**

```hcl
# WRONG: UDP before TCP
protocols = [
  {
    port     = "53"
    protocol = "udp"      # UDP should come AFTER all TCP entries
  },
  {
    port     = "443"
    protocol = "tcp"
  }
]

# WRONG: Ports not sorted within protocol type
protocols = [
  {
    port     = "443"      # Should be 22 first
    protocol = "tcp"
  },
  {
    port     = "22"
    protocol = "tcp"
  }
]
```

**Status:** **Fixed in v0.4.0** — The AfterSuccess hooks (`hookMyAppAfterSuccess.go`, `hookMyBulkAppAfterSuccess.go`) now sort protocols by type (alphabetically) then port (numerically ascending) in all API responses. Users no longer need to specify protocols in a specific order.

---

## Policy API Issues (12-13)

### 12. NPA Rules `group_id` is Write-Only

**Endpoints:**
- `POST /api/v2/policy/npa/rules` (create)
- `GET /api/v2/policy/npa/rules/{id}` (read)

**Issue:** The `group_id` field is accepted in create and update requests but is **not returned** in the GET response. This makes it a write-only field.

**Example:**

**Request (Create):**
```json
{
  "rule_name": "my-rule",
  "group_id": "873",
  "enabled": "1",
  "rule_data": {...}
}
```

**Response (GET):**
```json
{
  "data": {
    "rule_id": "4",
    "rule_name": "my-rule",
    "enabled": "1",
    "rule_data": {...}
    // Note: group_id is NOT present in the response
  },
  "status": "success"
}
```

**Impact:**
- Terraform cannot verify `group_id` after creation since it's not in the GET response
- Import operations cannot recover the `group_id` value
- If included in the response schema, RefreshFrom would set it to null, causing perpetual drift

**Workaround:** The `group_id` field has been removed from the response schema (`npa_policy_response_item`) in the OpenAPI spec. This preserves the user-configured value in Terraform state without being overwritten by the null response. The field is excluded from import state verification (`ImportStateVerifyIgnore`).

**Status:** API limitation - Workaround implemented in OAS (0.3.3)

---

### 13. API Tokens Cannot Resolve User Notification Templates for Block Rules

**Endpoint:** `POST /api/v2/policy/npa/rules`

**Issue:** When creating an NPA rule with `action_name: "block"`, the API requires a `template` field in `match_criteria_action` containing the notification page **file_name** (e.g. `1.html`, `block_page.html`). However, API tokens cannot resolve any template — all requests return `"Undefined template"`, even for the default public template `block_page.html`. The same operation succeeds when performed via the UI.

**Example:**

```json
// Request
{
  "rule_data": {
    "match_criteria_action": {
      "action_name": "block",
      "emit_alert": true,
      "template": "1.html"
    }
  }
}

// Response (200 OK with error)
{"message": "Undefined template: 1.html", "status": "error"}
```

**Additional Context:**
- The `template` field expects the `file_name` from user notification pages (e.g. `1.html`), **not** the `template_name` display name (e.g. "tf_test_template")
- The `/api/v2/templates/usernotifications` endpoint returns `"Permission Error"` for API tokens, so template file_names cannot be looked up programmatically
- The Terraform provider schema correctly includes `emit_alert` and `template` fields in `match_criteria_action` (added in v0.4.0), but block rules cannot be created via API until this permission issue is resolved
- Existing block rules created via UI can be read and imported by the provider

**Impact:** Block rules cannot be created via API or Terraform. Only allow rules can be automated.

**Workaround:** Create block rules manually via the UI. Use Terraform import to bring existing block rules under management.

**Additional Finding:** The `template` field has a **name/filename mismatch**:
- **Create/Update** requires the template **display name** (e.g. `"Default Template"`)
- **GET response** returns the template **file name** (e.g. `"block_page.html"`)

This causes perpetual drift in Terraform — the config specifies the display name, the state refreshes to the filename, and every subsequent plan shows a change. Using the filename on create returns "Undefined template".

```
# Perpetual drift example:
~ match_criteria_action = {
    ~ template = "block_page.html" -> "Default Template"
  }
```

**Workaround:** Use `lifecycle { ignore_changes }` to suppress the drift on `match_criteria_action`:

```hcl
resource "netskope_npa_rules" "block_rule" {
  rule_name = "my-block-rule"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"
    match_criteria_action = {
      action_name = "block"
      template    = "Default Template"
      emit_alert  = true
    }
    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  lifecycle {
    ignore_changes = [rule_data]
  }
}
```

**Status:** API inconsistency — Jira ticket raised. Block rules can be created but will show perpetual drift on the `template` field until the API returns a consistent value. Use `lifecycle { ignore_changes = [rule_data] }` as a workaround.

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
| Protocol ordering | **Fixed in v0.4.0** — hooks sort protocols automatically |
| NPA rules `group_id` write-only | `group_id` removed from response schema to preserve state (0.3.3) |
| Block rule template name/filename mismatch | Create requires display name, GET returns filename — causes perpetual drift |

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

9. **Order Protocols Correctly:** When defining multiple protocols for a private app, list them in the order the API returns them: TCP protocols first (sorted by port ascending), then UDP protocols (sorted by port ascending). Example: TCP:22, TCP:443, UDP:53, UDP:123.

---

## SDK Hooks Reference

The provider implements several SDK hooks to work around API issues:

| Hook File | Type | Purpose |
|-----------|------|---------|
| `hookErrorStatusResponse.go` | AfterSuccess | Detects 200 OK responses with `"status": "error"` and converts to proper errors |
| `hookMyAppAfterSuccess.go` | AfterSuccess | Maps protocol `transport` → `type`, sorts protocols/publishers/tags, populates `label_ids` from `labels` |
| `hookMyBulkAppAfterSuccess.go` | AfterSuccess | Same transformations as above for bulk (list) responses |
| `hookMyPolicyAfterSuccess.go` | AfterSuccess | Handles policy response transformations (privateApps bracket trimming) |
| `hookMyPolicyBeforeRequest.go` | BeforeRequest | Transforms policy request payloads (bracket wrapping, preserves `emit_alert`/`template` in `match_criteria_action`) |
| `hookPrivateAppRequest.go` | BeforeRequest | Strips empty `app_option`, `paths`, `uribypass_header_value` from PUT requests |
| `hookDebugRequest.go` | BeforeRequest | Debug hook for logging HTTP requests (disabled by default) |

These hooks are registered in `internal/sdk/internal/hooks/registration.go`.
