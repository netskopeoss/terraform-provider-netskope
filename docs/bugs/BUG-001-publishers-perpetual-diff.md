# BUG-001: List Attribute Perpetual Diff on `netskope_npa_private_app`

**Resource:** `netskope_npa_private_app`
**Severity:** High (affects every plan/apply cycle)
**Status:** Fixed in 0.3.4 (see also BUG-002 for config-order follow-up fix in 0.3.5)
**Branch:** `0.3.4-beta`
**Affected attributes:** `publishers`, `protocols`, `tags`

---

## Summary

Three list attributes on `netskope_npa_private_app` — `publishers`, `protocols`, and `tags` — can trigger an "update in-place" on every `terraform plan` / `terraform apply`, even when no changes have been made to the configuration. The root cause is that all three are defined as `ListNestedAttribute` (order-sensitive) but the API returns their elements in non-deterministic order. Additionally, the API sometimes returns publisher names with ghost leading whitespace.

## Symptoms

### Publishers (reported bug)

```
~ publishers = [
    ~ {
        ~ publisher_id   = "249" -> "256"
        ~ publisher_name = "AWS-LSAC3-Prod-green" -> " AWS-LSAC3-Prod-blue"
      },
    ~ {
        ~ publisher_id   = "256" -> "249"
        ~ publisher_name = " AWS-LSAC3-Prod-blue" -> "AWS-LSAC3-Prod-green"
      },
    # (2 unchanged elements hidden)
  ]
```

**Key observations:**
- 2-3 out of 4 publishers rotate positions on every plan
- Some publisher names have a leading space (e.g., `" AWS-LSAC3-Prod-blue"`) that does not exist in the Netskope UI
- `ignore_changes = [publishers]` works around it but prevents legitimate updates
- `trimspace()` in HCL and manual sorting do not help — the problem is in the provider, not the configuration

### Protocols (KNOWN_API_ISSUES #11)

```
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
```

The API internally sorts protocols by type then port ascending, but this was undocumented. Users had to manually order their HCL to match.

### Tags (potential)

Same structural risk as publishers — `tag_id` + `tag_name` as a List with no ordering guarantee.

## Root Causes

### 1. List vs Set ordering sensitivity

**File:** `internal/provider/npaprivateapp_resource.go`

All three attributes are defined as `schema.ListNestedAttribute`:
- `publishers` (line 135)
- `protocols` (line 109)
- `tags` (line 158)

In Terraform, **lists are order-sensitive** — if element [0] and element [1] swap, Terraform detects a change. The Netskope API does not guarantee a stable ordering for any of these.

### 2. Whitespace in `publisher_name` from API

The API sometimes returns publisher names with leading whitespace (e.g., `" AWS-LSAC3-Prod-blue"`). The existing hook code converted `publisher_id` from int to string but passed `publisher_name` through unchanged.

## Fix Applied

All issues were fixed in the existing AfterSuccess hooks that already transform private app response data.

### Files Changed

#### `internal/sdk/internal/hooks/hookMyAppAfterSuccess.go`

**Added imports:** `sort`, `strconv`

**1. Publisher whitespace trimming** (inside the publisher_id type conversion loop):
```go
responseMap.Data.ServicePublisherAssignments[i].PublisherName = strings.TrimSpace(
    responseMap.Data.ServicePublisherAssignments[i].PublisherName,
)
```

**2. Publisher sorting** by publisher_id (numerically):
```go
sort.Slice(responseMap.Data.ServicePublisherAssignments, func(i, j int) bool {
    idI := fmt.Sprintf("%v", responseMap.Data.ServicePublisherAssignments[i].PublisherID)
    idJ := fmt.Sprintf("%v", responseMap.Data.ServicePublisherAssignments[j].PublisherID)
    numI, errI := strconv.Atoi(idI)
    numJ, errJ := strconv.Atoi(idJ)
    if errI == nil && errJ == nil {
        return numI < numJ
    }
    return idI < idJ
})
```

