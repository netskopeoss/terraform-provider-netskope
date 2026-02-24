# BUG-009: Rule Creation Fails Immediately After Private App Creation

**Resource:** `netskope_npa_rules` (depends on `netskope_npa_private_app`)
**Severity:** High (rule creation fails non-deterministically)
**Status:** Fixed (provider-side mitigation) — API-side issue remains open
**GitHub Issue:** #65
**Affected operations:** Create (rule referencing a just-created private app)
**Fix branch:** `0.3.6-beta`

---

## Summary

When an NPA rule is created immediately after the referenced NPA private app within the same `terraform apply`, the rule creation intermittently fails with "Private app [app-X] doesn't exist". The private app resource reports successful creation, but the API backend has not yet propagated the record to the service handling rule validation. This indicates eventual consistency / read-after-write latency in the backend.

## Symptoms

```
netskope_npa_private_app.apps["app-X"]: Creating...
netskope_npa_private_app.apps["app-X"]: Creation complete after 1s

Error: failure to invoke API

  netskope_npa_rules.rules["rule-for-app-X"],
  API error: Private app [app-X] doesn't exist
```

The private app exists (confirmed via UI/API query), but the rules endpoint cannot see it yet.

## Root Cause

The API backend uses eventual consistency between the private apps service and the rules/policy service. After a private app is created and the API returns a 200 success, the record is not immediately visible to other backend services that validate rule references. The propagation delay is typically 5–10 seconds.

This is entirely API-side; the provider correctly passes the app ID returned from the create response.

## Fix

An HTTP client wrapper (installed via SDKInit hook) transparently retries rule creation requests that fail with "doesn't exist" or "does not exist" errors. The retry uses exponential backoff:

| Attempt | Delay | Cumulative |
|---------|-------|------------|
| 1 | 0s (first try) | 0s |
| 2 | 2s | 2s |
| 3 | 4s | 6s |
| 4 | 8s | 14s |
| 5 | 16s | 30s |
| 6 | 30s (capped) | 60s |

Most cases resolve by attempt 3 (~6s), matching the typical 5–10s propagation delay reported by the issue author. After all retries are exhausted, the last error response is returned and the user sees the original API error.

Implemented as an HTTP client wrapper rather than a BeforeRequest/AfterSuccess hook so the retry happens transparently before any hooks process the response. The retry runs while the BUG-008 serializer mutex is held, preventing other rule creates from interleaving during backoff.

### Why not use Speakeasy's `x-speakeasy-retries`?

The API returns HTTP 200 OK with the error in the JSON response body (`{"status": "error", "message": "... doesn't exist"}`). Speakeasy's retry extension only triggers on HTTP status codes (e.g., 5XX), so it cannot detect this failure pattern.

### Why not use Terraform's `retry.RetryContext`?

This is the HashiCorp best practice for eventual consistency, but the resource Create method is in generated code (`nparules_resource.go`) that gets overwritten on `speakeasy run`. The hooks directory is the only durable location for custom logic.

## Relevant Files

| File | Role |
|------|------|
| `internal/sdk/internal/hooks/hookRuleCreateRetry.go` | HTTP client wrapper with retry logic |
| `internal/sdk/internal/hooks/hookRuleCreateRetry_test.go` | Unit tests (11 tests) |
| `internal/sdk/internal/hooks/registration.go` | SDKInit hook registration |
| `internal/sdk/internal/hooks/hookErrorStatusResponse.go` | Converts "200 with error" to Go errors (runs after retry) |

## User Workaround (pre-0.3.6)

Add a `time_sleep` resource between app creation and rule creation:

```hcl
resource "time_sleep" "wait_for_app" {
  depends_on      = [netskope_npa_private_app.apps]
  create_duration = "10s"
}

resource "netskope_npa_rules" "rules" {
  depends_on = [time_sleep.wait_for_app]
  # ...
}
```

This workaround is no longer needed with the 0.3.6 fix.

## Related

- Issue #66 (BUG-008) — concurrent rule creation race condition, fixed with mutex serialization in the same release.
