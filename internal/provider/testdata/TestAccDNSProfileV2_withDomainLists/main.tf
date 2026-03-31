variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name

  domain_config = {
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
  }
}
