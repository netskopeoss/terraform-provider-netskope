variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name

  custom_config = {
    enable                   = true
    server_ip                = ["8.8.8.8"]
    bypass_original_dns      = false
    fallback_to_netskope_dns = true
  }
}
