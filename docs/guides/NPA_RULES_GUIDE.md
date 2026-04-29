# NPA Rules Guide

How to create and manage NPA policy rules with the Netskope Terraform provider.

## Concepts

NPA rules control access to private applications. The `rule_order` attribute controls where a rule is placed in the list.

---

## Basic Rule Creation

### Allow Rule

```hcl
resource "netskope_npa_rules" "allow_admins" {
  rule_name   = "allow-admin-access"
  description = "Allow IT admins to access infrastructure apps"
  enabled     = "1"
  group_id    = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.infra.private_app_name]
    user_groups   = ["IT-Administrators"]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}
```

### Block Rule

```hcl
resource "netskope_npa_rules" "block_terminated" {
  rule_name   = "block-terminated-users"
  description = "Deny all access for terminated employees"
  enabled     = "1"
  group_id    = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "block"
    }

    private_apps  = [netskope_npa_private_app.all_apps.private_app_name]
    user_groups   = ["Terminated-Users"]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}
```

### Rule with Device Classification

Requires provider >= 0.4.2. Device classification tags must be created in the Netskope UI first.

```hcl
data "netskope_device_classification_tag_list" "all" {}

locals {
  tag_ids = {
    for t in data.netskope_device_classification_tag_list.all.tags :
    t.name => t.tag_id
  }
}

resource "netskope_npa_rules" "edr_required" {
  rule_name = "require-edr"
  enabled   = "1"
  group_id  = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps             = [netskope_npa_private_app.secure_app.private_app_name]
    access_method            = ["Client"]
    device_classification_id = [tostring(local.tag_ids["CrowdStrike Installed"])]
  }

  rule_order = {
    order = "top"
  }
}
```

### Rule Matching by App Tags

Instead of listing app names, match all apps with a specific tag:

```hcl
resource "netskope_npa_rules" "web_access" {
  rule_name = "allow-web-apps"
  enabled   = "1"
  group_id  = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_app_tags = ["web-tier"]
    user_groups      = ["Engineering"]
    access_method    = ["Client", "Clientless"]
  }

  rule_order = {
    order = "bottom"
  }
}
```

---

## Rule Placement

The `rule_order` attribute controls where a rule appears in the list. This is for organizational purposes — it does not change evaluation behavior.

### Available Order Values

| Value | Description | Requires |
|-------|-------------|----------|
| `top` | Place at the top of the list | - |
| `bottom` | Place at the bottom of the list | - |
| `after` | Place after a specific rule | `rule_id` |
| `before` | Place before a specific rule | `rule_id` |

### Example: Placing Rules Relative to Each Other

```hcl
resource "netskope_npa_rules" "deny" {
  rule_name = "deny-blocked-users"
  # ...
  rule_order = { order = "top" }
}

resource "netskope_npa_rules" "allow" {
  rule_name = "allow-admin-access"
  # ...
  rule_order = {
    order   = "after"
    rule_id = tonumber(netskope_npa_rules.deny.id)
  }
  depends_on = [netskope_npa_rules.deny]
}
```

> **Note:** `rule_order` is write-only. The API accepts it on create/update but does not return it in GET responses, so it cannot be imported.

---

## Bulk Rule Management with `for_each`

For large deployments, use `for_each` to create rules from a list or map.

### Example: YAML-Driven

Define policies in a YAML file:

**policies.yaml:**

```yaml
- name: deny-terminated
  action: block
  apps: ["All-Internal-Apps"]
  groups: ["Terminated-Users"]

- name: allow-admins
  action: allow
  apps: ["Infrastructure"]
  groups: ["IT-Administrators"]

- name: allow-developers
  action: allow
  apps: ["Web-Apps"]
  groups: ["Engineering"]

- name: deny-all
  action: block
  apps: ["All-Internal-Apps"]
  groups: []
```

**main.tf:**

```hcl
locals {
  policies   = yamldecode(file("policies.yaml"))
  policy_map = { for p in local.policies : p.name => p }
}

resource "netskope_npa_rules" "bulk" {
  for_each = local.policy_map

  rule_name = each.value.name
  enabled   = "1"
  group_id  = var.group_id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = each.value.action
    }

    private_apps  = each.value.apps
    user_groups   = each.value.groups
    access_method = ["Client"]
  }

  rule_order = { order = "bottom" }
}
```

### Disabling a Rule

Add `enabled: "0"` — the rule stays but is skipped during evaluation:

```yaml
- name: allow-contractors
  enabled: "0"          # Disabled, not deleted
  action: allow
  apps: ["Portal"]
  groups: ["Contractors"]
```

