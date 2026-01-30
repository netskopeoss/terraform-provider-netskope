# Migration Guide: v0.2.x to v0.3.x

## Overview

Version 0.3.x is a complete rewrite of the Netskope Terraform Provider. The provider was rebuilt using Speakeasy code generation from OpenAPI specifications, replacing the hand-written `terraform-plugin-sdk/v2` implementation with the Terraform Plugin Framework (ProtoV6).

Every resource and data source has been renamed, and most schemas have changed. Existing Terraform state from v0.2.x is not compatible with v0.3.x.

## Why MoveResourceState Was Not Implemented

Terraform's Plugin Framework provides a `MoveResourceState` mechanism that allows a provider to accept state from an old resource type and transform it into a new schema. Combined with user-side `moved` blocks, this could theoretically automate the migration. We chose not to implement this for three reasons:

1. **The v0.2.x provider never implemented `Read`.** All four resource `Read` functions were no-ops that returned immediately without calling the API. This means existing state files do not contain accurate infrastructure data — they only hold whatever was returned at create time, with no subsequent refresh. Migrating this stale state would give users a false sense of accuracy.

2. **The provider is Speakeasy-generated.** `MoveResourceState` methods would need to be manually added to generated resource files and maintained across regeneration cycles. The ongoing maintenance cost is not justified for a one-time migration from a provider version with four resources.

3. **The schema changes are too extensive.** Field renames, type changes (string to int32), structural changes (flat fields to nested objects), and removed fields span every resource. The transformation logic would be complex, brittle, and difficult to test — especially given point 1.

The import-based migration below is more reliable because `terraform import` calls the real API and populates state with accurate, current data.

## Migration Steps

### Step 1: Inventory Existing Resources

Before making any changes, record what v0.2.x resources you have and their IDs:

```bash
terraform state list
terraform state show netskope_privateapps.example
terraform state show netskope_publishers.example
terraform state show netskope_publisher_upgrade_profiles.example
```

Save the output. You will need the resource IDs for re-import.

> **Note:** The v0.2.x `netskope_ipsec_tunnels` resource was non-functional — it stored a Unix timestamp as its ID, never captured the real API ID, and had a no-op `Read`. Any existing state for this resource should simply be removed. Use the new `netskope_ip_sec_tunnel` resource in v0.3.x to manage IPSec tunnels from scratch.

### Step 2: Remove Old Resources from State

Remove each resource from Terraform state **without destroying the actual infrastructure**:

```bash
terraform state rm netskope_privateapps.example
terraform state rm netskope_publishers.example
terraform state rm netskope_publisher_upgrade_profiles.example

# If you have any ipsec tunnel state, remove it too — it cannot be migrated
terraform state rm netskope_ipsec_tunnels.example
```

Repeat for every resource instance. This only removes Terraform's tracking — the real infrastructure is untouched.

### Step 3: Update Provider Configuration

**v0.2.x:**
```terraform
provider "netskope" {
  baseurl  = "https://mytenant.goskope.com/api/v2"
  apitoken = var.ns_api_token
}
```

**v0.3.x:**
```terraform
terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = "~> 0.3.3"
    }
  }
}

# Option 1: Environment variables (recommended)
# Set NETSKOPE_SERVER_URL and NETSKOPE_API_KEY, then use an empty block:
provider "netskope" {}

# Option 2: Explicit configuration
provider "netskope" {
  server_url = "https://mytenant.goskope.com/api/v2"
  api_key    = var.netskope_api_key
}
```

### Step 4: Rewrite Resource Configurations

Update each resource block to use the new resource type names and schemas.

#### Private Apps

```terraform
# v0.2.x
resource "netskope_privateapps" "example" {
  app_name              = "my-app"
  host                  = "app.internal.example.com"
  use_publisher_dns     = true
  clientless_access     = false
  trust_self_signed_certs = false
  protocols {
    type = "tcp"
    port = "443"
  }
  publisher {
    publisher_id   = "123"
    publisher_name = "pub01"
  }
  tags {
    tag_name = "production"
  }
}

# v0.3.x
resource "netskope_npa_private_app" "example" {
  private_app_name        = "my-app"
  private_app_hostname    = "app.internal.example.com"
  private_app_protocol    = "https"
  real_host               = "app.internal.example.com"
  use_publisher_dns       = true
  clientless_access       = false
  trust_self_signed_certs = false
  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]
  publishers = [
    {
      publisher_id   = 123
      publisher_name = "pub01"
    }
  ]
  tags = [
    {
      tag_name = "production"
    }
  ]
}
```

Key changes:
- `app_name` renamed to `private_app_name`
- `host` renamed to `private_app_hostname`
- `publisher` renamed to `publishers`
- `publisher_id` type changed from string to integer
- `protocols.type` renamed to `protocols.protocol`
- Added `private_app_protocol` and `real_host` fields

