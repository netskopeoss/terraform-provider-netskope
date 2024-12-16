resource "ns_npa_rules" "my_nparules" {
  description = "any"
  enabled     = "1"
  group_id    = "1"
  group_name  = "My policy group"
  rule_data = {
    access_method = [
      "Client"
    ]
    b_negate_net_location  = false
    b_negate_src_countries = false
    classification         = "...my_classification..."
    device_classification_id = [
      9
    ]
    dlp_actions = [
      {
        actions = [
          "bypass"
        ]
        dlp_profile = "Payment Card"
      }
    ]
    external_dlp = true
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
    private_apps_with_activities = [
      {
        activities = [
          {
            activity = "any"
            list_of_constraints = [
              "..."
            ]
          }
        ]
        app_name = "[172.31.12.135]"
      }
    ]
    show_dlp_profile_action_table = true
    src_countries = [
      "..."
    ]
    tss_actions = [
      {
        actions = [
          {
            action_name         = "allow"
            remediation_profile = "...my_remediation_profile..."
            severity            = "low"
            template            = "...my_template..."
          }
        ]
        tss_profile = [
          "..."
        ]
      }
    ]
    tss_profile = [
      "..."
    ]
    user_groups = [
      "..."
    ]
    user_type = "user"
    users = [
      "..."
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
  silent = "1"
}