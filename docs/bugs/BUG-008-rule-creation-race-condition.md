# BUG-008: Concurrent Rule Creation Fails with Duplicate Primary Key

**Resource:** `netskope_npa_rules`
**Severity:** High (rule creation fails non-deterministically under default parallelism)
**Status:** Fixed (provider-side mitigation) — API-side issue remains open
**GitHub Issue:** #66
**Affected operations:** Create (concurrent)
**Fix branch:** `0.3.6-beta`

---

## Summary

When two or more `netskope_npa_rules` resources are created concurrently (Terraform's default parallelism), the API intermittently fails with a MySQL duplicate primary key error. The backend assigns the same `rule_id` to both concurrent inserts, causing one to fail. This is a server-side race condition in rule ID generation.

## Symptoms

```
netskope_npa_rules.rules["rule-A"]: Creating...
netskope_npa_rules.rules["rule-B"]: Creating...
netskope_npa_rules.rules["rule-A"]: Creation complete after 0s [id=2]

Error: failure to invoke API

  with netskope_npa_rules.rules["rule-B"],
  API error: (pymysql.err.IntegrityError) (1062, "Duplicate entry '2' for key 'PRIMARY'")
  [SQL: INSERT INTO inline_policies (... rule_id, rule_name, ...)]
  [parameters: {'rule_id': 2, 'rule_name': 'rule-B', ...}]
```

Both rules are assigned `rule_id=2`; the second insert fails on the unique constraint.

## Root Cause

The API backend does not use an auto-increment or atomic sequence for `rule_id` generation. Under concurrent inserts, two transactions read the same max ID and both attempt to insert with `max + 1`, causing a primary key collision. This is entirely API-side; the provider has no control over ID assignment.

## Fix

A `sync.Mutex`-based serialization hook ensures only one `createNPARules` request is in-flight at a time within a single provider process. The mutex is acquired in `BeforeRequest` and released in `AfterSuccess` or `AfterError`.

```
Goroutine A:  BeforeRequest (LOCK) → HTTP POST → AfterSuccess (UNLOCK)
Goroutine B:  BeforeRequest (BLOCKS)............→ (LOCK) → HTTP POST → AfterSuccess (UNLOCK)
```

The hook is registered first in the chain so the lock wraps all other hooks and the HTTP round-trip. This follows the same `sync.Mutex` pattern already used by `hookOAuth2Token.go`.

**Scope:** Only `createNPARules` is serialized. Updates, reads, and deletes are unaffected.

**Limitation:** Process-scoped only — serializes concurrent creates within a single `terraform apply`. Separate CLI invocations against the same tenant could still race (same limitation as `-parallelism=1`).

### Alternatives considered

| Approach | Verdict |
|----------|---------|
| `x-speakeasy-retries` (OAS extension) | Not viable — API returns 200 OK with error in body, not a retryable HTTP status code |
| Retry with backoff in AfterSuccess hook | Possible but fragile — requires error body parsing, non-deterministic recovery |
| Mutex serialization | Chosen — simple, deterministic, prevents the error entirely |

## Relevant Files

| File | Role |
|------|------|
| `internal/sdk/internal/hooks/hookRuleCreateSerializer.go` | Serialization hook implementation |
| `internal/sdk/internal/hooks/hookRuleCreateSerializer_test.go` | Unit tests (concurrency, error path unlock, passthrough) |
| `internal/sdk/internal/hooks/registration.go` | Hook registration (first in chain) |
| `internal/sdk/internal/hooks/hookOAuth2Token.go` | Precedent for `sync.Mutex` in hooks |

## User Workaround (pre-0.3.6)

Set parallelism to 1 when creating rules:

```bash
terraform apply -parallelism=1
```

Or use `depends_on` to serialise rule creation:

```hcl
resource "netskope_npa_rules" "rule_b" {
  depends_on = [netskope_npa_rules.rule_a]
  # ...
}
```

These workarounds are no longer needed with the 0.3.6 fix.

## Related

- Issue #65 (BUG-009) — may share the same underlying eventual consistency problem in the rules backend.