#### Publishers

```terraform
# v0.2.x
resource "netskope_publishers" "example" {
  name                          = "pub01"
  lbrokerconnect                = false
  publisher_upgrade_profiles_id = 1
}
# Output: id (string), token (string)

# v0.3.x
resource "netskope_npa_publisher" "example" {
  publisher_name                = "pub01"
  lbrokerconnect                = false
  publisher_upgrade_profiles_id = 1
}
# Output: publisher_id (int32)
# Note: token is now a separate resource (netskope_npa_publisher_token)
```

Key changes:
- `name` renamed to `publisher_name`
- `id` replaced by `publisher_id` (int32)
- `token` removed — use `netskope_npa_publisher_token` resource instead

#### Publisher Upgrade Profiles

```terraform
# v0.2.x
resource "netskope_publisher_upgrade_profiles" "example" {
  name         = "weekly-upgrade"
  docker_tag   = "latest"
  enabled      = true
  frequency    = "weekly"
  release_type = "Beta"
  timezone     = "America/Los_Angeles"
}

# v0.3.x
resource "netskope_npa_publisher_upgrade_profile" "example" {
  name         = "weekly-upgrade"
  docker_tag   = "latest"
  enabled      = true
  frequency    = "weekly"
  release_type = "Beta"
  timezone     = "America/Los_Angeles"
}
```

Key changes:
- Resource type renamed (pluralized to singular)
- `id` replaced by `publisher_upgrade_profile_id` (int32)
- Schema fields are otherwise largely the same

### Step 5: Import Resources

After rewriting your HCL, import each existing resource into the new state:

```bash
# Private apps — use the private_app_id from the Netskope API
terraform import netskope_npa_private_app.example 12345

# Publishers — use the publisher_id
terraform import netskope_npa_publisher.example 67890

# Publisher upgrade profiles — use the profile ID
terraform import netskope_npa_publisher_upgrade_profile.example 42
```

### Step 6: Verify with Plan

Run a plan to confirm the imported state matches your configuration:

```bash
terraform plan
```

If you see unexpected diffs, adjust your HCL to match the actual infrastructure state. Common causes:
- Boolean fields that default differently in the new provider
- Optional fields that the API returns but you did not specify in HCL
- List ordering differences

### Data Source Migration

Data sources were also renamed. No state migration is needed — just update the references:

| v0.2.x | v0.3.x |
|---|---|
| `netskope_publishers` | `netskope_npa_publisher` or `netskope_npa_publishers_list` |
| `netskope_privateapps` | `netskope_npa_private_app` or `netskope_npa_private_apps_list` |
| `netskope_ipsec_tunnels` | `netskope_ip_sec_tunnels_list` |
| `netskope_ipsec_pops` | `netskope_ip_sec_po_ps_list` |
| `netskope_publisher_upgrade_profiles` | `netskope_npa_publisher_upgrade_profile` or `netskope_npa_publisher_upgrade_profiles_list` |

## Resource Name Quick Reference

| v0.2.x Resource | v0.3.x Resource |
|---|---|
| `netskope_privateapps` | `netskope_npa_private_app` |
| `netskope_publishers` | `netskope_npa_publisher` |
| `netskope_publisher_upgrade_profiles` | `netskope_npa_publisher_upgrade_profile` |
| `netskope_ipsec_tunnels` | None (was non-functional; see `netskope_ip_sec_tunnel`) |

## New Resources in v0.3.x

The following resources are new and have no functional v0.2.x equivalent:

- `netskope_npa_policy_groups` — NPA policy group management
- `netskope_npa_rules` — NPA policy rule management
- `netskope_npa_publisher_token` — Publisher registration tokens (was embedded in publisher resource in v0.2.x)
- `netskope_npa_publishers_alerts_configuration` — Publisher alert settings
- `netskope_npa_publishers_bulk_profile_updates` — Bulk publisher profile updates
- `netskope_npa_publishers_bulk_upgrade_request` — Bulk publisher upgrades
- `netskope_npa_private_app_public_host` — Private app public host configuration
- `netskope_gre_tunnel` — GRE tunnel management
- `netskope_ip_sec_tunnel` — IPSec tunnel management (replaces non-functional `netskope_ipsec_tunnels`)
- `netskope_npa_local_broker` — NPA local broker management
- `netskope_npa_local_broker_config` — Local broker configuration
- `netskope_npa_local_broker_token` — Local broker registration tokens

---

## Upgrading from v0.3.2 to v0.3.3

Version 0.3.3 includes schema corrections from auditing OAS files against the real API. Resource type names are unchanged, but several attribute-level breaking changes affect existing state. Users with `netskope_npa_rules`, `netskope_npa_policy_groups`, or `netskope_npa_private_app` resources in state must take action.

