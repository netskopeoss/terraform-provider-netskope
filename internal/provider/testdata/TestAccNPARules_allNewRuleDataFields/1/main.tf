# Step 1: create with notify, rule_data.description, users, user_groups,
# src_countries, private_app_tag_ids — mirrors the tf-test-all-npa policy (rule 21)
variable "name" {
  type = string
}

variable "test_user" {
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
  protocols = [{ port = "443", protocol = "tcp" }]
  publishers = [{
    publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
    publisher_name = netskope_npa_publisher.test.publisher_name
  }]
  use_publisher_dns       = true
  trust_self_signed_certs = false
}

resource "netskope_npa_rules" "test" {
  rule_name   = var.name
  description = "Acceptance test — all new rule_data fields"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type           = "private-app"
    match_criteria_action = { action_name = "allow" }
    private_apps          = [netskope_npa_private_app.test.private_app_name]
    access_method         = ["Client"]

    description = "rule-data-description-v1"

    notify = {
      emails   = ["test@example.com"]
      interval = "60"
      to_users = ["admin"]
    }

    users        = [var.test_user]
    user_groups  = ["admin_groups"]
    src_countries = ["AL"]

    # private_app_tag_ids: tag ID 1542 is a pre-provisioned test tag on the acceptance test tenant
    private_app_tag_ids = ["1542"]
  }
}
