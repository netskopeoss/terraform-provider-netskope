# BUG-011: API Adds Space to `private_app_hostname` Causing Perpetual Drift

**Resource:** `netskope_npa_private_app`
**Severity:** Medium (perpetual diff on every plan/apply for multi-host apps)
**Status:** Open
**Affected attributes:** `private_app_hostname`

---

## Summary

When a user configures `private_app_hostname` with multiple comma-separated hosts (e.g. `"10.33.5.143,webapp02.baincapital.com"`), the Netskope API normalizes the whitespace around commas on read-back. This causes the state value to differ from the config value, producing a perpetual plan diff on every subsequent `terraform plan`.

This is distinct from **BUG-004** (clientless hostname rewrite), which handles the case where `clientless_access = true` and the API replaces the entire hostname with an auto-generated `ns.<hash>`. This bug affects standard (non-clientless) private apps with multiple hosts.

## Symptoms

```
$ terraform plan

  # netskope_npa_private_app.example will be updated in-place
  ~ resource "netskope_npa_private_app" "example" {
      ~ private_app_hostname = "10.33.5.143, webapp02.baincapital.com" -> "10.33.5.143,webapp02.baincapital.com"
        # (other attributes unchanged)
    }
```

This repeats on every plan. The direction of the diff depends on whether the user's config has spaces after commas or not — the API enforces its own formatting either way.

## Reproduction

```hcl
resource "netskope_npa_private_app" "multi_host" {
  private_app_name     = "multi-host-test"
  private_app_hostname = "10.33.5.143,webapp02.baincapital.com"
  real_host            = "10.33.5.143"
  use_publisher_dns    = true

  protocols {
    port     = "443"
    protocol = "tcp"
  }

  publishers {
    publisher_id = "1"
  }

  tags {
    tag_name = "test"
  }
}
```

First apply succeeds. API returns `private_app_hostname` with different whitespace around commas. Second plan shows drift.

## Root Cause

`private_app_hostname` is `Computed: true, Optional: true` in the schema (`npaprivateapp_resource.go:148`). The FromSDK function stores the API response value directly into state with no normalization:

```go
// npaprivateapp_resource_sdk.go:65 (Read) and :159 (Create)
r.PrivateAppHostname = types.StringPointerValue(resp.PrivateAppHostname)
```

The API normalizes whitespace around commas in multi-host strings. For example:
- Config sends: `"10.33.5.143,webapp02.baincapital.com"` (no spaces)
- API returns: `"10.33.5.143, webapp02.baincapital.com"` (space after comma)

Or the reverse — user includes spaces, API strips them. The exact behavior may vary by API version, but the result is the same: config and state diverge on whitespace.

### Data flow

```
Config: "10.33.5.143,webapp02.baincapital.com"
  → Create/Update request: host = "10.33.5.143,webapp02.baincapital.com"
  → API response: host = "10.33.5.143, webapp02.baincapital.com"  (space added)
  → State: "10.33.5.143, webapp02.baincapital.com"
  → Re-plan: config "...143,webapp02..." ≠ state "...143, webapp02..." → DRIFT
```

The existing `suppressClientlessHostnameDrift` in `ModifyPlan()` only fires when `clientless_access = true`, so it does not cover this case.

## Customer Impact

Reported by customer running the `infrastructure-netskope-private-access` module with YAML-driven app definitions. Multiple YAML files use comma-separated hostnames with spaces (e.g. `"10.33.5.143, webapp02.baincapital.com"`). Every `terraform plan` shows in-place updates for these resources even when no configuration has changed.

## Proposed Fix

### Provider-side (recommended)

Add a hostname whitespace normalizer in `ModifyPlan()` within `npaprivateapp_resource_planmodify.go`. When both config and state have a value for `private_app_hostname`, normalize both by splitting on commas, trimming whitespace from each entry, and joining with a canonical separator before comparison:

```go
func suppressHostnameWhitespaceDrift(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
    var planHost, stateHost types.String
    resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("private_app_hostname"), &planHost)...)
    resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("private_app_hostname"), &stateHost)...)
    if resp.Diagnostics.HasError() {
        return
    }

    if planHost.IsNull() || planHost.IsUnknown() || stateHost.IsNull() || stateHost.IsUnknown() {
        return
    }

    // Normalize both values: split on comma, trim each entry, rejoin
    normalize := func(s string) string {
        parts := strings.Split(s, ",")
        for i, p := range parts {
            parts[i] = strings.TrimSpace(p)
        }
        return strings.Join(parts, ",")
    }

    if normalize(planHost.ValueString()) == normalize(stateHost.ValueString()) {
        // Whitespace-only difference — keep the state value to suppress diff
        resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("private_app_hostname"), stateHost)...)
    }
}
```

Then call it from `ModifyPlan()`:

```go
suppressHostnameWhitespaceDrift(ctx, req, resp)
```

### Module-side workaround

Until the provider is fixed, customers can normalize the hostname in their module before passing it to the resource:

```hcl
private_app_hostname = join(",", [for h in split(",", var.app.hostname) : trimspace(h)])
```

## Relevant Files

| File | Role |
|------|------|
| `internal/provider/npaprivateapp_resource.go:148` | Schema: `private_app_hostname` is `Computed + Optional` |
| `internal/provider/npaprivateapp_resource_sdk.go:65` | FromSDK (Read): stores API hostname directly in state |
| `internal/provider/npaprivateapp_resource_sdk.go:159` | FromSDK (Create): stores API hostname directly in state |
| `internal/provider/npaprivateapp_resource_sdk.go:286` | ToSDK (Create): sends config hostname to API |
| `internal/provider/npaprivateapp_resource_sdk.go:461` | ToSDK (Update): sends config hostname to API |
| `internal/provider/npaprivateapp_resource_planmodify.go:40` | ModifyPlan — needs new normalizer |

## Related

- **BUG-004** — Clientless app hostname rewritten by API (different root cause, same attribute)
- **BUG-010** — Port CSV drift (same pattern: API normalizes comma-separated values)
