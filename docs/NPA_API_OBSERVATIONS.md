# Netskope NPA API: Observed Behaviour, Terraform State, and Workarounds

This document describes the gap between the published OpenAPI specification for our NPA API and
what the API actually does.     

---

## Table of Contents

1. [Scope](#1-scope)
2. [Documented API vs Observed API](#2-documented-api-vs-observed-api)
   - [Steering — Private Apps](#steering--private-apps)
   - [Infrastructure — Publishers and Upgrade Profiles](#infrastructure--publishers-and-upgrade-profiles)
   - [Policy — NPA Rules](#policy--npa-rules)
   - [Device Classification](#device-classification)
3. [GET vs POST/PUT Data Model Differences](#3-get-vs-postput-data-model-differences)
4. [How Terraform Manages State](#4-how-terraform-manages-state)
5. [Quick Reference](#5-quick-reference)
6. [Guidance for Direct API Callers](#6-guidance-for-direct-api-callers)

---

## 1. Scope

This document covers the following NPA API surface areas:

| Area | Base path |
|------|-----------|
| Private apps | `/api/v2/steering/apps/private` |
| Publishers | `/api/v2/infrastructure/publishers` |
| Publisher upgrade profiles | `/api/v2/infrastructure/publisherupgradeprofiles` |
| NPA policy groups | `/api/v2/policy/npa/policygroups` |
| NPA policy rules | `/api/v2/policy/npa/rules` |
| Device classification rules | `/api/v2/deviceclassification/rules` |
| On-premises detection | `/api/v2/deviceclassification/client/onpremdetection` |
| DC steering mapping | `/api/v2/deviceclassification/steerings` |

Each issue below is tagged **Breaking**, **Cosmetic**, or **Informational**:

- **Breaking** — callers that follow the spec will receive errors or corrupt data.
- **Cosmetic** — the API works but returns data in a form that differs from what the spec implies,
  typically causing spurious diff detection in automation tools.
- **Informational** — the spec is incomplete or silent; callers need to know this to use the API
  correctly but will not get hard errors.

---

## 2. Documented API vs Observed API

### Steering — Private Apps

---

#### Issue 1 — App name is automatically wrapped in brackets `[Cosmetic]`

**Endpoint:** `POST /api/v2/steering/apps/private`

**What the spec says:** `app_name` is a plain string field.

**What actually happens:** The API silently wraps the value in square brackets on write. The name
you sent is not the name stored.

```
# Request
POST /api/v2/steering/apps/private
{
  "app_name": "CorpWiki"
}

# Stored value (visible in all subsequent GET responses)
{
  "app_name": "[CorpWiki]"
}
```

**Impact on callers:**
- Any lookup by name must use the bracketed form `[CorpWiki]`, not the form you sent.
- Comparing the configured name against the retrieved name will always fail without stripping
  brackets first.
- The Terraform provider strips brackets on every read so the state always stores the un-bracketed
  form, and re-adds them on write.

---

#### Issue 2 — `name` vs `app_name`: different keys for the same field `[Breaking]`

**Endpoints:**
- `GET /api/v2/steering/apps/private` (list)
- `GET /api/v2/steering/apps/private/{id}` (single)

**What the spec says:** The application name is a consistent field.

**What actually happens:** The field name differs depending on which endpoint you call.

```
# GET /api/v2/steering/apps/private/{id}  — single app
{
  "data": {
    "name": "CorpWiki",
    ...
  }
}

# GET /api/v2/steering/apps/private  — list
{
  "data": {
    "private_apps": [
      {
        "app_name": "CorpWiki",
        ...
      }
    ]
  }
}
```

**Impact on callers:** Code that processes both endpoints cannot use the same field name. Any
deserialization that maps a fixed key will silently get null on one of the two endpoints.

---

#### Issue 3 — Deleted resource returns `200 OK` with an error body instead of `404` `[Breaking]`

**Endpoints:** `GET /api/v2/steering/apps/private/{id}`,
`GET /api/v2/infrastructure/publisherupgradeprofiles/{id}`

**What the spec says:** A request for a non-existent resource returns `HTTP 404 Not Found`.

**What actually happens:** The API returns `HTTP 200 OK` with an error payload.

```
# Request for a deleted app
GET /api/v2/steering/apps/private/233

# Actual response
HTTP/1.1 200 OK
{
  "status": "error",
  "message": "No private app with id '233' is found."
}
```

**Impact on callers:** You cannot use the HTTP status code to determine whether a resource exists.
You must inspect the response body and check for `"status": "error"`. The same pattern applies to
the publisher upgrade profile endpoint.

---

#### Issue 4 — `app_id` vs `id`: the identifier field name differs between create and list `[Breaking]`

**Endpoints:**
- `POST /api/v2/steering/apps/private` (create — response uses `id`)
- `GET /api/v2/steering/apps/private` (list — uses `app_id`)
- `GET /api/v2/steering/apps/private/{id}` (single — uses `app_id`)

**What the spec says:** The identifier field is `id` in the create response.

**What actually happens:** The list and single-get endpoints return the identifier as `app_id`.
The create response calls it `id`. Both refer to the same logical resource identifier.

```
# POST response
{
  "data": {
    "id": 525,
    "app_name": "[CorpWiki]",
    ...
  }
}

# GET /api/v2/steering/apps/private (list)
{
  "data": {
    "private_apps": [
      {
        "app_id": 525,
        "app_name": "[CorpWiki]",
        ...
      }
    ]
  }
}
```

**Impact on callers:** If you try to look up an app from the list using the `id` returned by
create, you will find `app_id` — a differently named field with the same value. Automation that
stores the create response `id` and then looks for it in list responses must account for the rename.

---

#### Issue 5 — Protocol field is `type` on write, `transport` on read `[Breaking]`

**Endpoints:**
- `POST /api/v2/steering/apps/private` (create)
- `PUT /api/v2/steering/apps/private/{id}` (update)
- `GET /api/v2/steering/apps/private/{id}` (read)

**What the spec says:** Protocol type is specified using `type`.

**What actually happens:** Write operations accept `type`; read operations return `transport`. The
API stores them as the same value but exposes them under different names depending on the operation.

```
# POST request
{
  "protocols": [
    {"type": "tcp", "port": "443"}
  ]
}

# GET response (same resource)
{
  "protocols": [
    {
      "transport": "tcp",
      "port": "443",
      "id": 1024,
      "created_at": "2024-01-15T10:00:00Z",
      ...
    }
  ]
}
```

**Impact on callers:** You cannot use the same data structure for both reading and writing. A field
named `type` that you set on create will come back as `transport` on the next read. Any comparison
between what you configured and what the API returns will show a false mismatch.

---

#### Issue 6 — Empty arrays on PUT cause a backend serialization error `[Breaking]`

**Endpoint:** `PUT /api/v2/steering/apps/private/{id}`

**What the spec says:** `app_option`, `paths`, and `bypass_uris` are optional fields that can be
omitted or set to empty values.

**What actually happens:** Including these fields with empty values (`{}` or `[]`) triggers a
backend Python serialization bug that returns an HTTP 500 with this error body:

```
{
  "message": "(builtins.TypeError) Object of type bytes is not JSON serializable\n[SQL: UPDATE private_apps SET app_option=%(app_option)s, paths=%(paths)s ...]\n[parameters: [{'app_option': b'{}', 'paths': b'[]', ...}]]"
}
```

**Affected fields and their trigger values:**

| Field | Triggers error when |
|-------|---------------------|
| `app_option` | Sent as `{}` |
| `paths` | Sent as `[]` |
| `bypass_uris` | Sent as `[]` |
| `uribypass_header_value` | Sent as `null` or `""` |

**Impact on callers:** Any update to a private app that includes these fields with empty values
will fail with a 500. The workaround is to omit these fields entirely from the request body when
they have no value. Do not send `"paths": []` — omit the key altogether.

---

#### Issue 7 — Protocols returned in non-deterministic order `[Cosmetic]`

**Endpoint:** `GET /api/v2/steering/apps/private/{id}`

**What happens:** When an app has multiple protocols, the API returns them in an undocumented
internal sort order that is not the order you specified them on create.

The actual sort order applied by the backend is:
1. Protocol type alphabetically (`tcp` before `udp`)
2. Port number ascending within each type

So `[{tcp:443}, {tcp:22}, {udp:53}]` comes back as `[{tcp:22}, {tcp:443}, {udp:53}]`.

**Impact on callers:** Any tool that compares configured order against returned order will see a
false change. The Terraform provider normalises the order on every read to match what the API
will return, so `terraform plan` shows no diff regardless of the order you specified in config.

---

#### Issue 8 — Write-only fields: accepted on create/update, not returned on read `[Cosmetic]`

**Endpoint:** `GET /api/v2/steering/apps/private/{id}`

**Affected fields:**

| Field | Behaviour |
|-------|-----------|
| `allow_uri_bypass` | Accepted on write, returns `false` regardless of value set |
| `app_option` | Accepted on write (clientless apps only), returns `{}` always |
| `paths` | Accepted on write, returns `[]` always |

**What happens:** These fields are accepted and stored on create/update but the GET response does
not reflect what was actually stored.

```
# You send on POST/PUT:
{
  "allow_uri_bypass": true,
  "paths": [{"path": "/api", "display_name": "API endpoint"}]
}

# GET response returns:
{
  "allow_uri_bypass": false,
  "paths": []
}
```

**Impact on callers:** You cannot verify these fields after writing. Any system that reads back
the resource and compares to the configured value will always see a difference.

---

### Infrastructure — Publishers and Upgrade Profiles

---

#### Issue 9 — Publisher field names differ between create and list `[Breaking]`

**Endpoints:**
- `POST /api/v2/infrastructure/publishers` (create response)
- `GET /api/v2/infrastructure/publishers` (list)

**What happens:** The publisher identifier and name fields use different JSON keys depending on
the operation.

```
# POST response (create)
{
  "data": {
    "id": 405,
    "name": "dc-pub-01",
    "status": "not registered"
  }
}

# GET response (list)
{
  "data": {
    "publishers": [
      {
        "publisher_id": 405,
        "publisher_name": "dc-pub-01",
        "status": "connected"
      }
    ]
  }
}
```

**Impact on callers:** `id` from create is the same value as `publisher_id` in list responses.
You cannot use a single field mapping across both operations. Code that stores the create response
and later searches the list by field name will find nothing.

---

#### Issue 10 — Publisher upgrade profiles: `id` from create is `external_id` on all subsequent operations `[Breaking]`

**Endpoints:**
- `POST /api/v2/infrastructure/publisherupgradeprofiles` (create)
- `GET /api/v2/infrastructure/publisherupgradeprofiles/{id}` (read)

**What happens:** The create response returns a field called `id`. All subsequent
GET/PUT/DELETE operations use that value as the URL path parameter, but the GET response
body reveals that the value you stored as `id` is actually the resource's `external_id`.
The `id` field in the GET response is a different, internal database identifier.

```
# POST response
{
  "data": {
    "id": 33,          ← store this
    "name": "Prod Upgrade Profile"
  }
}

# GET /api/v2/infrastructure/publisherupgradeprofiles/33
{
  "data": {
    "id": 11,          ← internal DB id, different value
    "external_id": 33, ← the value from POST response
    "name": "Prod Upgrade Profile"
  }
}
```

**Impact on callers:** Always use the `id` from the POST response for URL path parameters in
subsequent calls. Never use the `id` field from GET responses for URL construction — use
`external_id` instead. The two values are not interchangeable.

---

#### Issue 11 — `connected_apps` has a different shape in list vs get-by-ID `[Breaking]`

**Endpoints:**
- `GET /api/v2/infrastructure/publishers` (list)
- `GET /api/v2/infrastructure/publishers/{id}` (single)

**What happens:** The `connected_apps` field is structurally different between the two endpoints.

```
# GET list response
{
  "data": {
    "publishers": [
      {
        "publisher_id": 6,
        "connected_apps": ["[CloudExchange]", "[WebServer]"]
      }
    ]
  }
}

# GET single publisher response
{
  "data": {
    "publisher_id": 6,
    "connected_apps": [
      {
        "name": "[CloudExchange]",
        "host": "192.168.1.10",
        "access_method": "client",
        "last_connected": null
      }
    ]
  }
}
```

**Impact on callers:** The list endpoint returns strings; the single-publisher endpoint returns
objects. A client that deserialises both responses into the same type will fail on one of the
two. This is a particularly insidious bug because publishers with zero connected apps return
an empty array `[]` on both endpoints — the divergence is invisible until you have at least
one connected app.

---

### Policy — NPA Rules

---

#### Issue 12 — `group_id` is accepted on write but never returned on read `[Informational]`

**Endpoints:**
- `POST /api/v2/policy/npa/rules` (create)
- `PATCH /api/v2/policy/npa/rules/{id}` (update)
- `GET /api/v2/policy/npa/rules/{id}` (read)

**What happens:** `group_id` (which assigns the rule to a policy group) is required for creating
and updating rules and is correctly stored by the API, but it is never included in GET responses.

```
# POST request
{
  "rule_name": "Allow-Engineering",
  "group_id": "873",
  "enabled": "1",
  "rule_data": { ... }
}

# GET response for the same rule
{
  "data": {
    "rule_id": "4",
    "rule_name": "Allow-Engineering",
    "enabled": "1",
    "rule_data": { ... }
    // group_id is not present
  }
}
```

**Impact on callers:** You cannot verify which policy group a rule belongs to by reading the rule.
You can only determine group membership from the policy group's own endpoints. Any tool that reads
a rule and tries to reconstruct its full configuration from the GET response will have an
incomplete picture.

---

#### Issue 13 — Block rule `template`: display name on write, filename on read `[Breaking]`

**Endpoint:** `POST /api/v2/policy/npa/rules`, `GET /api/v2/policy/npa/rules/{id}`

**What happens:** The `match_criteria_action.template` field (used when `action_name` is `block`)
uses different value formats depending on the operation direction:

- **Write (POST/PATCH):** the template's **display name** as shown in the UI (e.g. `"Default Template"`)
- **Read (GET):** the template's **filename** on the backend server (e.g. `"block_page.html"`)

```
# POST request — use the display name
{
  "rule_data": {
    "match_criteria_action": {
      "action_name": "block",
      "template": "Default Template"
    }
  }
}

# GET response for the same rule
{
  "rule_data": {
    "match_criteria_action": {
      "action_name": "block",
      "template": "block_page.html"
    }
  }
}
```

Sending a filename (`.html` suffix) on create or update returns:
```
{"message": "Undefined template: block_page.html", "status": "error"}
```

**Additional constraint:** The template list endpoint (`GET /api/v2/templates/usernotifications`)
requires a UI session token and returns `"Permission Error"` for API tokens. There is no
programmatic way to enumerate valid display names via the API — you must look them up in the
Netskope UI under Settings → Notifications → User Notifications.

**Impact on callers:** Always write the display name; expect to read back a filename. Any round-trip
comparison between the configured value and the stored value will show a mismatch because the
two representations are different strings for the same template.

---

#### Issue 14 — `device_classification_id`: string array on read, integer array required on write `[Breaking]`

**Endpoints:**
- `GET /api/v2/policy/npa/rules/{id}` (read — returns strings)
- `POST/PATCH /api/v2/policy/npa/rules` (write — requires integers)

**What happens:** The same field is typed differently depending on the direction of the call.

```
# GET response
{
  "rule_data": {
    "device_classification_id": ["5871", "5872"]   ← strings
  }
}

# POST/PATCH request — must send integers, not strings
{
  "rule_data": {
    "device_classification_id": [5871, 5872]        ← integers
  }
}
```

Sending strings on a write operation returns an HTTP 400 with a type validation error.

**Impact on callers:** You must convert the IDs from strings to integers before using them in a
write request. The OAS documents the field as `integer[]` in the request schema and the same in
the response schema, but the actual GET response sends strings — making the spec wrong for the
read path.

---

#### Issue 15 — Rule ordering has eventual consistency; there is no atomic reorder endpoint `[Informational]`

**Endpoint:** `PATCH /api/v2/policy/npa/rules/{id}` with `{"rule_order": {"order": "top"}}`

**What happens:** NPA rules have an ordered evaluation sequence. There is no endpoint that accepts
a complete desired order in a single call. Each rule must be repositioned individually using a
PATCH with a relative position instruction (`top`, `bottom`, `before`, `after`). The API applies
position changes asynchronously — a position change committed via PATCH may not be reflected
in the next immediate GET.

**Impact on callers:** Any automation that reorders rules must:
1. Apply positions sequentially, not in parallel
2. Wait a short time after the final PATCH before verifying order
3. Retry if the verification GET does not reflect the expected order

The Terraform provider works around this by processing rules in reverse desired order (each PATCH
to `top` builds the list from the back), sleeping 2 seconds after completion, and verifying up to
3 times with increasing delays.

---

#### Issue 16 — `periodic_reauth` rule template: display name stored verbatim; filenames rejected `[Breaking]`

**Endpoint:** `POST`/`PATCH`/`GET /api/v2/policy/npa/rules`

**What the spec says:** `rule_data.match_criteria_action.template` is a string.

**What actually happens (verified 2026-10-21 on two tenants, one where the templates endpoint returns 403 and one where it returns 200 in
`docs/api-audit/npa-rules-template/round2-*`):**

| Action | Display name on write | Filename on write | GET returns |
|---|---|---|---|
| `block` | accepted, resolved to the filename | `Undefined template: block_page.html` | filename |
| `periodic_reauth` | accepted, **stored verbatim** | `Undefined template: 3.html` | the display name string |

So the two actions are asymmetric in opposite directions, and for `periodic_reauth` the display
name is the **only** value the public API accepts. The Netskope UI, however, stores
`periodic_reauth` templates by filename (a UI-created rule reads back as `"3.html"`), so UI- and
API-created rules are stored differently.

**Consequences for callers:**
- Always send the display name. Never send a filename for either action.
- A `periodic_reauth` rule created through the API may show no template selected in the UI rule
  editor (the UI matches by filename). Whether enforcement differs has not been confirmed.
- `GET /api/v2/templates/usernotifications` (the name→filename catalogue) is **not needed** for
  either action and returns 403 to all API tokens on some tenants; the provider does not use it.

**Provider handling:** the display name is passed through unchanged for all actions. For
`block`, the filename returned on read is reconciled with the configured name via Terraform
private state. For `periodic_reauth`, state equals config.

#### Issue 17 — `rule_order` is silently ignored on create; effective only via PATCH `[Informational]`

**Endpoints:** `POST /api/v2/policy/npa/rules`, `PATCH /api/v2/policy/npa/rules/{id}`

**What happens:** Including `rule_order` in a POST create request does not set the rule's position.
The API silently ignores the field and creates the rule at the bottom of the evaluation list.
Ordering only takes effect when issued as a subsequent PATCH request.

```
# POST request — rule_order included but ignored
{
  "rule_name": "Allow-Engineering",
  "rule_order": {"order": "top"},
  "rule_data": { ... }
}

# Rule is created at the bottom of the list regardless.
# A subsequent PATCH is required to move it:
PATCH /api/v2/policy/npa/rules/{id}
{"rule_order": {"order": "top"}}
```

**Impact on callers:** Any automation that expects a rule to be at a specific position immediately
after create will observe the wrong order. Create and reorder must be separate API calls. The
Terraform provider sends a PATCH immediately after create when a non-default position is requested.

Additionally, the Terraform provider strips `rule_order` from PATCH (update) request bodies to
prevent accidental reordering — a PATCH that includes `rule_order` would move the rule, which is
not the intent of a configuration update that changes rule criteria.

---

### Device Classification

---

#### Issue 18 — Device classification rule create: body must be an array; 201 returns empty body `[Breaking]`

**Endpoint:** `POST /api/v2/deviceclassification/rules`

**What the spec says:** A standard REST create — POST a rule object, receive the created rule.

**What actually happens:** The API is a batch endpoint that expects the body to be a JSON array
even when creating a single rule. It returns HTTP 201 with an **empty body**, not the created
object.

```
# Correct request — single rule wrapped in an array
POST /api/v2/deviceclassification/rules
[
  {
    "name": "Corporate Laptop Check",
    "label": "corp-managed",
    "rule_data": { ... }
  }
]

# Response
HTTP/1.1 201 Created
(empty body)
```

**Impact on callers:** You cannot retrieve the new rule's ID from the create response because
there is no response body. You must perform a follow-up GET, filtering by the rule name or label
you sent. The Terraform provider does this automatically using an exponential backoff retry
(1s → 2s → 4s → 8s → 16s) because the rule may not be immediately visible after creation
(eventual consistency).

---

#### Issue 19 — On-premises detection create: non-standard 201 response `[Informational]`

**Endpoint:** `POST /api/v2/deviceclassification/client/onpremdetection`

**What happens:** Like device classification rules, the on-prem detection endpoint accepts an
array body. Its 201 response is not empty but uses a non-standard envelope:

```
HTTP/1.1 201 Created
{
  "status": true,
  "data": [42]       ← array of created IDs
}
```

**Impact on callers:** The created ID is inside a `data` array, not at a top-level `id` field as
most POST endpoints return. Parse `data[0]` for the ID of the newly created object.

---

## 3. GET vs POST/PUT Data Model Differences

This section shows the field-level differences for each resource in a side-by-side format. The
intention is to make it easy to write a client that handles both read and write paths correctly.

### Private App

| Field | POST body (write) | GET body (read, single) | GET body (read, list) |
|-------|-------------------|-------------------------|-----------------------|
| App name | `app_name` (string) | `name` (string, with `[]`) | `app_name` (string, with `[]`) |
| Resource ID | _(not in request body)_ | `id` (int) — but also `app_id` | `app_id` (int), `id` absent or 0 |
| Protocol type | `protocols[].type` | `protocols[].transport` | `protocols[].transport` |
| Protocol port | `protocols[].port` (string) | `protocols[].port` (string) | `protocols[].port` (string) |
| Publisher assignment ID | `service_publisher_assignments[].publisher_id` (string) | `service_publisher_assignments[].publisher_id` (int or float) | same as single |
| `allow_uri_bypass` | accepted | always `false` in response | always `false` in response |
| `paths` | accepted | always `[]` | always `[]` |
| `app_option` | accepted (clientless apps) | always `{}` | always `{}` |

**Practical rules:**
- When writing, use `app_name` and `type` (in protocols).
- When reading single, use `name` and `transport`.
- When reading from list, use `app_name` and `transport`.
- Strip `[]` brackets from name fields before comparing or storing.
- Treat `app_id` (from list/single reads) as the canonical resource ID; map to `id` when needed for URL paths.

---

### Publisher

| Field | POST create response | GET list |
|-------|----------------------|----------|
| Publisher ID | `id` (int) | `publisher_id` (int) |
| Publisher name | `name` (string) | `publisher_name` (string) |
| Connected apps | not present | `connected_apps` (string array) |

| Field | GET single |
|-------|------------|
| Publisher ID | `publisher_id` (int) — note: field name differs from list but same value |
| Connected apps | `connected_apps` (object array: `{name, host, access_method, last_connected}`) |

---

### Publisher Upgrade Profile

| Field | POST create response | GET single |
|-------|----------------------|------------|
| ID for URL paths | `id` (int) — **use this in all subsequent URLs** | `external_id` (int) |
| Internal DB id | _(same field, but it is actually external_id)_ | `id` (int, different value) |
| Name | `name` | `name` |

---

### NPA Policy Rule

| Field | Write body (POST, PUT or PATCH) | GET response (read) |
|-------|-----------------------|---------------------|
| `group_id` | accepted (required for create) | **not present** |
| `device_classification_id` | `[5871, 5872]` (integer array) | `["5871", "5872"]` (string array) |
| `match_criteria_action.template` (block) | `"Default Template"` (display name) | `"block_page.html"` (filename) |
| `match_criteria_action.template` (periodic_reauth) | `"Default Template"` (display name) | `"Default Template"` (display name — verbatim, no translation) |
| `privateApps` items | `"CorpWiki"` (plain string) | `"[CorpWiki]"` (bracket-wrapped) |
| `enabled` | `"1"` or `"0"` (string) | `"1"` or `"0"` (string — consistent) |

---

### Device Classification Rule

| Aspect | Create (POST) | Read (GET) |
|--------|---------------|------------|
| Request body shape | JSON array (batch endpoint) | N/A |
| Response body | Empty (201 with no body) | Array of rule objects |
| Rule ID | Not returned | Present in each rule object |
| Discovery after create | Must GET by label/name | Filter on `label` field |

---

## 4. How Terraform Manages State

This section explains why API inconsistencies create specific problems for Terraform, and how the
provider resolves them. You do not need to know Terraform's internal implementation — the concepts
are explained in API terms.

### What Terraform state is

Terraform keeps a local record of every resource it manages, called the **state file**. After any
`apply` operation, the state file holds the last-known configuration of each resource as returned
by the API.

On every `terraform plan`, Terraform:
1. Calls the API (`GET`) to read the current live state of the resource
2. Compares the live state to the state file and to your configuration
3. Reports any difference as a proposed change

For this to work correctly, the value the API returns on GET must be the same logical value you
configured. If the API stores something different from what you sent, or returns it in a different
form, Terraform will always see a difference and propose an update — even when nothing has
actually changed. This is called **perpetual drift**.

### Why the data model matters

The provider translates between your Terraform configuration and the API. It uses typed data
structures: every field in the API has a specific type (`string`, `integer`, `array of string`,
etc.). When the API returns a value in a different type than declared (e.g. `"5871"` as a string
when the field was declared as integer), the provider cannot deserialise the response — or will
silently misread the value.

Four categories of mismatch cause real problems:

| Category | What breaks |
|----------|-------------|
| Field renamed between read and write | Terraform always sees the field as missing on read |
| Type changes between read and write | Deserialisation fails or values compare as unequal |
| Field present on write, absent on read | Terraform overwrites state with null on every refresh |
| Non-deterministic ordering | Terraform sees item-by-item differences on every refresh |

The following subsections show each category with a concrete example.

---

### Example A — Field rename: `type` → `transport` (protocols)

You configure a private app with a TCP protocol on port 443. The provider sends this to the API:

```json
{
  "protocols": [{"type": "tcp", "port": "443"}]
}
```

On the next `terraform plan`, the provider reads the app back:

```json
{
  "protocols": [{"transport": "tcp", "port": "443", "id": 1024}]
}
```

The field `type` is absent. The field `transport` is present but is not what the provider is
looking for. Without intervention, Terraform sees:

```
~ protocols = [
    ~ {
        ~ type      = "tcp" -> null     # was set, now missing
          transport = "tcp"             # present but not tracked
      }
  ]
```

Every `terraform plan` would propose re-adding `type`, which would send the same PUT again,
which would not actually change anything — but would clutter the audit trail.

**How the provider resolves it:** An AfterSuccess hook intercepts every GET response and copies
`transport` into `type` before Terraform reads the data. Terraform only ever sees `type = "tcp"`,
so no drift is detected.

```
API returns:    {"transport": "tcp", "port": "443", "id": 1024}
After hook:     {"type": "tcp", "transport": "tcp", "port": "443", "id": 1024}
Terraform sees: type = "tcp"  ✓ matches config
```

---

### Example B — Type mismatch: `device_classification_id` (string vs integer)

The API returns device classification IDs as strings in GET responses but requires integers in
POST/PATCH requests. The OAS documents both as integers, which is wrong for the read path.

If the provider's data structure declares the field as `integer`, reading `"5871"` (a JSON string)
will fail to deserialise — or the field will be read as zero — causing create-then-read to fail
silently.

**How the provider resolves it:**

Step 1 — Declare the field as `string` internally, matching what GET actually returns:
```
GET returns:    ["5871", "5872"]   ← strings, stored in state as strings
```

Step 2 — A BeforeRequest hook converts strings to integers just before any POST/PATCH:
```
Config/state:   device_classification_id = ["5871", "5872"]
Before hook:    device_classification_id → [5871, 5872]  (integers)
API receives:   [5871, 5872]  ✓
```

The user writes strings in their config; the API sends and receives correctly; Terraform state
stays as strings so the next plan sees no difference.

---

### Example C — Write-only field: `group_id` on NPA rules

The `group_id` field is required when creating a rule but is never returned by GET. Without
special handling, this is what happens:

1. User configures `group_id = "873"` → provider sends it to the API on create ✓
2. Provider reads the rule back via GET → `group_id` is absent from response
3. Provider updates state with GET response → `group_id` is now `null` in state
4. Next `terraform plan` shows `group_id: "873" → null` as a proposed change
5. User runs apply → provider sends PATCH with `group_id = "873"` again → back to step 2

**How the provider resolves it:** The field is excluded from the GET response schema entirely.
When Terraform reads the GET response, it ignores `group_id` (because the schema says it is not
there). The value `"873"` that you set remains in the state file untouched. On each refresh,
Terraform compares `"873"` (from state) against `"873"` (from state, not overwritten) → no diff.

---

### Example D — Non-deterministic ordering: publishers on a private app

A private app has two publishers assigned. On two consecutive GET requests, the API returns them
in different orders:

```
# GET request 1
"service_publisher_assignments": [
  {"publisher_id": 2, "publisher_name": "dc-pub-02"},
  {"publisher_id": 1, "publisher_name": "dc-pub-01"}
]

# GET request 2
"service_publisher_assignments": [
  {"publisher_id": 1, "publisher_name": "dc-pub-01"},
  {"publisher_id": 2, "publisher_name": "dc-pub-02"}
]
```

Terraform represents lists as ordered sequences. If request 1 result is in state and request 2
result is the live read, Terraform sees publisher 1 and publisher 2 swapped — and proposes an
update to swap them back. After the update, the next plan swaps them again. Every plan shows a
change.

**How the provider resolves it:** An AfterSuccess hook sorts `service_publisher_assignments` by
`publisher_id` numerically on every GET response before Terraform processes it. The order is
always deterministic regardless of what the API returns, so no drift is detected.

The same technique is applied to:
- `protocols` — sorted by type (alphabetically) then port (numerically)
- `tags` — sorted by `tag_id`
- `label_ids` — sorted alphabetically

---

### Summary of fix mechanisms

| Fix type | Where it lives | What it does |
|----------|---------------|--------------|
| AfterSuccess hook | Provider hook layer | Transforms GET responses before Terraform reads them: rename fields, sort arrays, trim brackets, copy IDs |
| BeforeRequest hook | Provider hook layer | Transforms write payloads before sending to API: convert types, strip empty fields, add array wrappers |
| OAS annotation: `x-speakeasy-terraform-ignore` | OpenAPI spec | Removes a field from the Terraform schema entirely — Terraform never reads or writes it |
| OAS annotation: `x-speakeasy-param-suppress-computed-diff` | OpenAPI spec | Tells Terraform "do not diff this field against the API response" — preserves config value in state |
| Schema `Computed` flag | OpenAPI spec | Marks fields that are server-assigned (e.g. `modify_time`) — Terraform does not compare them to config |

---

## 5. Quick Reference

| # | Issue | Endpoint | Severity | Status |
|---|-------|----------|----------|--------|
| 1 | App name wrapped in `[]` on write | `POST /steering/apps/private` | Cosmetic | Provider normalises on read |
| 2 | `name` vs `app_name` key | `GET /steering/apps/private` vs `/{id}` | Breaking | Provider maps both |
| 3 | 200 OK with error body instead of 404 | `GET /steering/apps/private/{id}`, `GET /infrastructure/publisherupgradeprofiles/{id}` | Breaking | Provider checks `status` in body |
| 4 | `app_id` vs `id` in list vs create | `POST` vs `GET /steering/apps/private` | Breaking | Provider copies `app_id` → `id` |
| 5 | Protocol `type` vs `transport` | `POST`/`GET /steering/apps/private` | Breaking | Provider maps `transport` → `type` on read |
| 6 | Empty `[]`/`{}` on PUT causes SQL error | `PUT /steering/apps/private/{id}` | Breaking | Provider strips empty fields before sending |
| 7 | Write-only fields (`allow_uri_bypass`, `paths`) | `GET /steering/apps/private/{id}` | Cosmetic | Provider ignores on read |
| 8 | Protocol order non-deterministic | `GET /steering/apps/private/{id}` | Cosmetic | Provider sorts on every read |
| 9 | Publisher `id`/`name` vs `publisher_id`/`publisher_name` | `POST` vs `GET /infrastructure/publishers` | Breaking | Provider maps field names |
| 10 | `id` from upgrade profile create is `external_id` in GET | `POST`/`GET /infrastructure/publisherupgradeprofiles` | Breaking | Provider uses POST `id` as path parameter |
| 11 | `connected_apps` is string array in list, object array in single | `GET /infrastructure/publishers` vs `/{id}` | Breaking | Provider ignores `connected_apps` on single read |
| 12 | `group_id` not returned by GET on rules | `GET /policy/npa/rules/{id}` | Informational | Provider preserves from state |
| 13 | Block rule template display name vs filename | `POST`/`GET /policy/npa/rules` | Breaking | Provider sends the display name unchanged and reconciles the GET filename via resource private state (0.4.13) |
| 14 | `device_classification_id` string in GET, integer in POST | `GET`/`POST /policy/npa/rules` | Breaking | Provider stores as string, converts to int on write |
| 15 | Rule ordering eventually consistent | `PATCH /policy/npa/rules/{id}` | Informational | Provider retries with verify loop |
| 16 | `periodic_reauth` template: display name stored verbatim (no filename translation) | `POST`/`GET /policy/npa/rules` | Informational | Provider passes through unchanged; no drift suppression needed |
| 17 | `rule_order` ignored on create; `rule_order` stripped from PATCH updates | `POST`/`PATCH /policy/npa/rules/{id}` | Informational | Provider sends PATCH after create; strips field on update |
| 18 | DC rule create: array body, empty 201 | `POST /deviceclassification/rules` | Breaking | Provider wraps body, does follow-up GET |
| 19 | On-prem detection create: non-standard 201 envelope | `POST /deviceclassification/client/onpremdetection` | Informational | Provider reads `data[0]` for ID |

---

## 6. Guidance for Direct API Callers

If you are calling the NPA API directly (outside the Terraform provider), apply these rules to
avoid the issues documented above.

### Always check the response body, not just the HTTP status

Do not treat `HTTP 200` as success. Always parse the response JSON and check for
`"status": "error"`. This applies to GET, DELETE, and PUT across private apps and upgrade
profiles.

```python
response = requests.get(url, headers=auth_headers)
data = response.json()
if data.get("status") == "error":
    raise APIError(data.get("message"))
```

### Store the right ID after create

| Resource | POST response field | Use for subsequent operations |
|----------|---------------------|-------------------------------|
| Private app | `id` | URL path parameter |
| Publisher | `id` | URL path parameter |
| Upgrade profile | `id` | URL path parameter (this is the `external_id`) |
| DC rule | not in response | GET by label, use `id` from GET result |
| On-prem detection | `data[0]` | URL path parameter |

### Handle field name differences across endpoints

Build your client with awareness that field names change between endpoints:

| Resource | Create response | List GET | Single GET |
|----------|----------------|----------|------------|
| Private app ID | `id` | `app_id` | `app_id` |
| Private app name | `app_name` | `app_name` | `name` |
| Publisher ID | `id` | `publisher_id` | `publisher_id` |
| Publisher name | `name` | `publisher_name` | `publisher_name` |

### Strip brackets from app names before storing or comparing

The API wraps app names on create. If you store the name you sent (`CorpWiki`) and compare it
against the retrieved name (`[CorpWiki]`), they will never match. Strip `[]` on read:

```python
name = retrieved_app["app_name"].strip("[]")
```

### Convert `device_classification_id` before writing

Read as strings; convert to integers before writing (POST/PATCH for rules):

```python
# IDs from GET (strings)
ids = rule_data["device_classification_id"]  # ["5871", "5872"]

# Convert for write (POST/PATCH)
request_body["rule_data"]["device_classification_id"] = [int(i) for i in ids]
```

### Omit empty arrays and objects from PUT payloads

Never send `"paths": []`, `"bypass_uris": []`, `"app_option": {}`, or
`"uribypass_header_value": null` on private app updates. Omit these keys entirely if they have
no value.

### Use display names for block rule templates

Check the Netskope UI (Settings → Notifications → User Notifications) for the display name of
the template you want. Use that string when creating or updating a block rule. Do not use the
`.html` filename that the GET response returns.

### For NPA rules: read `group_id` from the policy group, not the rule

Because `group_id` is write-only on rule endpoints, you cannot determine a rule's group by
reading the rule itself. Query the policy group endpoint (`GET /api/v2/policy/npa/policygroups`)
to enumerate which rules belong to which group.

### For rule reordering: apply sequentially with a verify step

Apply each PATCH individually in the desired reverse order, then verify the final order via GET.
Expect to retry at least once due to eventual consistency. Applying reorder PATCHes in parallel
will produce unpredictable results.