---

## Device Classification in Rules

The Netskope UI shows "Managed" and "Unmanaged" as device categories, but these are convenience groupings — not distinct API entities. To replicate "Managed" in Terraform, list all individual device classification tag IDs that make up the managed set.

### Listing All Device Classification Tags

```hcl
data "netskope_device_classification_tag_list" "all" {}

output "all_tags" {
  value = [
    for t in data.netskope_device_classification_tag_list.all.tags : {
      id   = t.tag_id
      name = t.name
    }
  ]
}
```

Run `terraform apply` to see all available tag IDs on your tenant.

### Applying All Tags to a Rule (Equivalent of "Managed")

```hcl
data "netskope_device_classification_tag_list" "all" {}

resource "netskope_npa_rules" "managed_devices_only" {
  rule_name = "require-managed-device"
  enabled   = "1"
  group_id  = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.app.private_app_name]
    access_method = ["Client"]

    # All device classification tags = "Managed" equivalent
    device_classification_id = [
      for t in data.netskope_device_classification_tag_list.all.tags :
      tostring(t.tag_id)
    ]
  }

  rule_order = { order = "top" }
}
```

### Selecting Specific Tags

```hcl
locals {
  tag_ids = {
    for t in data.netskope_device_classification_tag_list.all.tags :
    t.name => t.tag_id
  }
}

resource "netskope_npa_rules" "edr_and_mdm" {
  rule_name = "require-edr-and-mdm"
  enabled   = "1"
  group_id  = data.netskope_npa_policy_groups_list.all.data[0].id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.app.private_app_name]
    access_method = ["Client"]

    device_classification_id = [
      tostring(local.tag_ids["CrowdStrike Installed"]),
      tostring(local.tag_ids["Intune Managed"]),
    ]
  }

  rule_order = { order = "top" }
}
```

---

## Handling CSV Port Strings

If your YAML or variable definitions use comma-separated port strings (e.g. `"80, 443"`), you must split them into individual protocol entries before passing to the provider. The API splits CSV ports on read-back, causing perpetual drift if the config sends them as a single string.

### Problem

```hcl
# BAD: CSV port string causes perpetual drift
protocols = [
  {
    port     = "80, 443"
    protocol = "tcp"
  }
]
```

The API accepts `"80, 443"` on create but returns two separate entries (`"80"` and `"443"`). Every subsequent plan tries to collapse them back — this never converges.

### Fix

```hcl
# GOOD: Split CSV ports into individual entries
locals {
  raw_protocols = [
    { port = "80, 443", protocol = "tcp" },
    { port = "8080",    protocol = "tcp" },
    { port = "53",      protocol = "udp" },
  ]

  # Flatten CSV port strings into individual port/protocol pairs
  protocols = flatten([
    for p in local.raw_protocols : [
      for port in split(",", p.port) : {
        port     = trimspace(port)
        protocol = p.protocol
      }
    ]
  ])
}

resource "netskope_npa_private_app" "example" {
  # ...
  protocols = local.protocols
}
```

This produces one entry per port, matching what the API returns on read-back.

---

## Common Mistakes

| Mistake | Error | Fix |
|---------|-------|-----|
| `enabled = true` | Type error | Use string: `enabled = "1"` |
| `enabled = 1` | Type error | Use string: `enabled = "1"` |
| App name in brackets | "Private app [[name]] doesn't exist" | Use plain strings: `app.name` |
| `rule_order = { order = "after", rule_id = "5" }` | Type error | Use `tonumber()`: `rule_id = tonumber(other_rule.id)` |
| CSV port strings: `port = "80, 443"` | Perpetual drift | Split: `for port in split(",", p.port)` |

---

## Rule Data Fields Reference

| Field | Type | Description |
|-------|------|-------------|
| `policy_type` | string | Always `"private-app"` for NPA |
| `match_criteria_action.action_name` | string | `"allow"` or `"block"` |
| `private_apps` | list(string) | App names to match |
| `private_app_tags` | list(string) | Match apps by tag (alternative to `private_apps`) |
| `user_groups` | list(string) | User groups to match (from IdP) |
| `access_method` | list(string) | `"Client"`, `"Clientless"`, or both |
| `device_classification_id` | list(string) | Device classification tag IDs (v0.4.2+) |
| `net_location_obj` | list(string) | Network location objects (requires feature flag) |
| `src_countries` | list(string) | Source country codes |
| `b_negate_src_countries` | bool | Negate country match |
| `organization_units` | list(string) | Organization units to match |
| `json_version` | number | Schema version (use `3`) |
