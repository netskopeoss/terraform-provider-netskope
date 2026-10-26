# BUG-021 — NPA rules template drift and rule_order reorder on update

**Fixed in:** v0.4.13
**Issues:** [#129](https://github.com/netskopeoss/terraform-provider-netskope/issues/129), [#124](https://github.com/netskopeoss/terraform-provider-netskope/issues/124), [#118](https://github.com/netskopeoss/terraform-provider-netskope/issues/118)
**Regresses:** [#105](https://github.com/netskopeoss/terraform-provider-netskope/issues/105), [#79](https://github.com/netskopeoss/terraform-provider-netskope/issues/79), [#116](https://github.com/netskopeoss/terraform-provider-netskope/issues/116)

## Problem

The `netskope_npa_rules` resource had two related defects that went through five release cycles of patches, each fixing one symptom while reintroducing another.

### RC-1 / RC-2 — Template drift (#129 P1, #124, #118)

The Netskope API accepts template display names on create/update but returns an internal `.html` filename on GET. Every release since 0.4.4 attempted to reconcile this mismatch using in-process caches (`npaTemplateCache`, `npaTemplatesAPICache`). Those caches do not survive between Terraform runs (Terraform Cloud, `terraform-plugin-testing`), so the reconciliation only worked within a single provider process.

On tenants where `/api/v2/templates/usernotifications` returns 403 (common on customer tenants), or under OAuth2 authentication, the cache never populated, and a different code path ran — one that was never covered by the acceptance tests (which run against the alliances tenant where the endpoint works).

### RC-5 — `rule_order` sent on every update (#129 P2)

`Update` called the same `ToSharedNpaPolicyRequest` as `Create`, which serializes `rule_order`. With `lifecycle { ignore_changes = [rule_order] }`, the prior value `{order = "bottom"}` was still in the plan and sent on every PUT. The API honours `rule_order` on PUT, silently moving block rules to the bottom of the policy — below allow rules.

## Fix

### Template reconciliation — private state

Terraform resource **private state** is an opaque provider-owned key/value blob stored inside the state file. It survives across processes and backends. The fix records the last display name sent to the API under the key `npa_template_display` in `ModifyPlan`. The plan modifier (`suppressTemplateDrift`) uses it to decide whether the `.html` filename in state corresponds to the configured display name:

- **Branch 1 (nothing to reconcile):** config or state empty, equal, config already a `.html` filename, or state not a filename → leave the plan alone.
- **Branch 2 (steady state):** Private state matches config → suppress the spurious diff. If another attribute also changed the plan, restore the display name so the PUT carries an accepted value.
- **Branch 3 (first apply after import/upgrade/create):** Private state absent or outdated → show an honest diff. Applying it sends the display name and persists Private. The next plan takes branch 2.

The session caches (`npaTemplateCache`, `npaTemplatesAPICache`) and every call to `/api/v2/templates/usernotifications` are removed; the provider no longer depends on that endpoint (guarded by `TestNPARules_NoTemplatesEndpointCall`). The template value from config is sent unchanged for every action type. For `block` rules the API returns a `.html` filename on GET, which state stores and the plan modifier reconciles; for `periodic_reauth` rules the API stores and returns the display name verbatim and rejects filenames (verified bespin + policydemo, 2026-10-21, `docs/api-audit/npa-rules-template.md`).

### `rule_order` strip on update

`hookMyPolicyBeforeRequest.go` sets `requestMap.RuleOrder = nil` when `OperationID == "updateNPARules"` **and the body contains `rule_data`**. The presence of `rule_data` distinguishes a full-rule update (strip) from the order-only PATCH that `netskope_npa_rules_order` sends (`{"rule_order": {...}}` with no `rule_data`), which must keep `rule_order` or reordering becomes a no-op. `rule_order` on `netskope_npa_rules` is therefore create-time placement only.

## Known limitations

1. An out-of-band UI template change (e.g. `5.html` → `7.html`) is not detected while the configured display name is unchanged. Branch 2 suppresses it.
2. One visible `"<file>.html" → "<display name>"` diff and PUT per rule after: upgrade from ≤0.4.12, `terraform import`, or first creation (Option A — the generated `Create` does not call `resp.Private.SetKey`).
3. `rule_order` is now create-time only. Changing it in config after creation produces a plan diff that applies without effect; use `netskope_npa_rules_order` for post-creation ordering.

## Tests

- Unit: `TestTemplatePlanDecision`, `TestSuppressTemplateDrift_*` (`nparules_resource_planmodify_test.go`); `TestBeforeRequest_UpdateStripsRuleOrder`, `_OrderOnlyUpdatePreservesRuleOrder`, `_CreatePreservesRuleOrder`; `TestAfterSuccess_TemplateFilenameNotRewritten_*`; `TestNPARules_NoTemplatesEndpointCall`.
- Acceptance (bespin, 403 tenant): `TestAccNPARules_blockRuleUpdate`, `_blockRuleTemplateChange`, `_importBlockRule`, `_denyRule`, `_updateWithIgnoredRuleOrder`, `_updateBlockRuleKeepsPosition`, `TestAccDrift_NPARules_BlockTemplate`, `TestAccNPARulesOrder_detectsOutOfBandReorder`.

## Files changed

- `internal/provider/nparules_resource_planmodify.go` — new `suppressTemplateDrift` using private state
- `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go` — rule_order strip on full-rule updates; template passed through unchanged for all action types
- `internal/sdk/internal/hooks/hookMyPolicyAfterSuccess.go` — removed template substitution block
- `internal/sdk/internal/hooks/hookNPATemplatesAPI.go` — deleted (templates endpoint no longer used)
- `internal/sdk/internal/hooks/hookNPATemplateCache.go` — deleted
- `internal/sdk/internal/hooks/hookPolicyTemplateAfterSuccess.go` — deleted
