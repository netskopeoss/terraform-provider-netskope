variable "name" {
  type = string
}

resource "netskope_npa_policy_groups" "test" {
  group_name = "${var.name}-group"

  group_order = {
    group_id = "2"
    order    = "after"
  }
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.name}-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = "${var.name}-app"
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}

resource "netskope_npa_rules" "rule_a" {
  rule_name = "${var.name}-rule-a"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type           = "private-app"
    match_criteria_action = { action_name = "allow" }
    private_apps          = [netskope_npa_private_app.test.private_app_name]
    access_method         = ["Client"]
  }

  rule_order = { order = "bottom" }

  lifecycle {
    ignore_changes = [rule_order]
  }
}

resource "netskope_npa_rules" "rule_b" {
  rule_name = "${var.name}-rule-b"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type           = "private-app"
    match_criteria_action = { action_name = "allow" }
    private_apps          = [netskope_npa_private_app.test.private_app_name]
    access_method         = ["Client"]
  }

  rule_order = { order = "bottom" }

  lifecycle {
    ignore_changes = [rule_order]
  }
}

resource "netskope_npa_rules" "rule_c" {
  rule_name = "${var.name}-rule-c"
  enabled   = "1"
  group_id  = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type           = "private-app"
    match_criteria_action = { action_name = "allow" }
    private_apps          = [netskope_npa_private_app.test.private_app_name]
    access_method         = ["Client"]
  }

  rule_order = { order = "bottom" }

  lifecycle {
    ignore_changes = [rule_order]
  }
}

# Order: A, B, C
resource "netskope_npa_rules_order" "test" {
  rule_ids = [
    netskope_npa_rules.rule_a.id,
    netskope_npa_rules.rule_b.id,
    netskope_npa_rules.rule_c.id,
  ]
}
