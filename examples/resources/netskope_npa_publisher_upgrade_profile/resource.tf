resource "netskope_npa_publisher_upgrade_profile" "my_npapublisherupgradeprofile" {
  docker_tag   = 8690
  enabled      = true
  frequency    = "0 0 1 * TUE"
  name         = "My Upgrade Profile"
  release_type = "Latest"
  timezone     = "US/Eastern"
}