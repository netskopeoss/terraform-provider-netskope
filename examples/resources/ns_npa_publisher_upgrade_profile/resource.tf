resource "ns_npa_publisher_upgrade_profile" "my_npapublisherupgradeprofile" {
  docker_tag   = "...my_docker_tag..."
  enabled      = 3
  frequency    = "...my_frequency..."
  name         = "Dr. Sheri Howe"
  release_type = "...my_release_type..."
  timezone     = "...my_timezone..."
}