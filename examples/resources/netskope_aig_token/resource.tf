resource "netskope_aig_token" "my_aigtoken" {
  expire_in = {
    unit  = "hour"
    value = 24
  }
  name           = "IT key-2025-06-01-1"
  token_group_id = "019812aa-bdba-7a48-baa2-2de8c6006fda"
}