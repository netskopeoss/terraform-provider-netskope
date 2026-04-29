# Netskope Terraform Provider — v0.4.2 / v0.4.3 Update

This document summarizes fixes and new features in the upcoming provider releases that address the issues you reported.

---

## 1. Hostname Whitespace Drift — Fixed in v0.4.3

### What was happening

When a private app is configured with multiple comma-separated hostnames, the Netskope API normalizes whitespace around the commas on read-back. For example:

- **Config sends:** `"10.33.5.143,webapp02.example.com"` (no space)
- **API returns:** `"10.33.5.143, webapp02.example.com"` (space added)

This caused Terraform to detect a difference between the configured value and the state on every `terraform plan`, producing a perpetual update-in-place diff:

```
  # netskope_npa_private_app.example will be updated in-place
  ~ resource "netskope_npa_private_app" "example" {
      ~ private_app_hostname = "10.33.5.143, webapp02.example.com" -> "10.33.5.143,webapp02.example.com"
    }
```

### What we fixed

The provider now normalizes whitespace in `private_app_hostname` before comparing the plan to the state. It splits on commas, trims each entry, and compares the normalized values. If they match, the diff is suppressed. No changes to your Terraform configuration are needed — the fix is entirely provider-side.

### Action required

Upgrade to provider v0.4.3. No config changes needed.

---

## 2. Port Drift from CSV Port Strings — Module-Side Fix Required

### What is happening

If your YAML definitions or Terraform variables define ports as comma-separated strings (e.g., `"80, 443"`), the provider sends this as a single protocol entry. The API accepts it on create, but splits it into separate entries on read-back:

```
# What Terraform sends:
protocols = [{ port = "80, 443", protocol = "tcp" }]    # 1 entry

# What the API returns:
protocols = [
  { port = "80",  protocol = "tcp" },                    # 2 entries
  { port = "443", protocol = "tcp" },
]
```

On the next `terraform plan`, Terraform sees one entry in config vs. two in state and plans an update. This update sends the CSV string again, the API splits it again, and the cycle repeats every run.

This is not a provider bug — the provider correctly stores what the API returns. The issue is that the API transforms the input, so the config and state can never converge when CSV strings are used.

### How to fix it

Split CSV port strings into individual protocol entries before passing them to the resource. If your apps are defined in YAML like this:

```yaml
# policies.yaml
- name: my-app
  hostname: "10.33.5.143"
  ports: "80, 443, 8080"
  protocol: tcp
```

Update your `main.tf` to split the ports:

```hcl
locals {
  apps = yamldecode(file("policies.yaml"))
}

resource "netskope_npa_private_app" "app" {
  for_each = { for app in local.apps : app.name => app }

  private_app_name     = each.value.name
  private_app_hostname = each.value.hostname

  # Split CSV port strings into individual protocol entries
  protocols = flatten([
    for port in split(",", each.value.ports) : {
      port     = trimspace(port)
      protocol = each.value.protocol
    }
  ])

  # ... publishers, use_publisher_dns, etc.
}
```

If you define ports inline in HCL, list them as separate entries:

```hcl
# WRONG — causes perpetual drift:
protocols = [
  { port = "80, 443", protocol = "tcp" }
]

# CORRECT — one entry per port:
protocols = [
  { port = "80",  protocol = "tcp" },
  { port = "443", protocol = "tcp" },
]
```

### Action required

Update your module's `main.tf` to split CSV port strings. After the change, run `terraform apply` once to normalize the state. Subsequent plans will be clean.

---

## 3. Device Classification Tags — Available in v0.4.2

### What's new

The provider now supports device classification tags via:

- **`netskope_device_classification_tag`** — Resource for creating and managing tags (CRUD with import)
- **`netskope_device_classification_tag_list`** — Data source to list all tags on the tenant
- **`netskope_device_classification_options_list`** — Data source to list available classification options

### Using "Managed" device classification in rules

The Netskope UI shows "Managed" and "Unmanaged" as device categories, but these are convenience groupings — they are not distinct entities in the API. "Managed" means the device matches **any** device classification tag. To replicate this in Terraform, include all tag IDs:

