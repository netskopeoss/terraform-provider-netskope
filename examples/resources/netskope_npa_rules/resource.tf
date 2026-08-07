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
    classification = [
      "..."
    ]
    description = "...my_description..."
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
      "27",
      "42",
    ]
    notify = {
      emails = [
        "..."
      ]
      from_user = "...my_from_user..."
      interval  = "30"
      to_users = [
        "..."
      ]
    }
    organization_units = [
      "engineering/qa",
    ]
    os = [
      "Windows",
      "Mac",
    ]
    periodic_reauth = {
      reauth_interval      = "60"
      reauth_interval_unit = "hours"
    }
    policy_type = "private-app"
    private_app_tag_ids = [
      "1",
      "2",
    ]
    private_app_tags = [
      "tag1",
      "tag2",
    ]
    private_apps = [
      "app1",
      "app2",
    ]
    schedule = [
      {
        time_interval_obj = [
          "..."
        ]
        time_range = [
          {
            end_date   = "MM/DD/YYYY"
            end_time   = "HH:MM"
            start_date = "MM/DD/YYYY"
            start_time = "HH:MM"
          }
        ]
      }
    ]
    src_countries = [
      "US",
      "AF",
      "CN",
    ]
    user_confidence = {
      index    = "351"
      operator = "lt"
    }
    user_groups = [
      "usergroup/group1",
    ]
    user_type = "user"
    users = [
      "user@example.com",
    ]
    version = 1
  }
  rule_name = "vantest"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = 1
    rule_name = "api-policy-managed"
  }
}