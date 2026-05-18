# BUG-014: GRE Tunnels List Data Source Missing `pop_names` and `options`

**Resource:** `netskope_gre_tunnels_list` (data source)
**Severity:** High (blocks backup/clone/restore workflows — `pop_names` is Required on the resource)
**Status:** Fixed (0.4.4) — manually added pop_names + options to generated files, added to .genignore
**Affected attributes:** `pop_names`, `options`

---

## Summary

The `netskope_gre_tunnels_list` data source does not expose `pop_names` or `options` fields, despite both being present in the API response and writable on the `netskope_gre_tunnel` resource.

`pop_names` is **Required** on the resource, so any workflow that reads tunnels from the list data source and attempts to recreate them (clone, backup/restore) will fail because the required field is missing.

## Symptoms

```hcl
data "netskope_gre_tunnels_list" "all" {}

# This works — these fields are in the list DS
output "sites" {
  value = [for t in data.netskope_gre_tunnels_list.all.result : t.site]
}

# This fails — pop_names is not in the list DS schema
# output "pops" {
#   value = [for t in data.netskope_gre_tunnels_list.all.result : t.pop_names]
# }
```

## Root Cause

The list data source schema (`gretunnelslist_data_source.go`) and its type (`GreTunnelListItem` in `internal/provider/types/`) do not include `pop_names` or `options` fields.

However, the **API already returns this data**. The shared SDK model `GreTunnelListItem` in `internal/sdk/models/shared/gretunnellistitem.go` includes:

```go
Pops []GreTunnelPopStatusItem `json:"pops,omitempty"`
```

Each `GreTunnelPopStatusItem` has a `Name *string` field containing the POP name. The data is available — it's just not mapped through to the Terraform schema.

## Fix

Three files need changes:

### 1. `internal/provider/types/gre_tunnel_list_item.go`

Add `PopNames` field to the `GreTunnelListItem` struct:

```go
PopNames []types.String `tfsdk:"pop_names"`
```

### 2. `internal/provider/gretunnelslist_data_source.go`

Add schema attribute inside the `result` nested object:

```go
"pop_names": schema.ListAttribute{
    Computed:    true,
    ElementType: types.StringType,
    Description: "List of POP names this tunnel connects to",
},
```

### 3. `internal/provider/gretunnelslist_data_source_sdk.go`

In the refresh method, extract POP names from `resultItem.Pops` and map to `PopNames`. Follow the existing pattern in `gretunnel_resource_sdk.go` (around line 182-184) which already maps `PopNames` from the single-GET response.

```go
// Extract pop names from Pops status items
var popNames []types.String
for _, pop := range resultItem.Pops {
    if pop.Name != nil {
        popNames = append(popNames, types.StringValue(*pop.Name))
    }
}
result.PopNames = popNames
```

## Impact

Without this fix, the `terraform-tenant-ops` module cannot implement full backup/clone/restore for GRE tunnels. The `pop_names` field is Required on the resource — any attempt to create a tunnel without it fails validation.

## Notes

- Check if these files are in `.genignore` — if not, changes will be overwritten by Speakeasy regeneration.
- The `options` field (containing XFF configuration) has the same gap but is Computed+Optional, so it's lower priority. `pop_names` is the critical blocker.
- The same pattern (list DS missing fields present in single GET) may also affect `netskope_ip_sec_tunnels_list` — verify when implementing IPSec tunnel.