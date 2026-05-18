# BUG-007: Empty protocols array causes API error on private app create/update

**Status:** Fixed in 0.3.6
**Branch:** `0.3.6-beta`
**Reported:** 2026-02-13
**Affected versions:** 0.3.2 — 0.3.5

## Symptom

Creating or updating a private app without specifying `protocols` fails with an API error. The API rejects an empty `protocols` array (`"protocols": []`).

## API behaviour confirmed

Tested 2026-02-23 against the live API. Protocols are **required for all private app types**:

```
# Clientless app without protocols
POST /api/v2/steering/apps/private
{"app_name": "test", "clientless_access": true, "real_host": "...", ...}

Response: {"status": "error", "message": "Required field 'protocols' can't be None or Empty."}

# Client-based app without protocols
POST /api/v2/steering/apps/private
{"app_name": "test", "clientless_access": false, "host": "...", ...}

Response: {"status": "error", "message": "Required field 'protocols' can't be None or Empty."}
```

No valid private app can exist without at least one protocol.

## Root cause

The OAS did not mark `protocols` as required, so Speakeasy generated `Optional: true, Computed: true` with no size validators. Users could omit protocols without a Terraform error, but the API would reject the request with a confusing error about port ranges.

The SDK also unconditionally creates an empty slice (`make([]shared.ProtocolItem, 0, ...)`), which serializes as `"protocols": []` rather than being omitted.

## Fix

Marked `protocols` as required with `minItems: 1` in the OAS for both `private_apps_request` and `private_apps_put_request` schemas, then regenerated with `speakeasy run`.

Speakeasy now generates:

```go
"protocols": schema.ListNestedAttribute{
    Required: true,
    Validators: []validator.List{
        listvalidator.SizeAtLeast(1),
    },
    // ...
}
```

Terraform rejects invalid configs at plan time with clear errors:
- Missing protocols: `The argument "protocols" is required, but no definition was found.`
- Empty protocols: `Attribute protocols list must contain at least 1 elements, got: 0`

### Previous mitigation (0.3.5)

The `hookPrivateAppRequest.go` hook was extended to strip empty `protocols` arrays from requests. This prevented the confusing API error but silently masked the config problem. The hook remains as a safety net but is no longer the primary fix.

## Files changed

| File | Change |
|------|--------|
| `endpoints/steering/npa_apps_private.yaml` | Added `required: [protocols]` and `minItems: 1` to both request schemas |
| `internal/provider/npaprivateapp_resource.go` | Regenerated — `protocols` is now `Required: true` with `SizeAtLeast(1)` validator |
| `internal/sdk/internal/hooks/hookPrivateAppRequest.go` | Previous mitigation (empty array stripping) — retained as safety net |

## Breaking change
    
Users with configs that omit `protocols` will get a plan error after upgrading. However, their configs were already broken — the API rejects them — so this surfaces an existing failure earlier with a clearer message.

## Test coverage

- `TestAccNPAPrivateApp_clientlessAccess` — Creates a clientless app with a single protocol (TCP/80), verifies creation succeeds
