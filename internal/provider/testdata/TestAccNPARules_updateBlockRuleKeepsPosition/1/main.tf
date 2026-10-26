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

# Step 1: block rule_a created at bottom with lifecycle ignore_changes on rule_order.
# After create, private state is empty; the next plan triggers the one-time converging
# diff (block_page.html -> "Default Template"). ExpectNonEmptyPlan: true in the test.
resource "netskope_npa_rules" "rule_a" {
  rule_name   = "${var.name}-a"
  description = "block rule keeps position — step 1"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_order = {
    order = "bottom"
  }

  lifecycle {
    ignore_changes = [rule_order]
  }

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "block"
      emit_alert  = true
      template    = "Default Template"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}

# rule_b is placed immediately after rule_a so the live order is [..., rule_a, rule_b].
# Position is asserted via checkNPARulePosition after every step.
resource "netskope_npa_rules" "rule_b" {
  rule_name   = "${var.name}-b"
  description = "anchor allow rule — stays below rule_a"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_order = {
    order   = "after"
    rule_id = tostring(netskope_npa_rules.rule_a.id)
  }

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.test.private_app_name]
    access_method = ["Client"]
  }
}
