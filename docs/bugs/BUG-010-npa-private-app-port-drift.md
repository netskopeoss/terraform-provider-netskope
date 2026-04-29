# Port Drift Test Results

**Date:** 2026-04-22
**Tenant:** bespin.goskope.com
**Provider:** netskopeoss/netskope v0.4.2
**Test config:** `test-drift/main.tf`

## Test Setup

Two `netskope_npa_private_app` resources created side-by-side:

| Resource | Description | Protocols sent |
|----------|-------------|---------------|
| `csv_no_fix` | CSV string passed as-is (current behavior) | `[{port = "80, 443", protocol = "tcp"}]` |
| `csv_with_fix` | CSV split into individual entries (proposed fix) | `[{port = "80", protocol = "tcp"}, {port = "443", protocol = "tcp"}, {port = "8080", protocol = "tcp"}, {port = "1025-5000", protocol = "udp"}, {port = "100-200", protocol = "tcp"}, {port = "300-400", protocol = "tcp"}]` |

## Finding 1: CSV Strings Cause Perpetual Drift (Confirmed)

On the very first `terraform apply` re-run after creation, `csv_no_fix` showed drift:

```
~ resource "netskope_npa_private_app" "csv_no_fix" {
    ~ protocols = [
        ~ {
            ~ port     = "80" -> "80, 443"
              # (1 unchanged attribute hidden)
          },
        - {
            - port     = "443" -> null
            - protocol = "tcp" -> null
          },
      ]
  }
```

**What happened:** The API accepted `"80, 443"` on create but returned it as two separate entries (`"80"` and `"443"`). Every subsequent plan tries to collapse them back into the CSV string -- this will never converge.

## Finding 2: API Reorders Protocol Entries (New Discovery)

The `csv_with_fix` resource had the correct number of split entries, but the API returned them in a **different order** than the config specified:

```
# Config order:                    # API returned order:
port = "80"        tcp             port = "100-200"   tcp
port = "443"       tcp             port = "300-400"   tcp
port = "8080"      tcp             port = "80"        tcp
port = "1025-5000" udp             port = "443"       tcp
port = "100-200"   tcp             port = "8080"      tcp
port = "300-400"   tcp             port = "1025-5000" udp
```

This means splitting CSV strings alone is **necessary but not sufficient**. The provider also reorders protocol entries on read-back, causing ordering drift even when the port values themselves match.

## Subsequent Plan Output (Post-Apply)

```
# csv_no_fix: perpetual drift (CSV re-collapse)
~ protocols = [
    ~ { ~ port = "80" -> "80, 443" },
    - { - port = "443" -> null, - protocol = "tcp" -> null },
  ]

# csv_with_fix: ordering drift
~ protocols = [
    ~ { ~ port = "80" -> "100-200" },
    ~ { ~ port = "443" -> "300-400" },
    ~ { ~ port = "8080" -> "80" },
    ~ { ~ port = "1025-5000" -> "443", ~ protocol = "udp" -> "tcp" },
    ~ { ~ port = "100-200" -> "8080" },
    ~ { ~ port = "300-400" -> "1025-5000", ~ protocol = "tcp" -> "udp" },
  ]
```

## Root Causes Summary

| Issue | Scope | Fix location |
|-------|-------|-------------|
| CSV port strings split by API on read-back | Terraform module code | Split CSV before passing to provider (`split(",", ...)` + `flatten`) |
| Protocol entry ordering not preserved by API | Netskope provider | Provider should sort or use `TypeSet` instead of `TypeList` for protocols |

## Conclusion

Two issues need to be addressed:

1. **Terraform-side (module fix):** Split CSV port strings into individual port/protocol entries before passing to the resource. This prevents the entry-count mismatch drift.

2. **Provider-side (provider fix):** The provider needs to handle protocol ordering so that the order returned by the API matches what Terraform expects. Options include:
    - Using `TypeSet` instead of `TypeList` for the protocols attribute (order-independent)
    - Sorting protocols in the provider's read function to produce a deterministic order
    - Implementing a `DiffSuppressFunc` that compares protocols as sets rather than lists

Without both fixes, drift will persist.

## Test Resources

These test resources (`Drift-Test-CSV-NoFix`, `Drift-Test-CSV-WithFix`) remain in the bespin tenant and should be destroyed after provider work is complete:

```bash
cd test-drift
source ../.env && export TF_VAR_netskope_api_key="$NETSKOPE_API_KEY"
terraform destroy
```
