resource "terraform_npa_policy" "my_npapolicy" {
  description = "any"
  enabled     = "1"
  group_name  = "My policy group"
  rule_name   = "vantest"
  rule_order = {
    order     = "bottom"
    position  = 5
    rule_id   = "1"
    rule_name = "api-policy-managed"
  }
}