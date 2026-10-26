---
title: BUG-019 — Block rules with template produce phantom update on every plan; any update fails with "Undefined template: *.html"
status: Fixed
github_issue: 105
---

## Summary

Any `netskope_npa_rules` block rule that has a `template` set produces a phantom
update-in-place on every plan, and when that update is applied it always fails:

```
Error: failure to invoke API
API error: Undefined template: 23.html
```

This also breaks *legitimate* updates to a block rule (e.g. toggling `enabled`) —
they fail with the same error.

## Root Cause

Three problems stack:

### 1. `group_name` is Computed-only but never returned by GET-by-ID

`group_name` has no `UseStateForUnknown` plan modifier. On every update plan,
Terraform marks it `(known after apply)`. This keeps the plan "dirty" (an update
is required) even after `suppressTemplateDrift` has equalised the `template` field.

### 2. `suppressTemplateDrift` sets plan.template to the state value (file name)

`suppressTemplateDrift` reads the state template value (the API-returned file name,
e.g. `"23.html"`) and writes it into the plan to suppress the display-name vs
file-name diff. This works for eliminating phantom diffs, but means the *plan
value* is now the file name.

### 3. When any update fires, the PUT body contains the file name

Terraform serialises the plan into the PUT body. Since plan.template is `"23.html"`
(set by `suppressTemplateDrift`), the request sends `"template": "23.html"`. The API
rejects any file name with `"Undefined template: 23.html"`.

This means even a real, intentional change to a block rule (changing `enabled`,
adding a private app, etc.) will fail if a template is set — the file name leaks
into every subsequent update payload.

## Fix

### Part 1 — Eliminate the `group_name` phantom update (`nparules_resource_planmodify.go`)

Extended `preserveComputedStringAttr` to handle null state (previously it only
preserved when state had a real value). Added a call for `group_name`:

```go
preserveComputedStringAttr(ctx, req, resp, path.Root("group_name"))
```

When state is null (as it always is for `group_name`, since GET-by-ID doesn't
return it) and plan is unknown, the plan is set to null. `null == null` → no diff
→ the phantom update is eliminated.

### Part 2 — Strip file names from update payloads (`hookMyPolicyBeforeRequest.go`)

Added logic to strip `.html` template values from `updateNPARules` payloads:

```go
if hookCtx.OperationID == "updateNPARules" &&
    requestMap.RuleData.MatchCriteriaAction != nil &&
    requestMap.RuleData.MatchCriteriaAction.Template != nil &&
    strings.HasSuffix(*requestMap.RuleData.MatchCriteriaAction.Template, ".html") {
    requestMap.RuleData.MatchCriteriaAction.Template = nil
}
```

If `suppressTemplateDrift` set the plan template to a file name, stripping it from
the PUT body means the API preserves its existing template. Display names (no `.html`
suffix) are left unchanged — intentional template changes still work.

Create payloads are not affected. The API accepts display names on create and they
are never `.html` file names from user config.

## Files Changed

- `internal/provider/nparules_resource_planmodify.go` — extended `preserveComputedStringAttr` to handle null state; added `group_name` call
- `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go` — strip `.html` template values from update payloads

## Tests Added

- `TestBeforeRequest_HtmlTemplateStrippedOnUpdate` — `.html` template stripped from update
- `TestBeforeRequest_DisplayNamePreservedOnUpdate` — display name kept on update
- `TestBeforeRequest_DisplayNamePreservedOnCreate` — display name kept on create
- `TestAccNPARules_denyRule` — unskipped (block rule create/read now works with display names)
- `TestAccDrift_NPARules_BlockTemplate` — unskipped (verifies no phantom drift after create)

## Related

- `docs/KNOWN_API_ISSUES.md` #13 updated: block rules can be created with display names
- Issue reporter noted KNOWN_API_ISSUES #13 was out of date — confirmed correct
