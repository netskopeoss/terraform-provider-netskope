# BUG-015: IPSec Tunnels List Data Source Missing `pop_names`

**Resource:** `netskope_ip_sec_tunnels_list` (data source)
**Severity:** High (blocks backup/clone/restore workflows — `pop_names` is Required on the resource)
**Status:** Fixed (0.4.4) — manually added pop_names to generated files, added to .genignore
**Affected attributes:** `pop_names`
**Related:** BUG-014 (same issue on GRE tunnels list)

---

## Summary

Same pattern as BUG-014. The `netskope_ip_sec_tunnels_list` data source does not expose `pop_names`, despite it being **Required** on the `netskope_ip_sec_tunnel` resource and likely present in the API list response.

## Fix

Apply the same 3-file pattern as BUG-014:
1. `internal/provider/types/` — add `PopNames` to the list item type
2. `internal/provider/ipsectunnelslist_data_source.go` — add schema attribute
3. `internal/provider/ipsectunnelslist_data_source_sdk.go` — map from API response

## Impact

Without this fix, the `terraform-tenant-ops` module cannot implement full backup/clone/restore for IPSec tunnels.