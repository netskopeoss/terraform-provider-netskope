resource "terraform_npa_policy_group" "my_npapolicygroup" {
  group_name = "...my_group_name..."
  group_order = {
    group_id = "1"
    order    = "before"
  }
  id = "ca75d546-2042-4ad5-926c-75d642b6450a"
}