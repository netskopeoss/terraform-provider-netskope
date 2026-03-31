variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name

  tunnel_config = {
    enable     = true
    allow_list = ["dns2tcp", "iodine", "vpnoverdns"]
  }
}
