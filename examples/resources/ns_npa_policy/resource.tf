resource "ns_npa_policy" "my_npapolicy" {
  enabled    = "1"
  group_name = "My policy group"
  id         = "50c698ae-0859-4eba-ab5d-becc58716c1a"
  rule_name  = "vantest"
  rule_order = {
    order     = "before"
    position  = 5
    rule_id   = "1"
    rule_name = "api-policy-managed"
  }
}