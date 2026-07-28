variable "name" {
  type = string
}

resource "netskope_rbac_role" "test" {
  name        = var.name
  description = "RBAC role with IP allow list"
  api_groups = [
    {
      api_group_id = 107
      permission   = "r"
    }
  ]
  ip_allow_list = {
    enable_ip_allow_list = true
    ip_list              = ["10.0.0.1", "192.168.1.0/24"]
  }
}
