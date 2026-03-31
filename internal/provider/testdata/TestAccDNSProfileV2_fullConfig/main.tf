variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name        = var.name
  description = "Full config acceptance test"
  log_traffic = "All DNS"

  domain_config = {
    security_categories = [
      {
        name   = "Newly Registered Domain"
        action = "Block"
      },
      {
        name   = "Security Risk - Botnets"
        action = "Sinkhole"
      }
    ]
    allow_list = [
      {
        record_types = ["A", "AAAA"]
        domain_names = ["allowed.example.com"]
      }
    ]
    block_list = [
      {
        record_types = ["All Record Types"]
        domain_names = ["blocked.example.com"]
      }
    ]
    sinkhole_ip              = "1.2.3.4"
    block_all_except_allow_list = false
  }

  tunnel_config = {
    enable     = true
    allow_list = ["dns2tcp", "iodine"]
  }

  custom_config = {
    enable                   = true
    server_ip                = ["8.8.8.8"]
    bypass_original_dns      = true
    fallback_to_netskope_dns = true
  }
}
