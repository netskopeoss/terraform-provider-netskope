---
title: BUG-022 — periodic_reauth rules lose template on create/update; template drift after read
status: Fixed
github_issue: 118
---

## Summary

`netskope_npa_rules` resources with `action_name = "periodic_reauth"` had three related problems:

1. **Template not applied on create or update**
2. **Template drift on every read** (`.html` filenames appearing in state)
3. **`terraform import` failure** (`ImportStateVerify` mismatch on the `template` field)

All three stem from **asymmetric REST API behaviour** between `block` and `periodic_reauth` rules
for the `template` field.

## REST API Behaviour — Block vs Periodic_reauth

The two action types handle the `template` field differently across the NPA rules and templates APIs.

### Template name formats

The Netskope UI displays templates by a human-readable **display name** (e.g. `"Default Block Page"`).
Internally, each template is stored as an **`.html` filename** (e.g. `"1.html"`).

### POST/PUT behaviour (write path)

| | Block | Periodic_reauth |
|---|---|---|
| Accepts display name? | Yes — server translates to `.html` | No — value stored verbatim |
| Accepts `.html` filename? | No — rejected: "Undefined template: *.html" | Yes — stored and applied |
| Sends display name → result | Template applied correctly | Template ignored silently |
| Sends `.html` filename → result | Request rejected | Template applied correctly |

Block rules and periodic_reauth rules require **opposite** formats on write:
- Block: send the display name; the API translates it to `.html` server-side
- Periodic_reauth: send the `.html` filename; the API stores it verbatim

### GET behaviour (read path)

Both action types return the `.html` filename in GET responses. The display name is never
returned in rule responses.

```
GET /api/v2/policy/npa/rules/{id}
→ rule_data.match_criteria_action.template = "1.html"   (always a filename, regardless of action type)
```

### Templates lookup API

Display name ↔ `.html` filename mappings are available from a separate endpoint:

```
GET /api/v2/templates/usernotifications
→ [{ "file_name": "1.html", "name": "Default Block Page", "action_type": "block" }, ...]
```

Note: templates are created with `action_type` set at creation time in the UI (typically `"block"`).
A template used in a periodic_reauth rule retains its original `action_type` in this response —
it does not change to `"periodic_reauth"`. This means filename → display name lookups keyed on
`action_type` can miss when the same template is used across action types.

## Why the Lookup Cache Is Required

Users specify `template` by display name in Terraform config (consistent with the UI).
The provider must translate in both directions before sending to or after receiving from the API:

**Write path (BeforeRequest):**
- Block: display name passes through unchanged — the API translates it
- Periodic_reauth: display name must be translated to `.html` filename before POST/PUT

**Read path (AfterSuccess):**
- Both: GET returns `.html` filename; provider translates back to display name for state

Because the templates lookup API is the only source of these mappings, the provider caches the
full response from `GET /api/v2/templates/usernotifications` at the start of each session. The
cache stores bidirectional mappings:

- `filename → display name` (for the read path)
- `display name → filename` (for the periodic_reauth write path)

The cache lookup falls back to a cross-`action_type` search when the action_type-specific key
is not found — necessary because a template created for `"block"` in the UI retains
`action_type = "block"` even when used in a periodic_reauth rule.

## Related

- [BUG-019](BUG-019-block-rule-template-phantom-update.md) — block rule template phantom update
- KNOWN_API_ISSUES.md #13 — documents display name support and fallback behaviour
- GitHub issue #118
