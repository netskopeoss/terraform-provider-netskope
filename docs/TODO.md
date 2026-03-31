# TODO

## Pending

### Add unit tests for remaining hooks

**Priority:** Medium
**Target:** 0.4.0
**Added:** 2026-02-08

Only 2 of 7 active hooks have unit tests (29% coverage). The following hooks need tests:

| Hook | Priority | What it does |
|------|----------|--------------|
| `hookErrorStatusResponse.go` | High | Handles API 200-with-error responses, "not found" pattern matching |
| `hookPrivateAppRequest.go` | High | Removes problematic fields from PUT requests to avoid API errors |
| `hookMyPolicyAfterSuccess.go` | Medium | Trims brackets from privateApps, nil checks |
| `hookMyBulkPolicyAfterSuccess.go` | Medium | Same as above for bulk endpoints |
| `hookMyPolicyBeforeRequest.go` | Medium | Policy request transformations |

**Key test cases needed:**

1. **hookErrorStatusResponse:**
   - Non-200 responses passthrough unchanged
   - JSON parse failure → passthrough
   - `status: "error"` → returns Go error
   - "not found" patterns → returns `ErrResourceNotFound`
   - `status: "success"` → passthrough

2. **hookPrivateAppRequest:**
   - Non-update operations passthrough
   - Removes `app_option` when `clientless_access=false`
   - Removes empty `paths`, `app_option`, `uribypass_header_value`

3. **hookMyPolicyAfterSuccess:**
   - `RuleData` nil → passthrough
   - `PrivateApps` nil → passthrough
   - Bracket trimming: `[app]` → `app`

**Reference:** See `hookPublisherSort_test.go` for test patterns.

---

### Migrate tests to `testdata/` + `package provider_test`

**Priority:** High
**Target:** 0.4.0
**Added:** 2026-02-13 (replaces previous DRY refactoring item from 2026-02-08)

Migrate all acceptance tests to the Speakeasy-recommended pattern: `package provider_test` (black-box) with HCL configs in `testdata/` directories. This aligns with Speakeasy's intended direction for auto-generated tests and provides IDE support for HCL configs.

**Current state (31 test files, ~5,500 lines):**
- `package provider` (white-box testing)
- HCL configs embedded in Go via `fmt.Sprintf`
- ~56 config builder functions with ~300 lines of duplication
- Common dependency patterns (publisher, private app, policy group) copy-pasted across 11+ files

**Target state:**
- `package provider_test` (black-box testing)
- HCL configs in `testdata/{TestFunctionName}/main.tf`
- `ConfigDirectory` + `ConfigVariables` in test steps
- Shared HCL modules or variable-driven configs for common dependencies
- Export test helpers: `TestAccPreCheck`, `TestAccProviderConfig`, etc.

**Migration approach:**
1. Export shared test helpers (rename to uppercase)
2. Switch `package provider` → `package provider_test` in all test files
3. Extract HCL configs to `testdata/` directories one resource at a time
4. Replace `Config:` with `ConfigDirectory:` + `ConfigVariables:`
5. Verify all tests pass after each resource migration

**Key constraint:** `ConfigDirectory` cannot use `ExternalProviders` — must define `required_providers` in the HCL files.

**Files affected:** All `internal/provider/*_test.go` files (31 test files)