**3. Protocol sorting** by type alphabetically, then port numerically:
```go
sort.Slice(responseMap.Data.Protocols, func(i, j int) bool {
    ti := responseMap.Data.Protocols[i].Type
    tj := responseMap.Data.Protocols[j].Type
    if ti != tj {
        return ti < tj
    }
    pi, errI := strconv.Atoi(responseMap.Data.Protocols[i].Port)
    pj, errJ := strconv.Atoi(responseMap.Data.Protocols[j].Port)
    if errI == nil && errJ == nil {
        return pi < pj
    }
    return responseMap.Data.Protocols[i].Port < responseMap.Data.Protocols[j].Port
})
```

**4. Tag sorting** by tag_id:
```go
sort.Slice(responseMap.Data.Tags, func(i, j int) bool {
    return responseMap.Data.Tags[i].TagID < responseMap.Data.Tags[j].TagID
})
```

#### `internal/sdk/internal/hooks/hookMyBulkAppAfterSuccess.go`

**Added imports:** `sort`, `strconv`

All four changes (publisher trim, publisher sort, protocol sort, tag sort) applied inside the bulk app loop (`for i := range responseMap.BulkApps.AppData`), with identical logic.

### Files Added

#### `internal/sdk/internal/hooks/hookPublisherSort_test.go`

19 unit tests covering all three attributes plus edge cases:

| Test | What it verifies |
|------|------------------|
| **Publishers** | |
| `TestPublishersSortedByID` | Single-app hook sorts publishers by ID (all 3 operation IDs) |
| `TestPublisherNameWhitespaceTrimmed` | Single-app hook trims leading/trailing whitespace |
| `TestBulkPublishersSortedByID` | Bulk-app hook sorts publishers by ID for every app |
| `TestBulkPublisherNameWhitespaceTrimmed` | Bulk-app hook trims whitespace |
| `TestPublishersIdempotent` | Running the hook twice produces identical results |
| **Protocols** | |
| `TestProtocolsSortedByTypeThenPort` | Sorts by type (tcp before udp) then port (22 before 443) |
| `TestProtocolsSortedSameType` | Sorts by port when all protocols share the same type |
| `TestBulkProtocolsSorted` | Bulk-app hook sorts protocols for each app |
| **Tags** | |
| `TestTagsSortedByID` | Single-app hook sorts tags by tag_id |
| `TestBulkTagsSorted` | Bulk-app hook sorts tags for each app |
| **Combined** | |
| `TestAllFieldsIdempotent` | All three fields produce identical results across multiple hook runs |
| **Edge Cases** | |
| `TestEmptyListsDoNotPanic` | Empty publishers/protocols/tags arrays handled gracefully |
| `TestBulkEmptyListsDoNotPanic` | Bulk hook handles empty lists without panic |
| `TestNilPublisherIdHandling` | Nil publisher_id doesn't cause panic during sort |
| `TestIntTypedPublisherId` | Int-typed publisher_id (vs float64) converted correctly |
| `TestStringPublisherIdPassthrough` | String publisher_id passes through and sorts correctly |
| `TestNonNumericPortFallback` | Non-numeric ports (e.g., "1-1024") fall back to string comparison |
| `TestAppNameBracketTrimming` | App name brackets trimmed: `[app]` → `app` |
| `TestNonMatchingOperationIdPassthrough` | Unrecognized operation IDs pass through unchanged |

## Test Results

**Before fix:** 9 FAIL, 2 PASS (publisher idempotent + all-fields idempotent passed trivially)
**After fix:** 19 PASS, build clean

## Why hooks instead of OAS/schema change

- The schema is auto-generated by Speakeasy as `ListNestedAttribute`. Changing to `SetNestedAttribute` would require Speakeasy to support that annotation, and would be a schema-breaking change for existing state files.
- The hooks already transform publisher and protocol data (int→string conversion, transport→type mapping), so adding sorting is a minimal, consistent extension.
- The fix runs transparently at the HTTP response level before Terraform ever compares state, so no provider-level code changes are needed.

## Scope Analysis

All `ListNestedAttribute` fields with `Computed: true` + `Optional: true` on resource types were audited. The three affected fields are all on `netskope_npa_private_app`. No other resources in the provider have this pattern.

## Related

- **KNOWN_API_ISSUES.md #11** — Protocol ordering (now fixed by this change)
- `internal/provider/npaprivateapp_resource_test.go:84` — `ImportStateVerifyIgnore: []string{"publishers", ...}` was already present, indicating this was a known problem area