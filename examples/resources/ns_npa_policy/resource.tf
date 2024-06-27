resource "ns_npa_policy" "my_npapolicy" {
  description = "any"
  enabled     = "1"
  group_name  = "My policy group"
  id          = "6078b1a2-6433-4712-8526-e2e9c5e43473"
  rule_name   = "vantest"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = "1"
    rule_name = "api-policy-managed"
  }
}