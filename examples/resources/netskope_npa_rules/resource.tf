resource "netskope_npa_rules" "my_nparules" {
  description = "any"
  enabled     = "1"
  group_id    = "1"
  rule_data = {
    access_method = [
      "Clientless"
    ]
    b_negate_net_location  = false
    b_negate_src_countries = false
    device_classification_id = [
      9
    ]
    json_version = 3
    match_criteria_action = {
      action_name = "allow"
    }
    net_location_obj = [
      "..."
    ]
    organization_units = [
      "..."
    ]
    policy_type = "private-app"
    private_app_tag_ids = [
      "..."
    ]
    private_app_tags = [
      "..."
    ]
    private_apps = [
      "..."
    ]
    src_countries = [
      "..."
    ]
    user_groups = [
      "..."
    ]
    users = [
      "..."
    ]
  }
  rule_name = "vantest"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = 1
    rule_name = "api-policy-managed"
  }
}