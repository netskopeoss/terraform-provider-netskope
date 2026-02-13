# BUG-004: Clientless App Hostname Rewritten by API Causes Plan Drift

**Resource:** `netskope_npa_private_app`
**Severity:** Medium (affects every plan/apply for clientless apps that set `private_app_hostname`)
**Status:** Fixed in 0.3.5
**Discovered:** 2026-02-13 via `test-examples.sh` against `full-deployment` example

---

## Summary

When `clientless_access = true`, the Netskope API replaces the user-provided `private_app_hostname` with an auto-generated internal hostname (`ns.<hash>`). On re-plan, Terraform sees the config value differs from state and proposes an update, causing perpetual drift.

## Reproduction

```hcl
resource "netskope_npa_private_app" "web_app" {
  private_app_name     = "production-internal-web"
  private_app_hostname = "web.production.internal"   # User provides this
  clientless_access    = true
  real_host            = "web.internal.local"
  # ...
}
```

First apply succeeds. API returns `private_app_hostname = "ns.9699421f"`. Second plan:

```
~ private_app_hostname = "ns.9699421f" -> "web.production.internal"
```

## Root Cause

`private_app_hostname` is `Computed: true, Optional: true`. For clientless apps, the API auto-generates the hostname — the user's value is only used as a seed. The provider stores the API-returned value in state, but the config still has the original.

### Data flow

```
Config: "web.production.internal"
  → Create request: host = "web.production.internal"
  → API response: host = "ns.9699421f"  (auto-generated for clientless)
  → State: "ns.9699421f"
  → Re-plan: config "web.production.internal" ≠ state "ns.9699421f" → DRIFT
```

Neither the AfterSuccess hook nor the plan modifier previously handled this case.

## Fix

Add logic to `ModifyPlan()` in `npaprivateapp_resource_planmodify.go`: when `clientless_access = true` and the state already has a hostname, preserve the state value in the plan. The API controls this field for clientless apps, so the user's config value should not override it after initial creation.

```go
// Suppress private_app_hostname drift for clientless apps.
// When clientless_access is true, the API auto-generates the hostname.
var clientless types.Bool
resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("clientless_access"), &clientless)...)
if !clientless.IsNull() && !clientless.IsUnknown() && clientless.ValueBool() {
    preserveStateForUnknownString(ctx, req, resp, path.Root("private_app_hostname"))
    // Also treat it as a "known from state" value when not unknown
    var planHost, stateHost types.String
    req.Plan.GetAttribute(ctx, path.Root("private_app_hostname"), &planHost)
    req.State.GetAttribute(ctx, path.Root("private_app_hostname"), &stateHost)
    if !stateHost.IsNull() && !stateHost.IsUnknown() && !planHost.IsNull() {
        resp.Plan.SetAttribute(ctx, path.Root("private_app_hostname"), stateHost)
    }
}
```

## Files Modified

| File | Change |
|------|--------|
| `internal/provider/npaprivateapp_resource_planmodify.go` | Add clientless hostname suppression in `ModifyPlan()` |

## Related

- **BUG-002** — Config-order plan drift on list attributes (v0.3.5)
- **BUG-005** — Tag name case mismatch drift
