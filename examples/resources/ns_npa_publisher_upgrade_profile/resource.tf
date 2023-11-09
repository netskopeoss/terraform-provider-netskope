resource "ns_npa_publisher_upgrade_profile" "my_npapublisherupgradeprofile" {
  docker_tag   = "...my_docker_tag..."
  enabled      = 6
  frequency    = "...my_frequency..."
  name         = "Jerald Graham MD"
  release_type = "...my_release_type..."
  timezone     = "...my_timezone..."
}