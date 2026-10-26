---
title: BUG-020 — netskope_npa_rules crashes on import when action_name is "periodic_reauth"; schema rejects "periodic_reauth" as a valid action
status: Fixed
github_issue: 116
---

## Summary

Two related failures for NPA rules using the Periodic Authentication action:

1. **Import crash** — `terraform import` of any rule with the Periodic Authentication action set
   (via the Netskope UI) crashes immediately with:
   ```
   Error: failure to invoke API
   error unmarshaling json response body: invalid value for ActionName: periodic_reauth
   ```

2. **Schema rejection** — Setting `action_name = "periodic_reauth"` in HCL config fails at
   plan time with a validation error, so users cannot manage Periodic Authentication rules
   via Terraform at all.

## Root Cause

The `action_name` field in `match_criteria_action` had an incomplete enum in the OAS:

```yaml
action_name:
  type: string
  enum: [allow, block]   # missing: periodic_reauth
```

Speakeasy generated a Go enum type with a strict `UnmarshalJSON`:

```go
func (e *ActionName) UnmarshalJSON(data []byte) error {
    switch v {
    case "allow": fallthrough
    case "block":
        *e = ActionName(v)
        return nil
    default:
        return fmt.Errorf("invalid value for ActionName: %v", v)
    }
}
```

And a Terraform schema validator:

```go
stringvalidator.OneOf("allow", "block")
```

Both the SDK deserializer and the schema validator blocked any rule using
`"periodic_reauth"` as the action. Rules created or modified via the UI with
the Periodic Authentication action were entirely unusable with this provider.

## Impact

- `terraform import` of any Periodic Authentication rule → crash on every state refresh
- Setting `action_name = "periodic_reauth"` in HCL → plan-time validation error
- Rules with `periodic_reauth` set in rule_data but `action_name = "allow"` silently
  had no periodic auth enforcement (the API ignores the `periodic_reauth` block
  when `action_name` is `"allow"`)

## API Behaviour (confirmed from docs and live export)

The Netskope API accepts and returns three action names for NPA rules:
- `"allow"` — grant access
- `"block"` — deny access (requires notification template display name)
- `"periodic_reauth"` — Periodic Authentication (requires `os` set to Windows and/or macOS;
  notification template display name recommended)

The `template` field in `match_criteria_action` has the same display-name/file-name mismatch
for `periodic_reauth` as for `block`: the API accepts display names on create/update but
returns `.html` file names on GET. The existing `suppressTemplateDrift` plan modifier and
BeforeRequest hook `.html` stripping already handle this transparently for any action —
no additional changes were needed.

## Fix

Added `periodic_reauth` to the `action_name` enum in the OAS:

```yaml
action_name:
  type: string
  enum: [allow, block, periodic_reauth]
```

Ran `speakeasy run` to regenerate `internal/sdk/models/shared/npapolicyruledata.go`
(adds `ActionNamePeriodicReauth` constant and `case "periodic_reauth"` to `UnmarshalJSON`)
and `internal/provider/nparules_resource.go` (adds `"periodic_reauth"` to the
`stringvalidator.OneOf` validator and updates the field description).

No hook changes were required — the BeforeRequest `.html` stripping and `suppressTemplateDrift`
plan modifier are both action-name-agnostic.

## Files Changed

- `~/speakeasy/netskope-apiv2-oas/endpoints/policy/npa_policy.yaml` — added `periodic_reauth` to enum; updated `template` description
- `internal/sdk/models/shared/npapolicyruledata.go` — regenerated (new enum constant + unmarshal case)
- `internal/provider/nparules_resource.go` — regenerated (validator + description)
- `internal/provider/testdata/TestAccNPARules_withPeriodicReauth/1/main.tf` — updated to use `action_name = "periodic_reauth"` + `os = ["Windows"]`
- `internal/provider/testdata/TestAccNPARules_withPeriodicReauth/2/main.tf` — same
- `internal/provider/testdata/TestAccNPARules_periodicReauthImport/main.tf` — new import test config
- `internal/provider/nparules_resource_test.go` — updated `TestAccNPARules_withPeriodicReauth`; added `TestAccNPARules_periodicReauthImport`
- `internal/provider/drift_detection_test.go` — updated `testAccDriftNPARulesPeriodicReauthConfig` to use correct action

## Tests Added / Updated

- `TestAccNPARules_withPeriodicReauth` — now creates with `action_name = "periodic_reauth"` and checks the attribute round-trips correctly; update step also verifies no drift
- `TestAccNPARules_periodicReauthImport` — creates a periodic_reauth rule then imports it; verifies the import no longer crashes (regression for issue #116)
- `TestAccDrift_NPARules_PeriodicReauth` — updated config to use `action_name = "periodic_reauth"` + `os = ["Windows"]`
