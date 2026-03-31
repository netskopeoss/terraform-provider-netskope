variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name

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
    sinkhole_ip = "1.2.3.4"
  }
}
