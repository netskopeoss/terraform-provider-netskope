resource "ns_npa_policy" "my_npapolicy" {
  description = "any"
  enabled     = "1"
  group_name  = "My policy group"
  id          = "2d55fb8e-8671-44cd-bf16-bf3588e0f08e"
  rule_name   = "vantest"
  rule_order = {
    order     = "top"
    position  = 5
    rule_id   = "1"
    rule_name = "api-policy-managed"
  }
}