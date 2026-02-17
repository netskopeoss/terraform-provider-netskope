# BUG-006: `private_app_tag_ids` Perpetual Diff on `netskope_npa_rules`

**Resource:** `netskope_npa_rules`
**Severity:** Medium (perpetual diff on every plan/apply, workaround available)
**Status:** Open
**Branch:** `0.3.6-beta`
**Affected attributes:** `rule_data.private_app_tag_ids`

---

## Summary

When a user configures an NPA rule with `private_app_tags` (by name), the API resolves the tag names to IDs and returns `private_app_tag_ids` in the response. On subsequent plans, Terraform sees `private_app_tag_ids` in state but not in the user's config, so it plans to remove the field every cycle.

## Symptoms

```
$ terraform apply -target=netskope_npa_rules.test_policy

  # netskope_npa_rules.test_policy will be updated in-place
  ~ resource "netskope_npa_rules" "test_policy" {
      ~ id        = "27" -> (known after apply)
      ~ rule_data = {
          ~ private_app_tag_ids      = [
              - "49",
            ]
            # (14 unchanged attributes hidden)
        }
        # (2 unchanged attributes hidden)
    }
```

This repeats on every apply. The tag IDs are removed, the API re-resolves them from `private_app_tags`, and the cycle continues.

## Workaround

```hcl
resource "netskope_npa_rules" "example" {
  # ...
  lifecycle {
    ignore_changes = [
      rule_data.private_app_tag_ids,
    ]
  }
}
```

## Root Cause

`privateAppTagIds` in the OAS (`npa_policy.yaml:174`) is missing `x-speakeasy-terraform-ignore: true`. This causes Speakeasy to generate a schema with `Computed: true, Optional: true, Default: []`, making Terraform treat it as a user-settable field.

The data flow:

1. User sets `private_app_tags = ["terraform"]` in config
2. ToSDK sends both `privateAppTags` and `privateAppTagIds` (empty) to the API
3. API resolves tags to IDs, returns `privateAppTagIds: ["49"]`
4. FromSDK stores `["49"]` in state
5. Next plan: config has no `private_app_tag_ids` (defaults to `[]`), state has `["49"]` — diff detected
6. The existing ModifyPlan normalizer can't fix this because it's not a reordering issue — it's a field that shouldn't be user-settable at all

## Relevant Files

| File | Role |
|------|------|
| `endpoints/policy/npa_policy.yaml:174` | OAS definition missing `x-speakeasy-terraform-ignore` |
| `internal/provider/nparules_resource.go:149` | Generated schema: `Computed + Optional + Default: []` |
| `internal/provider/nparules_resource_sdk.go:262` | ToSDK sends empty tag IDs to API |
| `internal/provider/nparules_resource_sdk.go:107` | FromSDK reads tag IDs from API response |
| `internal/provider/nparules_resource_planmodify.go:45` | ModifyPlan normalizes order but can't fix this |