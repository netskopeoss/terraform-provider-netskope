# BUG-003: BeforeRequest Hook Unmarshals rule_order.rule_id as String, SDK Sends Int

**Resource:** `netskope_npa_rules`
**Severity:** High (blocks creation of any rule using `rule_order.order = "after"` with a `rule_id`)
**Status:** Fixed in 0.3.5
**Discovered:** 2026-02-13 via `test-examples.sh` drift detection run against `full-deployment` example

---

## Summary

Creating an `netskope_npa_rules` resource with `rule_order = { order = "after", rule_id = <id> }` fails during the `BeforeRequest` hook with:

```
ERROR: Unable to unmarshal response: json: cannot unmarshal number into Go
struct field RuleOrder.rule_order.rule_id of type string
```

The SDK model serializes `rule_id` as a JSON number (`*int64`), but the BeforeRequest hook's local `RuleOrder` struct declares it as `*string`. When the hook reads the request body to wrap private app names in brackets, `json.Unmarshal` fails on the type mismatch.

## Reproduction

The `full-deployment` example triggers this. The first rule creates successfully, but the second rule (which references the first via `rule_id`) fails:

```hcl
resource "netskope_npa_rules" "web_app_access" {
  rule_name = "web-app-access"
  # ...
  rule_order = { order = "top" }  # No rule_id - works fine
}

resource "netskope_npa_rules" "ssh_access" {
  rule_name = "ssh-access"
  # ...
  rule_order = {
    order   = "after"
    rule_id = tonumber(netskope_npa_rules.web_app_access.id)  # Fails here
  }
}
```

## Root Cause

Three different `RuleOrder` structs exist with inconsistent types for `rule_id`:

| Location | File | `rule_id` Type |
|----------|------|----------------|
| SDK request model | `internal/sdk/models/shared/npapolicyrequest.go:45` | `*int64` |
| Terraform framework model | `internal/provider/types/rule_order.go:12` | `types.Int64` |
| BeforeRequest hook model | `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go:57` | `*string` |

The SDK serializes the request body with `rule_id` as a JSON number (matching `*int64`). The BeforeRequest hook then reads this body and tries to unmarshal it into its own struct where `rule_id` is `*string`. Go's `encoding/json` does not coerce numbers to strings, so it fails.

### Data flow

```
Terraform config (types.Int64)
  → SDK model (*int64)
  → JSON body: {"rule_order": {"order": "after", "rule_id": 4}}
  → BeforeRequest hook: json.Unmarshal into RuleOrder{RuleID *string}
  → FAIL: cannot unmarshal number into string
```

## Fix

Change `RuleID` from `*string` to `*int64` in the hook's `RuleOrder` struct at `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go:57`:

```go
// Before (broken)
type RuleOrder struct {
	Order    *string `json:"order"`
	Position *int64  `json:"position"`
	RuleID   *string `json:"rule_id"`
	RuleName *string `json:"rule_name"`
}

// After (fixed)
type RuleOrder struct {
	Order    *string `json:"order"`
	Position *int64  `json:"position"`
	RuleID   *int64  `json:"rule_id"`
	RuleName *string `json:"rule_name"`
}
```

The hook does not read or modify `RuleOrder` — it only processes `RuleData.PrivateApps` (wrapping names in brackets). The `RuleOrder` struct just needs to round-trip correctly through unmarshal/marshal.

## Files to Modify

| File | Change |
|------|--------|
| `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go:57` | `RuleID *string` → `RuleID *int64` |

## Verification

```bash
# Build check
go build ./...

# Run the full-deployment example which exercises rule_order with rule_id
cd ../terraform-netskope-examples
./scripts/test-examples.sh

# Or test manually
cd examples/full-deployment
terraform apply -auto-approve
```

## Related

- **BUG-001** — Non-deterministic API response ordering (v0.3.4)
- **BUG-002** — Config-order-dependent plan drift on list attributes (v0.3.5)
