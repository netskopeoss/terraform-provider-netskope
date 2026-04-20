# NPA Rule Ordering Examples

The `rule_order` block controls where a rule is placed in the policy evaluation order. Rules are evaluated top-to-bottom; the first matching rule wins.

## Supported Order Values

| Value    | Description                                    | Requires        |
|----------|------------------------------------------------|-----------------|
| `top`    | Place rule at the top of the list              | -               |
| `bottom` | Place rule at the bottom of the list           | -               |
| `after`  | Place rule immediately after a specific rule   | `rule_id`       |
| `before` | Place rule immediately before a specific rule  | `rule_id`       |

> **Note:** `rule_order` is write-only. The API accepts it on create/update but does not return it in GET responses, so it cannot be imported.

## Shared Infrastructure

All examples below assume this shared infrastructure is defined:

```hcl
resource "netskope_npa_policy_groups" "example" {
  group_name = "example-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "example" {
  publisher_name = "example-publisher"
}

resource "netskope_npa_private_app" "example" {
  private_app_name     = "example-app"
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.example.publisher_id)
      publisher_name = netskope_npa_publisher.example.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
```

## Example 1: Place a Rule After Another

Creates two rules where `rule2` is placed immediately after `rule1`.

**Result:** rule1, rule2

```hcl
resource "netskope_npa_rules" "rule1" {
  rule_name = "allow-app-admins"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}

resource "netskope_npa_rules" "rule2" {
  rule_name = "allow-app-users"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order   = "after"
    rule_id = tonumber(netskope_npa_rules.rule1.id)
  }
}
```

## Example 2: Place a Rule Before Another

Creates two rules where `rule2` is inserted before `rule1`, making `rule2` evaluate first.

**Result:** rule2, rule1

```hcl
resource "netskope_npa_rules" "rule1" {
  rule_name = "allow-general-access"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}

resource "netskope_npa_rules" "rule2" {
  rule_name = "block-restricted-users"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order   = "before"
    rule_id = tonumber(netskope_npa_rules.rule1.id)
  }
}
```

## Example 3: Top and Bottom Placement

Creates two rules where `rule1` goes to the top and `rule2` goes to the bottom of the evaluation list.

**Result:** rule1, ...(other existing rules)..., rule2

```hcl
resource "netskope_npa_rules" "rule1" {
  rule_name = "priority-allow"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order = "top"
  }
}

resource "netskope_npa_rules" "rule2" {
  depends_on = [netskope_npa_rules.rule1]

  rule_name = "catchall-allow"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.example.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.example.private_app_name]
    access_method = ["Client"]
  }

  rule_order = {
    order = "bottom"
  }
}
```

> **Tip:** When using `top`/`bottom` without an explicit dependency between rules, add `depends_on` to ensure consistent ordering. Without it, Terraform may create the rules in parallel and the final order depends on which API call completes first.

## Verifying Rule Order with a Data Source

The `netskope_npa_rules_list` data source returns rules in their current evaluation order. You can use it to inspect or output the live rule ordering:

```hcl
data "netskope_npa_rules_list" "all" {
  depends_on = [
    netskope_npa_rules.rule1,
    netskope_npa_rules.rule2,
  ]
}

output "rule_evaluation_order" {
  description = "Rules in evaluation order (first match wins)"
  value = [
    for rule in data.netskope_npa_rules_list.all.data : {
      id   = rule.id
      name = rule.rule_name
    }
  ]
}
```

After `terraform apply`, the output shows the live order:

```
rule_evaluation_order = [
  { id = "215", name = "allow-app-admins" },
  { id = "216", name = "allow-app-users" },
  ...
]
```

You can also look up a single rule by ID to confirm its attributes:

```hcl
data "netskope_npa_rules" "rule1" {
  id = netskope_npa_rules.rule1.id
}

output "rule1_details" {
  value = {
    name    = data.netskope_npa_rules.rule1.rule_name
    enabled = data.netskope_npa_rules.rule1.enabled
    action  = data.netskope_npa_rules.rule1.rule_data.match_criteria_action.action_name
  }
}
```

## Test Results

These examples are backed by acceptance tests that verify the actual rule evaluation order via the NPA rules list API. All tests passed on 2026-04-20:

```
=== RUN   TestAccNPARules_ruleOrderAfter
--- PASS: TestAccNPARules_ruleOrderAfter (7.29s)
=== RUN   TestAccNPARules_ruleOrderBottom
--- PASS: TestAccNPARules_ruleOrderBottom (13.89s)
=== RUN   TestAccNPARules_ruleOrderBefore
--- PASS: TestAccNPARules_ruleOrderBefore (13.71s)
PASS
```

Each test creates the rules, then queries the list endpoint and asserts the rules appear in the expected evaluation order.
