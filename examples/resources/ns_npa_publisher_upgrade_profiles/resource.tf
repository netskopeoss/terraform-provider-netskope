resource "ns_npa_publisher_upgrade_profiles" "my_npapublisherupgradeprofiles" {
  docker_tag   = "...my_docker_tag..."
  enabled      = 0
  frequency    = "...my_frequency..."
  name         = "Craig Vandervort"
  release_type = "...my_release_type..."
  timezone     = "...my_timezone..."
}