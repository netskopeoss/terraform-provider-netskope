# BUG-005: Tag Name Case Mismatch Between Config and API Response

**Resource:** `netskope_npa_private_app`
**Severity:** Medium (affects every plan/apply when tag names differ in case from API canonical form)
**Status:** Fixed in 0.3.5
**Discovered:** 2026-02-13 via `test-examples.sh` against `full-deployment` example

---

## Summary

Tags are sent to the API by name only (no ID). The API performs a case-insensitive lookup and returns the tag with its canonical casing — e.g. config sends `"production"`, API returns `"Production"`. The plan modifier's `normalizeTagsOrder()` compared tag names case-sensitively, so when case differed it could not recognise the tags as matching and drift propagated.

## Reproduction

```hcl
resource "netskope_npa_private_app" "web_app" {
  # ...
  tags = [
    { tag_name = "production" },      # lowercase in config
    { tag_name = "infrastructure" },
  ]
}
```

API returns:
```json
{
  "tags": [
    {"tag_id": 1, "tag_name": "Production"},
    {"tag_id": 53, "tag_name": "infrastructure"}
  ]
}
```

Re-plan shows:

```
~ tags = [
    ~ { ~ tag_id = 1 -> (known after apply), ~ tag_name = "Production" -> "production" },
    ~ { ~ tag_id = 53 -> (known after apply),   tag_name = "infrastructure" },
  ]
```

## Root Cause

The `normalizeTagsOrder()` function in `npaprivateapp_resource_planmodify.go` builds sorted key lists from plan and state tag names, then compares them element-by-element:

```go
for i := range planKeys {
    if planKeys[i] != stateKeys[i] {  // case-sensitive!
        return  // exit without normalizing
    }
}
```

When `"production" != "Production"`, the function returns early. Without normalization:
- `tag_name` shows a diff (config value vs API canonical form)
- `tag_id` becomes `(known after apply)` because the state values aren't preserved

## Fix

Use `strings.EqualFold()` for case-insensitive comparison in `normalizeTagsOrder()`:

```go
for i := range planKeys {
    if !strings.EqualFold(planKeys[i], stateKeys[i]) {
        return
    }
}
```

This recognises `"production"` and `"Production"` as the same tag, allowing the state list (with canonical casing and resolved `tag_id`) to replace the plan list.

## Files Modified

| File | Change |
|------|--------|
| `internal/provider/npaprivateapp_resource_planmodify.go` | Use `strings.EqualFold()` in `normalizeTagsOrder()` |

## Related

- **BUG-002** — Config-order plan drift on list attributes (v0.3.5)
- **BUG-004** — Clientless hostname drift
