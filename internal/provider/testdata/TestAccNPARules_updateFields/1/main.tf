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

resource "netskope_npa_private_app" "app1" {
  private_app_name     = "${var.name}-app1"
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

resource "netskope_npa_private_app" "app2" {
  private_app_name     = "${var.name}-app2"
  private_app_hostname = "192.168.1.101"

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

resource "netskope_npa_rules" "test" {
  rule_name   = "${var.name}-original"
  description = "Original description"
  enabled     = "1"
  group_id    = netskope_npa_policy_groups.test.id

  rule_data = {
    policy_type = "private-app"

    match_criteria_action = {
      action_name = "allow"
    }

    private_apps  = [netskope_npa_private_app.app1.private_app_name]
    access_method = ["Client"]
  }
}
