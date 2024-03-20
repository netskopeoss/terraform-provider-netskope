resource "terraform_npa_policy_group" "my_npapolicygroup" {
  group_name = "...my_group_name..."
  group_order = {
    group_id = "1"
    order    = "after"
  }
}