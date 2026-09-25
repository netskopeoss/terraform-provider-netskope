# Onboarding — Netskope Terraform Provider

**Audience:** Experienced Terraform provider developers joining this project.  
**Scope:** Everything you need to get a working development environment and understand how changes flow from API spec to shipped provider.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Repositories](#2-repositories)
3. [OAS Files — Location and Role](#3-oas-files--location-and-role)
4. [Speakeasy Setup](#4-speakeasy-setup)
5. [Local Development Workflow](#5-local-development-workflow)
6. [Environment Variables](#6-environment-variables)
7. [Running Tests](#7-running-tests)
8. [GitHub Actions](#8-github-actions)
9. [Claude Code and Available Skills](#9-claude-code-and-available-skills)
10. [Writing Hook Unit Tests](#10-writing-hook-unit-tests)
11. [Key Documentation](#11-key-documentation)
12. [Golden Rules](#12-golden-rules)

---

## 1. Prerequisites

| Tool | Version | Notes |
|---|---|---|
| Go | 1.21+ | Check `go.mod` for the exact minimum |
| Terraform | ~1.x | Used by acceptance tests; any recent 1.x is fine |
| Speakeasy CLI | see `.speakeasy/workflow.yaml` `speakeasyVersion` | Install from speakeasy.bar or via `brew install speakeasy-api/homebrew-tap/speakeasy` |
| Git | any | — |
| Claude Code | latest | Optional but the skills save significant time — see Section 9 |

Install Speakeasy and authenticate:

```bash
brew install speakeasy-api/homebrew-tap/speakeasy   # macOS
speakeasy auth login                                 # authenticates to registry.speakeasyapi.dev
```

---

## 2. Repositories

This project spans **two repositories**. Both must be cloned:

```
~/PycharmProjects/terraform-provider-netskope/    ← this repo (provider implementation)
~/speakeasy/netskope-apiv2-oas/                   ← OAS files (API specifications)
```

Clone them:

```bash
git clone git@github.com:netskopeoss/terraform-provider-netskope.git \
    ~/PycharmProjects/terraform-provider-netskope

git clone git@github.com:netskopeoss/netskope-apiv2-oas.git \
    ~/speakeasy/netskope-apiv2-oas
```

The provider repo's `.speakeasy/workflow.yaml` references the OAS repo via the environment variable `$NETSKOPE_OAS_DIR`. Set this in your shell profile:

```bash
export NETSKOPE_OAS_DIR="$HOME/speakeasy/netskope-apiv2-oas"
```

---

## 3. OAS Files — Location and Role

The provider is **fully generated** by Speakeasy from OpenAPI Specification (OAS) YAML files. You rarely edit Go directly.

### OAS repo structure

```
$NETSKOPE_OAS_DIR/
├── base_oas.yaml                          ← shared info block, servers, security schemes
├── terraform_overlay.yaml                 ← Speakeasy x-speakeasy-* annotation overrides
└── endpoints/
    ├── steering/
    │   ├── npa_apps_private.yaml          ← private apps
    │   ├── ipsec_tunnels.yaml
    │   └── gre_tunnels.yaml
    ├── policy/
    │   ├── npa_policy.yaml                ← NPA rules
    │   ├── npa_policygroup.yaml
    │   └── urllist.yaml
    ├── infrastructure/
    │   ├── npa_publishers.yaml
    │   ├── npa_upgrade_profiles.yaml
    │   └── lbrokers.yaml
    ├── profiles/
    │   ├── dns_profiles_v2.yaml
    │   ├── destination_profiles.yaml
    │   ├── custom_categories.yaml
    │   └── service_objects.yaml
    ├── platform/
    │   ├── rbac_labels.yaml
    │   ├── rbac_roles.yaml
    │   ├── deviceclassification_tags.yaml
    │   ├── deviceclassification_rules.yaml
    │   ├── deviceclassification_onpremdetection.yaml
    │   ├── device_tags.yaml
    │   └── cci_data.yaml
    └── aig/
        ├── aig-appliances-openapi.yaml
        ├── aig-aiproviders-openapi.yaml
        ├── aig-mcpservers-openapi.yaml
        ├── aig-ratelimits-openapi.yaml
        └── aig-tokens-openapi.yaml
```

### Key Speakeasy annotations on OAS fields

| Annotation | Effect |
|---|---|
| `x-speakeasy-terraform-ignore: true` | Field excluded from Terraform schema (invisible to users) |
| `x-speakeasy-ignore: true` | Field excluded from SDK Go struct (JSON deserializer skips it) |
| `x-speakeasy-param-suppress-computed-diff: true` | Suppresses drift for computed fields that the API returns but users don't set |
| `x-speakeasy-name-override: foo` | Renames the field in the generated Go struct |

When the OAS has a `$ref` to a shared schema, add the annotation both at the reference site **and** on the referenced schema — Speakeasy requires both.

---

## 4. Speakeasy Setup

### `.speakeasy/workflow.yaml`

This file in the provider repo defines the generation pipeline. It lists every OAS input file, the overlay file, and the Speakeasy registry target. The `$NETSKOPE_OAS_DIR` variable is expanded at runtime.

### Running generation

```bash
cd ~/PycharmProjects/terraform-provider-netskope
speakeasy run
```

This pulls all OAS inputs, merges them with the overlay, and regenerates:
- `internal/sdk/` — Go SDK (models, operations, HTTP client)
- `internal/provider/` — Terraform resource and data source implementations
- `docs/` — provider documentation pages
- `examples/` — example HCL configurations

**`speakeasy run` is not scoped.** Changing one OAS file regenerates *everything*. Always `git diff` the full output before committing.

### What `.genignore` protects

The `.genignore` file tells Speakeasy to skip certain files entirely during regeneration. These files contain manual edits that would be destroyed if regenerated:

| Protected file | Why it is genignored |
|---|---|
| `internal/provider/provider.go` | Manually registers resources/data sources that Speakeasy doesn't know about (e.g. `NewNPARulesOrderResource`, device tag resources). Speakeasy would drop them. |
| `internal/sdk/internal/utils/retries.go` | Two bugs in the generated retry logic were fixed manually: (1) `Retry-After` header values are now capped to `MaxInterval` so a large server-supplied value can't stall the provider; (2) the elapsed-budget check runs before sleeping, not after, so the provider fails fast when time runs out. The comment block at the top of that file documents both fixes. |
| `internal/provider/*_resource_planmodify.go` | Custom plan modifier implementations for private app, NPA rules, IPSec, and GRE tunnels. These suppress false diffs that the generated code cannot handle. |
| `docs/index.md` | The Speakeasy-generated version number in this file would be wrong; we maintain it manually. |
| `internal/sdk/internal/hooks/registration.go` | Hook registration is entirely manual. |
| All `docs/data-sources/*.md` and `docs/resources/*.md` | Most doc pages are genignored to allow custom descriptions and examples. |

### Files Speakeasy DOES overwrite — revert these after every run

These files are **not** in `.genignore`, so Speakeasy overwrites them. Each has caused production regressions in the past. Revert them immediately after `speakeasy run` before doing anything else:

---

#### `internal/sdk/internal/utils/json.go`

**What Speakeasy does:** Rewrites the custom JSON marshaller used by every SDK request and response.

**Why it breaks things:** This file controls how Go structs are serialised to JSON. The critical behaviour is how it handles **empty slices**. When Speakeasy upgrades its internal version of this file, it sometimes changes the `omitEmpty` logic to omit empty arrays (`[]`) from the JSON output.

The GRE tunnel resource has an `xff.iplist` field (`XffIPList []string`, tagged `json:"iplist,omitempty"`) that holds X-Forwarded-For IP addresses. When a user removes all IPs, the provider needs to send `"iplist": []` in the PUT request so the API clears the list. If the marshaller omits empty slices, the field is silently dropped from the JSON and the API leaves the old list in place — the update appears to succeed but the state drifts immediately on the next plan.

**Revert command:**
```bash
git checkout -- internal/sdk/internal/utils/json.go
```

After reverting, confirm the existing test still passes:
```bash
go test -v ./internal/sdk/internal/utils/...
```

---

#### `go.mod` and `go.sum`

**What Speakeasy does:** Bumps Go module dependencies — sometimes to versions that introduce breaking interface changes in `terraform-plugin-go`.

**Why it breaks things:** The Terraform Plugin Framework occasionally makes breaking interface changes between minor versions. A dependency bump that passes `go build` may still break acceptance tests or produce subtle runtime panics. Dependency upgrades should be a deliberate, tested decision — not a side effect of running `speakeasy run`.

**Revert command:**
```bash
git checkout -- go.mod go.sum
```

If you intentionally need a dependency bump, do it separately with `go get` and a dedicated commit so the change is visible in the PR diff.

---

#### `examples/provider/provider.tf`

**What Speakeasy does:** Updates the provider version number in the example configuration to its own internal counter, which diverges from our actual release version.

**Why it breaks things:** Users copy this file from the registry docs. If it references a version that doesn't exist on the registry yet (or that we haven't tagged), `terraform init` fails. The version in this file must match our actual release tag.

**Revert command:**
```bash
git checkout -- examples/provider/provider.tf
```

---

### The revert one-liner

Run this immediately after every `speakeasy run` before committing anything:

```bash
git checkout -- \
    go.mod \
    go.sum \
    internal/sdk/internal/utils/json.go \
    examples/provider/provider.tf
```

Then validate and diff the remaining changes:

```bash
git diff --stat                        # see all changed files
go build ./...                         # confirm it compiles
make test                              # confirm unit tests pass
go test -v ./internal/sdk/internal/hooks/  # hook-specific unit tests
```

Also watch `internal/sdk/internal/utils/utils.go` — it is not in `.genignore` and is not always reverted, but serialization changes here affect every resource. If Speakeasy modified it, read the diff carefully before deciding whether to revert.

#### If new resources or data sources were added

Run `tfplugindocs` to generate the Terraform Registry doc pages. Without this, the registry has no documentation for new resources even if `examples/` are present in GitHub:

```bash
tfplugindocs generate
git add docs/
```

Check the output for "generating new template" lines — each one is a doc page that was missing.

---

## 5. Local Development Workflow

### Editing OAS vs editing Go

```
OAS file → speakeasy run → generated Go
                              ↑
                    DO NOT edit this directly
```

The only Go you write by hand is in `internal/sdk/internal/hooks/`. Everything else is overwritten on the next `speakeasy run`.

**Exception:** if a file absolutely cannot be generated correctly, add it to `.genignore`. This is a last resort — prefer fixing the OAS or adding a hook.

### Fix path decision tree

```
Bug reported
    │
    ├─ Is this a field type / annotation problem?
    │      → Fix the OAS, regenerate
    │
    ├─ Does the API return wrong data shape / need request transformation?
    │      → Write or update a hook in internal/sdk/internal/hooks/
    │
    ├─ Is this perpetual drift from list ordering?
    │      → Sort in an AfterSuccess hook  (not a plan modifier — see DEVELOPER_GUIDE.md §6)
    │
    └─ Is this computed field drift (API returns a value, user doesn't set it)?
           → x-speakeasy-param-suppress-computed-diff: true on the OAS response schema
```

### Building and installing locally

```bash
# Quick compile check
go build ./...

# Build a macOS binary
make build-darwin          # outputs to bin/mac/

# Install to $GOPATH/bin for local Terraform use
make install

# Use a dev_overrides block in ~/.terraformrc to point terraform at your local binary:
# provider_installation {
#   dev_overrides {
#     "registry.terraform.io/netskopeoss/netskope" = "/path/to/your/GOPATH/bin"
#   }
# }
```

---

## 6. Environment Variables

### Required for acceptance tests

| Variable | Example | Notes |
|---|---|---|
| `TF_ACC` | `1` | Enables acceptance tests (required) |
| `NETSKOPE_SERVER_URL` | `https://alliances.goskope.com/api/v2` | **Must include `/api/v2` suffix.** Without it all requests 404. |
| `NETSKOPE_API_KEY` | `<base64-encoded key>` | API key from your tenant |

### Optional / feature-specific

| Variable | When needed |
|---|---|
| `NETSKOPE_OAUTH2_CLIENT_ID` | OAuth2 provider auth tests (`testacc-oauth2-provider`) |
| `NETSKOPE_OAUTH2_CLIENT_SECRET` | Same |
| `NETSKOPE_TEST_USER` | Valid user email on the tenant — enables `TestAccNPARules_allNewRuleDataFields` (skips if unset) |
| `TF_RUN_DESTINATION_PROFILES` | Set to `1` to run destination profile tests (licensed feature, policydemo only) |
| `TF_SKIP_TIME_INTERVAL` | Set to `1` when running against policydemo (no time interval ID 3) |
| `NETSKOPE_OAS_DIR` | Path to the OAS repo — required for `speakeasy run` |

### Credential files

All credentials live in `.env` at the repo root (gitignored). Source it before running tests:

```bash
source .env
```

Policydemo credentials (for destination profile tests) are in a separate file in the internal repo:
```bash
source ../terraform-provider-netskope-internal/.env
# then override: NETSKOPE_SERVER_URL="$POLICY_DEMO_SERVER_URL" NETSKOPE_API_KEY="$POLICY_DEMO_API_KEY"
```

Ask `jharris@netskope.com` for the credential files — they are not checked in anywhere.

---

## 7. Running Tests

### Unit tests (no API credentials)

```bash
make test               # or: go test -v ./... -timeout 10m
```

These run in CI on every push and PR. They include hook unit tests and plan modifier logic tests.

### Acceptance tests — two-tenant strategy

The test suite requires two tenants for full coverage. Use **bespin first** (covers ~90% of tests), then **policydemo** for destination profiles.

#### bespin (`bespin.goskope.com`) — primary tenant

```bash
source .env    # sets NETSKOPE_SERVER_URL, NETSKOPE_API_KEY, NETSKOPE_TEST_USER
TF_ACC=1 go test -v ./internal/provider/... -timeout 120m -parallel 8
```

Expected: ~207 PASS, ~11 SKIP (IPS status, destination profiles, a few known skips), 0 FAIL.

#### policydemo (`policydemo.goskope.com`) — destination profiles only

```bash
source ../terraform-provider-netskope-internal/.env
NETSKOPE_SERVER_URL="$POLICY_DEMO_SERVER_URL" \
NETSKOPE_API_KEY="$POLICY_DEMO_API_KEY" \
TF_RUN_DESTINATION_PROFILES=1 \
TF_SKIP_TIME_INTERVAL=1 \
TF_ACC=1 go test -v ./internal/provider/... \
  -run "TestAccDestinationProfile|TestAccDrift_Destination" \
  -timeout 30m -parallel 4
```

Use `-parallel 4` on policydemo — the load balancer drops connections under higher concurrency.

> **Do not run the full suite against policydemo.** Tests requiring bespin-only resources (block template, network location ID 1, time interval ID 3, AIG, DNS Profile V2) will hard-fail, not skip.

#### By resource area

```bash
make testacc-rules
make testacc-privateapp
make testacc-publisher
make testacc-policygroups
make testacc-datasources
make testacc-deviceclassification
make testacc-rbaclabels
make testacc-aig
make testacc-customcategory
make testacc-serviceobject

# Single test
TF_ACC=1 go test -v ./internal/provider/... -run TestAccNPARules_basic -timeout 30m
```

**Run acceptance tests locally before every PR.** CI runs against bespin with `-parallel 8` — match that locally before pushing.

#### Cleanup after failed tests

Test resources are prefixed `tf-acc-test`. If tests fail mid-run, clean up with the sweeper:

```bash
source .env
TF_ACC=1 go test -v ./internal/provider/... -sweep=bespin.goskope.com -sweep-run="netskope_*"
```

---

## 8. GitHub Actions

The CI configuration is in `.github/workflows/test.yml`. It runs on:
- Every push to `main` or any release branch (`X.Y.Z*`)
- Every pull request targeting `main`
- Manual dispatch (`workflow_dispatch`)

### Jobs

#### `gitleaks` — Secret Scan

Runs on every push and PR. Scans commits since `origin/main` with [gitleaks](https://github.com/gitleaks/gitleaks). Any leaked credential fails the workflow immediately. This is the first gate — it runs before unit or acceptance tests.

#### `unit` — Unit Tests

Runs `go test -v ./... -timeout 10m`. No credentials required. This catches compilation errors, hook logic bugs, and plan modifier regressions.

#### `acceptance` — Acceptance Tests

Runs after `unit`. Condition:
- **Always runs** on pushes to `main` or release branches
- **Runs on PRs from this repo** (not forks — fork PRs cannot access repo secrets)
- **Manual dispatch** always runs it

Uses `go test -v ./internal/provider/... -timeout 120m -parallel 8`.

Secrets injected by CI:

| Secret | Variable |
|---|---|
| `NETSKOPE_API_KEY` | `NETSKOPE_API_KEY` |
| `NETSKOPE_SERVER_URL` | `NETSKOPE_SERVER_URL` |
| `NETSKOPE_OAUTH2_CLIENT_ID` | `NETSKOPE_OAUTH2_CLIENT_ID` |
| `NETSKOPE_OAUTH2_CLIENT_SECRET` | `NETSKOPE_OAUTH2_CLIENT_SECRET` |
| `NETSKOPE_TEST_USER` | `NETSKOPE_TEST_USER` |
| `TF_RUN_DESTINATION_PROFILES` | `TF_RUN_DESTINATION_PROFILES` |

> **Tip:** Because forks cannot access secrets, acceptance tests only run on PRs from branches in this repo. This means you must push your branch to `netskopeoss/terraform-provider-netskope` (not a fork) to get acceptance test results in CI.

---

## 9. Claude Code and Available Skills

This project uses [Claude Code](https://claude.ai/code) with a set of custom skills that automate common workflows. Install Claude Code and run it from the provider repo root.

### Available skills (invoke with `/skill-name`)

| Skill | Invoke | What it does |
|---|---|---|
| `new-endpoint` | `/new-endpoint` | 7-phase workflow for adding a new resource or data source: (1) API audit via MCP + curl, (2) OAS annotation, (3) regeneration via `/speakeasy`, (4) hook implementation if needed, (5) provider registration, (6) acceptance tests (5 required test types), (7) `tfplugindocs` + doc updates. Use this any time a new endpoint is being added. |
| `speakeasy` | `/speakeasy` | Runs `speakeasy run`, reverts the known-bad files (`json.go`, `go.mod`, `go.sum`, `examples/provider/provider.tf`), builds, runs hook unit tests, and optionally runs `tfplugindocs`. Use any time OAS files change. |
| `pr` | `/pr` | Pre-PR validation + PR creation: checks for never-commit files, verifies CHANGELOG date, confirms new resources are registered in `provider.go`, checks drift tests and `ImportStateVerifyIgnore` comments exist, runs acceptance tests, then creates the PR with a structured description. |
| `bugfix` | `/bugfix` | Takes a bug number (`BUG-NNN`) or GitHub issue URL. Investigates root cause, chooses the right fix approach (OAS → hook → genignore in priority order), implements the fix, adds hook unit tests and acceptance tests, updates the bug file, CHANGELOG, and any affected docs. |
| `testacc` | `/testacc` | Runs acceptance tests against the right tenant. Knows the two-tenant strategy (bespin for ~90% of tests at `-parallel 8`; policydemo at `-parallel 4` for destination profiles only), all skip flags, which tests hard-fail on the wrong tenant, and how to run the sweeper for cleanup. |
| `publish` | `/publish [version]` | Tags and publishes a release: verifies CHANGELOG entry, checks version consistency across files, runs `tfplugindocs`, creates an annotated git tag, and pushes it to trigger the registry publish. |
| `review` | `/review` | Reviews a pull request for correctness, drift risk, and adherence to project conventions. |
| `security-review` | `/security-review` | Security-focused review of pending changes on the current branch. |
| `commit-message` | `/commit-message` | Generates a semantic commit message from staged changes. |

### MCP servers

The project has two Netskope MCP servers configured (`netskope` and `netskope-bespin`) that connect to real tenant APIs. The `new-endpoint` skill uses these to test actual API behavior before annotating the OAS — critical for catching field type mismatches and undocumented computed fields.

---

## 10. Writing Hook Unit Tests

Every hook you write or modify needs a unit test. Hook unit tests live alongside the hook in `internal/sdk/internal/hooks/` and run as part of the standard unit suite (`make test`) — no API credentials needed.

Full guidance, patterns, and worked examples are in **[`docs/HOOK_UNIT_TEST_GUIDE.md`](HOOK_UNIT_TEST_GUIDE.md)**. Key points:

- **Mock the HTTP layer directly** — create a fake `*http.Response` with a JSON body that matches what the real API returns. No mocking frameworks; plain `io.NopCloser(strings.NewReader(...))`.
- **Always test the passthrough case** — call the hook with a non-matching `OperationID` and verify it returns the response unchanged.
- **Test all operation IDs the hook handles** — use `t.Run(opID, ...)` to loop over them.
- **Test edge cases** — empty lists, nil field values, and missing fields should not panic.
- **Test idempotency** — running the hook twice on the same input must produce identical output.
- **JSON numbers unmarshal as `float64`** — not `int`. If the API returns `"publisher_id": 256`, your test data must use `float64(256)`.

Run hook tests in isolation:

```bash
go test -v ./internal/sdk/internal/hooks/
go test -v ./internal/sdk/internal/hooks/ -run TestPublishersSortedByID
go test -v ./internal/sdk/internal/hooks/ -cover   # show coverage
```

---

## 11. Key Documentation

Read these before touching the code. The suggested order is reflected in the table:

| Document | Path | Read when |
|---|---|---|
| **This guide** | `docs/ONBOARDING.md` | Now (environment setup) |
| Developer Guide | `docs/DEVELOPER_GUIDE.md` | Next — architecture, hook mechanics, drift root causes, and a full worked example (BUG-001) |
| Known API Issues | `docs/KNOWN_API_ISSUES.md` | Before adding a new endpoint or debugging unexpected behaviour — 17 documented API quirks with their workarounds |
| Hook Unit Test Guide | `docs/HOOK_UNIT_TEST_GUIDE.md` | Before writing your first hook test — Go test fundamentals, mock HTTP response patterns, table-driven tests |
| Running Tests | `docs/RUNNING_TESTS.md` | Quick reference for test flags, patterns, and troubleshooting |
| Acceptance Tests | `docs/ACCEPTANCE_TESTS.md` | When adding or modifying tests — lists every test, what it covers, and required env vars |
| Bug reports | `docs/bugs/BUG-*.md` | When a symptom matches a known pattern — each file has root cause, fix, and which files were touched |
| TODO | `docs/TODO.md` | Current backlog |

---

## 12. Golden Rules

1. **Don't edit generated Go.** Fix the OAS or write a hook. Files outside `internal/sdk/internal/hooks/` are regenerated on every `speakeasy run`.

2. **Audit before annotating.** New OAS files from the API team are often inaccurate. Use the `new-endpoint` skill (or the MCP tools) to test the real API before adding `x-speakeasy-*` annotations.

3. **Revert the known-bad files** after every `speakeasy run`. Run `/speakeasy` to do this automatically.

4. **Run acceptance tests locally before PRs.** CI acceptance tests run on bespin at `-parallel 8` — match that locally. Broken tests block everyone on the shared tenant.

5. **Every new endpoint needs acceptance tests.** No exceptions. The `new-endpoint` skill scaffolds them.

6. **No AI attribution in commits.** Don't include "Claude" or "Co-Authored-By" lines in commit messages.

7. **NETSKOPE_SERVER_URL must end with `/api/v2`.** Without it every acceptance test request returns 404.

8. **Don't commit:** `*.tfvars`, `planning_docs/`, `CLAUDE.md`, `*.test`, `.env*`