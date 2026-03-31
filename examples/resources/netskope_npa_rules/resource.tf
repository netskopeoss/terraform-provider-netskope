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
      "..."
    ]
    json_version = 3
    match_criteria_action = {
      action_name = "allow"
      emit_alert  = true
      template    = "...my_template..."
    }
    net_location_obj = [
      "190.123.150.10",
      "190.218.0.0/16",
    ]
    organization_units = [
      "engineering/qa",
    ]
    policy_type = "private-app"
    private_app_tags = [
      "tag1",
      "tag2",
    ]
    private_apps = [
      "app1",
      "app2",
    ]
    src_countries = [
      "US",
      "AF",
      "CN",
    ]
    user_groups = [
      "usergroup/group1",
    ]
    users = [
      "vphan@netskope.com",
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