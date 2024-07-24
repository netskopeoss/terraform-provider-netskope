resource "ns_npa_publisher_upgrade_profiles" "my_npapublisherupgradeprofiles" {
  docker_tag   = 8690
  enabled      = true
  frequency    = "0 0 1 * TUE"
  name         = "My Upgrade Profile"
  release_type = "Latest"
  timezone     = "US/Eastern"
}