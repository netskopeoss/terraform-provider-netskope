variable "name" {
  type = string
}

resource "netskope_dns_profile_v2" "test" {
  name = var.name
}

data "netskope_dns_profile_v2_list" "test" {
  depends_on = [netskope_dns_profile_v2.test]
}
