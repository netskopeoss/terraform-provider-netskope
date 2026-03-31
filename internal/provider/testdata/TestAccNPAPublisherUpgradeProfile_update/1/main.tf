variable "name" {
  type = string
}

data "netskope_npa_publishers_releases_list" "releases" {}

resource "netskope_npa_publisher_upgrade_profile" "test" {
  name         = var.name
  docker_tag   = data.netskope_npa_publishers_releases_list.releases.data[0].docker_tag
  enabled      = true
  frequency    = "0 0 * * *"
  release_type = "Beta"
  timezone     = "US/Pacific"
}