```hcl
data "netskope_device_classification_tag_list" "all" {}

resource "netskope_npa_rules" "managed_devices_only" {
  rule_name = "require-managed-device"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.terraform.id

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

To select specific tags instead:

```hcl
locals {
  tag_ids = {
    for t in data.netskope_device_classification_tag_list.all.tags :
    t.name => t.tag_id
  }
}

# Only devices with CrowdStrike OR Intune
device_classification_id = [
  tostring(local.tag_ids["CrowdStrike Installed"]),
  tostring(local.tag_ids["Intune Managed"]),
]
```

A full working example is available in the examples repository:
https://github.com/netskopeoss/terraform-netskope-examples/tree/main/examples/npa/device-classification

---

## 4. NPA Rule Ordering — New `netskope_npa_rules_order` Resource

### Ordering individual rules with `rule_order`

Each rule has a `rule_order` attribute that controls where it appears in the list:

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

Supported values: `top`, `bottom`, `after` (requires `rule_id`), `before` (requires `rule_id`).

### Bulk ordering with `netskope_npa_rules_order`

For larger deployments where you manage many rules with `for_each`, the new `netskope_npa_rules_order` resource lets you control the list position of all rules in a single place. Define your policies in a YAML file — the list order becomes the display order in the Netskope UI.

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

- name: allow-dba
  action: allow
  apps: ["Database-Servers"]
  groups: ["Database-Admins"]

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

# Step 1: Create all rules at bottom (parallel, fast)
resource "netskope_npa_rules" "bulk" {
  for_each = local.policy_map

  rule_name = each.value.name
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.terraform.id

  rule_data = {
    policy_type           = "private-app"
    match_criteria_action = { action_name = each.value.action }
    private_apps          = each.value.apps
    user_groups           = each.value.groups
    access_method         = ["Client"]
  }

  rule_order = { order = "bottom" }

  lifecycle {
    ignore_changes = [rule_order]
  }
}

# Step 2: Set list positions (list order = display order in UI)
resource "netskope_npa_rules_order" "main" {
  rule_ids = [for p in local.policies : netskope_npa_rules.bulk[p.name].id]
}
```

With this pattern:
- **Adding a rule:** add a line to the YAML at the desired position. Only the new rule is created; existing rules are untouched. The `netskope_npa_rules_order` resource updates to include the new rule.
- **Removing a rule:** remove the line from the YAML. The rule is destroyed and the order resource updates.
- **Reordering:** move lines in the YAML. Only the order resource updates — no rules are re-created.
- **Disabling a rule:** add `enabled: "0"` to the YAML entry. The rule stays in the list but is skipped during evaluation.

### Best practice: use a dedicated policy group

We recommend placing Terraform-managed rules in their own policy group to keep them isolated from manually-created rules:

```hcl
resource "netskope_npa_policy_groups" "terraform" {
  group_name = "Terraform-Managed"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}
```

---

## Summary of Required Actions

| Issue | Fix Location | Action |
|-------|-------------|--------|
| Hostname whitespace drift | Provider (v0.4.3) | Upgrade provider — no config changes |
| CSV port string drift | Your module | Split CSV ports into individual protocol entries |
| Device classification | Provider (v0.4.2) | Use `device_classification_tag_list` data source |
| Rule ordering | Provider (v0.4.3) | Optional: use `netskope_npa_rules_order` for bulk management |

---

## Resources

- [NPA Rules Guide](https://github.com/netskopeoss/terraform-provider-netskope/blob/main/docs/guides/NPA_RULES_GUIDE.md)
- [Device Classification Example](https://github.com/netskopeoss/terraform-netskope-examples/tree/main/examples/npa/device-classification)
- [Policy as Code Example](https://github.com/netskopeoss/terraform-netskope-examples/tree/main/examples/npa/policy-as-code)
- [Rule Ordering Examples](https://github.com/netskopeoss/terraform-provider-netskope/blob/main/docs/examples/RULE_ORDERING.md)
- [Changelog](https://github.com/netskopeoss/terraform-provider-netskope/blob/main/CHANGELOG.md)
