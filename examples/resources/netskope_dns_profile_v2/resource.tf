resource "netskope_dns_profile_v2" "my_dnsprofilev2" {
  custom_config = {
    bypass_original_dns      = true
    enable                   = false
    fallback_to_netskope_dns = true
    server_ip = [
      "..."
    ]
  }
  description = "...my_description..."
  domain_config = {
    allow_list = [
      {
        domain_names = [
          "..."
        ]
        record_types = [
          "..."
        ]
      }
    ]
    block_all_except_allow_list = true
    block_list = [
      {
        domain_names = [
          "..."
        ]
        record_types = [
          "..."
        ]
      }
    ]
    security_categories = [
      {
        action = "Sinkhole"
        name   = "...my_name..."
      }
    ]
    sinkhole_ip = "...my_sinkhole_ip..."
  }
  inheritance_groups = [
    "..."
  ]
  log_traffic = "Blocked DNS"
  name        = "...my_name..."
  tunnel_config = {
    allow_list = [
      "..."
    ]
    enable = true
  }
}