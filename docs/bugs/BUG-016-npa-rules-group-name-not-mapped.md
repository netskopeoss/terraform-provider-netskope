# BUG-016: NPA Rules `group_name` Returned by API but Not Mapped in Provider

**Resource:** `netskope_npa_rules` (resource), `netskope_npa_rules_list` (data source)
**Severity:** High (rules lose policy group assignment on backup/restore — all rules fall into default group)
**Status:** Fixed (0.4.4) — removed `x-speakeasy-terraform-ignore` from `group_name` in both request and response OAS schemas, added `readOnly: true` to request schema, regenerated with Speakeasy
**Affected attributes:** `group_name` (missing), `group_id` (write-only)

---

## Summary

The Netskope API returns `group_name` in every NPA rules response (`GET /policy/npa/rules` and `GET /policy/npa/rules/{id}`), but the provider never maps it into state. The SDK model `NpaPolicyResponseItem` correctly defines `GroupName *string` with the JSON tag `json:"group_name,omitempty"`, and a getter `GetGroupName()` exists — but `RefreshFromSharedNpaPolicyResponseItem` ignores it.

Meanwhile, `group_id` is write-only: it can be set on create/update but the API never returns it in GET responses. This means once a rule is read back from the API, its group assignment is lost from state.

## Root Cause

The Speakeasy-generated refresh logic in `nparules_resource_sdk.go` only maps `Enabled`, `ID`, `RuleData`, and `RuleName` from the API response. Five fields present in the SDK response model are silently dropped:

- **`GroupName`** — the policy group this rule belongs to
- `ModifyBy` — last modifier (metadata)
- `ModifyTime` — last modification time (metadata)
- `ModifyType` — modification type (metadata)
- `PolicyType` — policy type (metadata)

Of these, `GroupName` is the only one with functional impact.

## Affected Files

### Resource (`netskope_npa_rules`)

1. **`internal/provider/nparules_resource.go`** — Add `GroupName` to `NPARulesResourceModel` and add `group_name` schema attribute:
   ```go
   // In NPARulesResourceModel struct:
   GroupName types.String `tfsdk:"group_name"`

   // In Schema():
   "group_name": schema.StringAttribute{
       Computed:    true,
       Description: "Policy group name this rule belongs to",
   },
   ```

2. **`internal/provider/nparules_resource_sdk.go`** — Map `GroupName` in `RefreshFromSharedNpaPolicyResponseItem`:
   ```go
   // After: r.RuleName = types.StringPointerValue(resp.RuleName)
   r.GroupName = types.StringPointerValue(resp.GroupName)
   ```

### List Data Source (`netskope_npa_rules_list`)

3. **`internal/provider/types/npa_policy_response_item.go`** — Add `GroupName` to type:
   ```go
   type NpaPolicyResponseItem struct {
       Enabled   types.String       `tfsdk:"enabled"`
       GroupName types.String       `tfsdk:"group_name"`
       ID        types.String       `tfsdk:"id"`
       RuleData  *NpaPolicyRuleData `tfsdk:"rule_data"`
       RuleName  types.String       `tfsdk:"rule_name"`
   }
   ```

4. **`internal/provider/nparuleslist_data_source.go`** — Add schema attribute in `data` nested object:
   ```go
   // After "enabled" attribute:
   "group_name": schema.StringAttribute{
       Computed: true,
   },
   ```

5. **`internal/provider/nparuleslist_data_source_sdk.go`** — Map in `RefreshFromSharedNpaPolicyResponse` loop:
   ```go
   // After: data.Enabled = types.StringPointerValue(dataItem.Enabled)
   data.GroupName = types.StringPointerValue(dataItem.GroupName)
   ```

## Impact

Without this fix, the `terraform-tenant-ops` module cannot preserve policy group assignments during backup/restore or cross-tenant clone/promote. All restored rules end up in the default policy group, requiring manual reassignment.

## Verification

After applying the fix, a `terraform plan` on an existing rule with a non-default group assignment should show `group_name` as a new Computed attribute in state with the correct group name, and no changes to `group_id`.
