resource "ns_npa_policy" "my_npapolicy" {
  description = "any"
  enabled     = "1"
  group_name  = "My policy group"
  rule_data = {
    access_method          = "Clientless"
    b_negate_net_location  = true
    b_negate_src_countries = true
    classification         = "...my_classification..."
    dlp_actions = [
      {
        actions = [
          "bypass",
        ]
        dlp_profile = "Payment Card"
      },
    ]
    external_dlp = false
    json_version = 3
    match_criteria_action = {
      action_name = "block"
    }
    net_location_obj = [
      "...",
    ]
    organization_units = [
      "...",
    ]
    policy_type = "private-app"
    private_app_ids = [
      "...",
    ]
    private_apps = [
      "...",
    ]
    private_apps_with_activities = [
      {
        activities = [
          {
            activity = "any"
            list_of_constraints = [
              "...",
            ]
          },
        ]
        app_name = "[172.31.12.135]"
      },
    ]
    private_app_tag_ids = [
      "...",
    ]
    private_app_tags = [
      "...",
    ]
    show_dlp_profile_action_table = false
    src_countries = [
      "...",
    ]
    user_groups = [
      "...",
    ]
    users = [
      "...",
    ]
    user_type = "user"
    version   = 1
  }
  rule_name = "van-test"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = 1
    rule_name = "api-policy-managed"
  }
}