**References:**
- [Speakeasy testing docs](https://www.speakeasy.com/docs/terraform/customize/testing)
- [HashiCorp acceptance test configuration](https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests/configuration)

---

### Improve Terraform Registry documentation

**Priority:** Medium
**Target:** 0.4.0
**Added:** 2026-02-08

The Terraform Registry documentation lacks examples, best practices, and guidance on HCL structure.

**Current issues:**

1. **Lack of examples** — Most resource docs have placeholder values like `"..."`
2. **No guides** — Missing `docs/guides/` for tutorials and best practices
3. **No protocol ordering guidance** — Users don't know protocols must be sorted (tcp before udp, port ascending)
4. **No publisher dependency guidance** — Private apps require publishers but this isn't clear
5. **Generated docs get overwritten** — Only `npa_private_app.md` has good custom content

**Protected docs (added to .genignore):**
- `docs/index.md`
- `docs/resources/npa_private_app.md`

**Recommended improvements:**

1. **Add `docs/guides/` directory** for tutorials:
   ```
   docs/guides/
   ├── getting-started.md           # Quick start tutorial
   ├── private-app-best-practices.md    # Protocol ordering, publisher deps
   ├── authentication.md            # Detailed auth setup
   └── troubleshooting.md           # Common issues
   ```

2. **Improve key resource docs** and add to `.genignore`:
   - `npa_publisher.md` — Add realistic examples
   - `npa_rules.md` — Document rule_data structure
   - `npa_policy_groups.md` — Document group_order

3. **Document known quirks:**
   - Protocol ordering (tcp before udp, port ascending)
   - Publisher names with whitespace
   - List attribute ordering sensitivity

4. **Expand restore-docs.sh** to restore all custom docs from templates

**Reference:**
- [HashiCorp provider docs guidelines](https://developer.hashicorp.com/terraform/registry/providers/docs)
- Current good example: `docs/resources/npa_private_app.md`

---

### Add write-only fields to `netskope_npa_private_app`

**Priority:** Medium
**Target:** 0.4.0
**Added:** 2026-02-13

Four fields on the private app resource are accepted by the API on create/update but not returned in GET responses. The API returns `null`, empty objects, or omits them entirely. This prevents Terraform from reconciling state, causing a perpetual plan diff on every run. The fields are currently excluded from the provider schema for this reason.

**Fields to add:**

| Field | Type | Notes |
|-------|------|-------|
| `allow_uri_bypass` | `bool` | Not returned in GET |
| `bypass_uris` | `list(string)` | Not returned in GET |
| `uribypass_header_value` | `string` | Not returned in GET; empty values cause API serialization bug (Issue #8) |
| `app_option` | `object` | Only valid when `clientless_access = true`; not returned in GET; empty values cause API serialization bug (Issue #8) |

**Solution:** Terraform 1.11 introduced write-only arguments (`WriteOnly: true`) — attributes that are sent to the API but never persisted to plan or state. This avoids the perpetual diff because no state reconciliation occurs for these fields. The OAS annotation `x-speakeasy-terraform-write-only: true` generates the correct schema.

**Breaking change:** Requires minimum Terraform version bump from 1.0 to 1.11. This is acceptable in a 0.4.0 release but was not appropriate for 0.3.x, where existing users on older Terraform versions would be unable to upgrade the provider without also upgrading Terraform.

**Existing mitigations to keep:**
- `hookPrivateAppRequest.go` strips empty `app_option`, `paths`, and `uribypass_header_value` from PUT requests to avoid the API serialization bug (Issue #8)
- `hookPrivateAppRequest.go` removes `app_option` entirely when `clientless_access = false` (API rejects it for non-clientless apps)

**References:**
- [Terraform write-only arguments](https://developer.hashicorp.com/terraform/plugin/framework/resources/write-only-arguments)
- [Known API Issues #8 and #10](KNOWN_API_ISSUES.md)

---

### Consider debug strategy for hooks

**Priority:** Low
**Added:** 2026-02-05

Review the current debug flag approach in hooks (`myAppResponseDebug`, `myBulkAppResponseDebug`, etc.) and consider a more general debug strategy.

**Options to evaluate:**

1. **Runtime configuration** — Read from environment variable (e.g., `NETSKOPE_DEBUG_HOOKS=true`) instead of compile-time constants
2. **Centralized debug flag** — Single flag that enables all hook debugging instead of per-hook flags
3. **Log levels** — Integrate with a proper logging framework with log levels (DEBUG, INFO, WARN, ERROR)
4. **Debug hook** — The existing `hookDebugRequest.go` logs to `/tmp/netskope_debug.log` but is commented out in `registration.go`

**Current state:**
- 5+ separate debug flags scattered across hook files
- All hardcoded as `bool = false`
- Require code change and rebuild to enable
- Added by Justin Adrian during development (Feb 2025)

**Files involved:**
- `internal/sdk/internal/hooks/hookMyAppAfterSuccess.go` — `myAppResponseDebug`
- `internal/sdk/internal/hooks/hookMyBulkAppAfterSuccess.go` — `myBulkAppResponseDebug`
- `internal/sdk/internal/hooks/hookMyPolicyAfterSuccess.go`
- `internal/sdk/internal/hooks/hookMyBulkPolicyAfterSuccess.go` — `myBulkPolicyDebug`
- `internal/sdk/internal/hooks/hookMyPolicyBeforeRequest.go` — `myPolicyRequestDebug`
- `internal/sdk/internal/hooks/hookPrivateAppRequest.go` — `privateAppRequestDebug`
- `internal/sdk/internal/hooks/hookDebugRequest.go` — standalone debug hook (disabled)
- `internal/sdk/internal/hooks/registration.go` — hook registration

---

## Completed

### BUG-008: Concurrent rule creation fails with duplicate primary key

**Completed:** 2026-02-23
**Branch:** `0.3.6-beta`
**Issue:** [#66](https://github.com/netskopeoss/terraform-provider-netskope/issues/66)
**Docs:** [docs/bugs/BUG-008-rule-creation-race-condition.md](bugs/BUG-008-rule-creation-race-condition.md)

Added `hookRuleCreateSerializer` — a `sync.Mutex`-based hook that serializes concurrent `createNPARules` requests within a single provider process. Prevents the backend race condition where the API assigns duplicate rule IDs. Users no longer need `-parallelism=1` or `depends_on` workarounds.

**Files:** `internal/sdk/internal/hooks/hookRuleCreateSerializer.go`, `hookRuleCreateSerializer_test.go`, `registration.go`

---

### BUG-007: Make `protocols` required on `netskope_npa_private_app`

**Completed:** 2026-02-23
**Branch:** `0.3.6-beta`
**Docs:** [docs/bugs/BUG-007-clientless-empty-protocols.md](bugs/BUG-007-clientless-empty-protocols.md)

Marked `protocols` as `required` with `minItems: 1` in the OAS for both `private_apps_request` and `private_apps_put_request`. Regenerated with `speakeasy run`. Terraform now rejects configs without protocols at plan time (`Required: true`, `listvalidator.SizeAtLeast(1)`). Confirmed via live API testing that protocols are required for all app types.

**Files:** `endpoints/steering/npa_apps_private.yaml`, regenerated `npaprivateapp_resource.go`

---

### BUG-009: Rule creation fails after private app creation (eventual consistency)

**Completed:** 2026-02-23
**Branch:** `0.3.6-beta`
**Issue:** [#65](https://github.com/netskopeoss/terraform-provider-netskope/issues/65)
**Docs:** [docs/bugs/BUG-009-rule-after-app-eventual-consistency.md](bugs/BUG-009-rule-after-app-eventual-consistency.md)

Added `hookRuleCreateRetry` — an HTTP client wrapper (SDKInit hook) that retries rule creation requests when the API returns "doesn't exist" errors due to backend propagation delay. Uses exponential backoff (2s → 30s cap, 6 attempts, ~60s max). Users no longer need `time_sleep` workarounds.

**Files:** `internal/sdk/internal/hooks/hookRuleCreateRetry.go`, `hookRuleCreateRetry_test.go`, `registration.go`

---

### Add drift detection test for multi-publisher and tags scenarios

**Completed:** 2026-02-08
**Branch:** `0.3.4-beta`

Added `TestAccDrift_PrivateApp_MultiPublisherWithTags` acceptance test to verify BUG-001 fix works end-to-end. Test creates a private app with:
- 2 publishers
- 3 protocols (2 TCP, 1 UDP)
- 2 tags

This is the regression test for BUG-001 — verifies no perpetual drift with multiple list elements.

**File:** `internal/provider/drift_detection_test.go`

---

### BUG-001: List attribute perpetual diff on `netskope_npa_private_app`

**Completed:** 2026-02-05
**Branch:** `0.3.4-beta`
**Issue:** [#54](https://github.com/netskopeoss/terraform-provider-netskope/issues/54)
**Docs:** [docs/bugs/BUG-001-publishers-perpetual-diff.md](bugs/BUG-001-publishers-perpetual-diff.md)

Fixed perpetual plan drift on `publishers`, `protocols`, and `tags` attributes by sorting lists in AfterSuccess hooks and trimming whitespace from publisher names.
