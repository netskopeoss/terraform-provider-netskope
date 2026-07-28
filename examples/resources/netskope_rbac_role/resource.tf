resource "netskope_rbac_role" "my_rbacrole" {
  api_groups = [
    {
      api_group_id = 2
      permission   = "rwa"
    }
  ]
  description = "...my_description..."
  ip_allow_list = {
    enable_ip_allow_list = true
    ip_list = [
      "..."
    ]
  }
  name = "...my_name..."
}