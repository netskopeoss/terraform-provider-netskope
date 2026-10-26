---
title: BUG-018 — netskope_npa_rules read crashes when rule_data.classification is set
status: Fixed
github_issue: 106
---

## Summary

`terraform import` (and any state refresh) of an NPA rule with a Device Classification
(managed/unmanaged) set in the tenant UI crashes with:

```
Error: failure to invoke API
ERROR: Unable to unmarshal response: json: cannot unmarshal array into Go struct field RuleData.data.rule_data.classification of type string
```

## Root Cause

`GET /api/v2/policy/npa/rules/{id}` returns `rule_data.classification` as an **array of
strings** for rules that have a device classification set:

```json
{
  "data": {
    "rule_id": "224",
    "rule_data": {
      "classification": ["unmanaged"],
      ...
    }
  }
}
```

The `AfterSuccess` hook (`hookMyPolicyAfterSuccess.go`) runs on every `getNPARules`
response. It reads the raw response body and deserializes it into a local
`myPolicyResponse` struct, which embeds `models.PolicyData` → `models.RuleData`.
Both `models.RuleData.Classification` and the hook-local `RuleData.Classification`
were typed as `*string`. When `json.Unmarshal` encounters an array where it expects
a string pointer, it returns an error — matching the error format used by the hook:

```
ERROR: Unable to unmarshal response: <json error>
```

The Speakeasy-generated `NpaPolicyRuleData.Classification []string` (the public SDK
model) was already correctly typed as a slice — this bug was entirely in the two
manually-maintained hook structs.

## Impact

Any NPA rule with a UI-set managed/unmanaged device classification is unusable:
- `terraform import` fails
- Every state refresh fails if the rule is in state

Rules created by the provider itself are unaffected because `classification` has
`x-speakeasy-terraform-ignore: true` — the provider never writes it.

## Fix

Changed `Classification *string` → `Classification []string` in:

1. `internal/sdk/internal/models/npaPolicyStruct.go` — `RuleData.Classification`
2. `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go` — local `RuleData.Classification`

No OAS changes needed. The Speakeasy SDK model was already correct.

## Files Changed

- `internal/sdk/internal/models/npaPolicyStruct.go`
- `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go`
- `internal/sdk/internal/hooks/hookMyPolicyAfterSuccess_test.go` (new)

## Tests

Added unit test `TestAfterSuccess_ClassificationArrayUnmarshal` in
`hookMyPolicyAfterSuccess_test.go` that verifies the AfterSuccess hook successfully
processes a `getNPARules` response containing `"classification": ["unmanaged"]`.
