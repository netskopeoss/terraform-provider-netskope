variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name
}

data "netskope_dns_profile_v2" "test" {
  profile_id = netskope_dns_profile_v2.test.profile_id
}
