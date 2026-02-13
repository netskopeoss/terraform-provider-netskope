# BUG-002: Config-Order-Dependent Plan Drift on List Attributes

**Resources:** `netskope_npa_private_app`, `netskope_npa_rules`
**Severity:** High (affects every plan/apply cycle when config order differs from state order)
**Status:** Fixed in 0.3.5
**Branch:** `0.3.5-beta`
**Predecessor:** BUG-001 (fixed in 0.3.4 — hook-side sorting)

**Affected attributes:**

| Resource | Attribute | Type |
|----------|-----------|------|
| `npa_private_app` | `protocols` | `ListNestedAttribute` |
| `npa_private_app` | `publishers` | `ListNestedAttribute` |
| `npa_private_app` | `tags` | `ListNestedAttribute` |
| `npa_rules` | `rule_data.private_apps` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.users` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.user_groups` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.access_method` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.src_countries` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.organization_units` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.net_location_obj` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.private_app_tags` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.private_app_tag_ids` | `ListAttribute[string]` |
| `npa_rules` | `rule_data.device_classification_id` | `ListAttribute[int64]` |

---

## Summary

BUG-001 (v0.3.4) made state deterministic by sorting API responses in AfterSuccess hooks. However, Terraform's `ListNestedAttribute` and `ListAttribute` compare plan vs state by **position** — index 0 vs index 0, index 1 vs index 1. When the user's HCL config order differs from the hook's sorted state order, every `terraform plan` shows spurious "update in-place" diffs even though the set of elements is identical.

## Symptoms

```hcl
# User writes protocols in a natural order (primary port first)
protocols = [
  { port = "443", protocol = "tcp" },
  { port = "22",  protocol = "tcp" },
]
```

Hook sorts state to `[tcp:22, tcp:443]` (port ascending). Next plan:

```
~ protocols = [
    ~ { ~ port = "80"  -> "443" },   # state[0] vs plan[0]
    ~ { ~ port = "443" -> "80"  },   # state[1] vs plan[1]
  ]
```

The elements are identical — only their positions differ.

## Root Cause

BUG-001 fixed the **read side** (API responses sorted before writing to state). Nothing normalized the **plan side** (from user's HCL config). Terraform sees positional differences and reports changes.

```
Plan (from config)                     State (from hooks)
  protocols[0] = tcp:443                 protocols[0] = tcp:22
  protocols[1] = tcp:22                  protocols[1] = tcp:443
                                         ↑ same elements, different positions
```

## Fix

Uses the Terraform Plugin Framework's `ResourceWithModifyPlan` interface to intercept the plan before comparison. For each list attribute:

1. Get the plan list and state list
2. If either is null/unknown, or lengths differ — skip (create, destroy, or real change)
3. Extract a key from each element (the field the user controls)
4. Sort both key lists and compare
5. If identical — replace the plan list with the state list (same elements + computed values)
6. If different — leave plan untouched (genuine change, show diff)

The plan is replaced with the state list (not just reordered) because state elements contain computed sub-attributes (like `tag_id`) that the plan doesn't have. This also eliminates "known after apply" noise.

### Key fields used for comparison

| Attribute | Key | Why |
|-----------|-----|-----|
| `protocols` | `protocol:port` (e.g. `tcp:443`) | Both fields are user-controlled |
| `publishers` | `publisher_id` (fallback: `publisher_name`) | ID preferred; name used when ID is unknown |
| `tags` | `tag_name` | `tag_id` is Computed-only, never in config |
| `npa_rules` string lists | the string value itself | Flat list, element is its own key |
| `npa_rules` int64 lists | the int64 value itself | Flat list, element is its own key |

## Files Added

| File | Purpose |
|------|---------|
| `internal/provider/npaprivateapp_resource_planmodify.go` | `ModifyPlan` for private app (protocols, publishers, tags) |
| `internal/provider/npaprivateapp_resource_planmodify_test.go` | Unit tests for key extraction and sorted comparison |
| `internal/provider/nparules_resource_planmodify.go` | `ModifyPlan` for rules (10 list attributes under rule_data) |
| `internal/provider/nparules_resource_planmodify_test.go` | Unit tests for string and int64 list comparison |

## Files Modified

| File | Change |
|------|--------|
| `internal/provider/drift_detection_test.go` | Added 3 acceptance tests with deliberately unsorted HCL configs |
| `.genignore` | Added both `_planmodify.go` files to prevent Speakeasy overwrite |

## How Go Makes This Work

In Go, methods on a struct can be defined in any file within the same package. The generated `npaprivateapp_resource.go` defines the struct and CRUD methods; our new file adds the `ModifyPlan` method. The framework discovers `ResourceWithModifyPlan` via runtime type assertion — no registration needed.

## Edge Cases

| Scenario | Behaviour | Correct? |
|----------|-----------|----------|
| Create (no prior state) | Skip — `req.State.Raw.IsNull()` | Yes |
| Destroy (null plan) | Skip — `req.Plan.Raw.IsNull()` | Yes |
| Add element (plan has 3, state has 2) | Length mismatch — skip | Yes |
| Remove element (plan has 2, state has 3) | Length mismatch — skip | Yes |
| Change element value | Keys don't match — skip | Yes |
| Change + reorder | Keys don't match — skip | Yes |
| Plan already matches state order | Keys match — set plan to state (no-op) | Yes |
| Unknown values (new resource ref) | Bail out, no normalization | Yes |
| Duplicate elements | Sorted comparison handles correctly | Yes |

## Verification

```bash
# Unit tests
go test -v ./internal/provider/... -run "TestProtocolKey|TestPublisherKey|TestTagKey|TestSortedKeysMatch"
go test -v ./internal/provider/... -run "TestStringList|TestInt64List"

# Acceptance tests (requires API credentials)
TF_ACC=1 go test -v ./internal/provider/... -run "TestAccDrift_PrivateApp_Unsorted" -timeout 30m -parallel 1
TF_ACC=1 go test -v ./internal/provider/... -run "TestAccDrift_NPARules_Unsorted" -timeout 30m -parallel 1

# Existing tests (no regression)
TF_ACC=1 go test -v ./internal/provider/... -run "TestAccDrift" -timeout 30m -parallel 1
```

## Related

- **BUG-001** — Fixed non-deterministic API response ordering via hook-side sorting (v0.3.4)
- **BUG-002-private-app-lifecycle.md** — Detailed pseudo-code tracing the full lifecycle
- **KNOWN_API_ISSUES.md #11** — Protocol ordering