### Affected Resources

#### `netskope_npa_rules`

**State-breaking changes:**

| Change | v0.3.2 | v0.3.3 | Impact |
|---|---|---|---|
| ID field renamed | `rule_id` (string) | `id` (string) | State key mismatch; import uses `id` |
| `silent` removed | Optional string | Removed | State contains unknown attribute |
| `rule_order.rule_id` type changed | string | int64 | State value incompatible |
| `rule_data.tss_actions.tss_profile` type changed | list(string) | string | State value incompatible |
| `rule_data.private_apps_with_activities.app_id` removed | list(string) | Removed | State contains unknown attribute |
| `rule_data.periodic_reauth` added | — | Optional nested object | No impact (new, computed) |

**HCL changes required:**

```terraform
# v0.3.2
resource "netskope_npa_rules" "example" {
  # ...
  rule_order = {
    rule_id = "42"    # was string
    # ...
  }
}

# v0.3.3
resource "netskope_npa_rules" "example" {
  # ...
  rule_order = {
    rule_id = 42      # now int64
    # ...
  }
}
```

Remove any `silent` attribute from your HCL.

#### `netskope_npa_policy_groups`

**State-breaking changes:**

| Change | v0.3.2 | v0.3.3 | Impact |
|---|---|---|---|
| ID field renamed | `group_id` (string) | `id` (string) | State key mismatch; import uses `id` |
| `silent` removed | Optional string | Removed | State contains unknown attribute |
| `status` removed | Computed string | Removed | State contains unknown attribute |
| `group_order` flattened | `group_order.group_order.{group_id, order}` | `group_order.{group_id, order}` | State structure incompatible |

**HCL changes required:**

```terraform
# v0.3.2
resource "netskope_npa_policy_groups" "example" {
  group_name = "my-group"
  group_order = {
    group_order = {
      group_id = "1"
      order    = "after"
    }
  }
}

# v0.3.3
resource "netskope_npa_policy_groups" "example" {
  group_name = "my-group"
  group_order = {
    group_id = "1"
    order    = "after"
  }
}
```

Remove any `silent` attribute from your HCL.

#### `netskope_npa_private_app`

**State-breaking changes:**

| Change | v0.3.2 | v0.3.3 | Impact |
|---|---|---|---|
| `app_name` removed | Optional string | Removed | State contains unknown attribute |
| `app_option` removed | Optional nested object | Removed | State contains unknown attribute |
| `policies` removed | Computed list(string) | Removed | State contains unknown attribute |
| `reachability` removed | Computed nested object | Removed | State contains unknown attribute |
| `service_publisher_assignments` removed | Computed list | Removed | State contains unknown attribute |
| `steering_configs` removed | Computed list(string) | Removed | State contains unknown attribute |
| `tags.tag_id` type changed | string | int32 | State value incompatible |

**HCL changes required:**

Remove `app_name`, `app_option`, and any references to computed-only fields that were removed. If you were setting `app_name`, use `private_app_name` instead (which existed in both versions).

#### `netskope_npa_publisher` (minor)

| Change | v0.3.2 | v0.3.3 | Impact |
|---|---|---|---|
| `connected_apps` removed | Computed list(string) | Removed | State contains unknown attribute |
| `labels`, `label_ids` added | — | New computed fields | No impact |

#### `netskope_npa_publisher_upgrade_profile` (minor)

| Change | v0.3.2 | v0.3.3 | Impact |
|---|---|---|---|
| `timezone_id` added | — | New computed field | No impact |

### Upgrade Steps

**For `netskope_npa_publisher` and `netskope_npa_publisher_upgrade_profile`:** No action required. New computed fields are additive. The removed `connected_apps` field may cause a warning on first refresh but the provider will recover.

**For `netskope_npa_rules`, `netskope_npa_policy_groups`, and `netskope_npa_private_app`:** The ID renames, type changes, and removed attributes make existing state incompatible. Re-import these resources:

```bash
# 1. Record existing resource IDs
terraform state show netskope_npa_rules.example
terraform state show netskope_npa_policy_groups.example
terraform state show netskope_npa_private_app.example

# 2. Remove from state
terraform state rm netskope_npa_rules.example
terraform state rm netskope_npa_policy_groups.example
terraform state rm netskope_npa_private_app.example

# 3. Update HCL (remove silent, fix group_order nesting, fix rule_order.rule_id type)

# 4. Re-import using the resource ID from step 1
terraform import netskope_npa_rules.example <rule_id>
terraform import netskope_npa_policy_groups.example <group_id>
terraform import netskope_npa_private_app.example <private_app_id>

# 5. Verify
terraform plan
```