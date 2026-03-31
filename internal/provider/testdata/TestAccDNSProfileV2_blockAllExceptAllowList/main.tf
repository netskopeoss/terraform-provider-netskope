variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name

  domain_config = {
    block_all_except_allow_list = true
    allow_list = [
      {
        record_types = ["A"]
        domain_names = ["permitted.example.com"]
      }
    ]
  }
}
