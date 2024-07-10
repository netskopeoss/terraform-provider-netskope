resource "ns_npa_publishers" "my_npapublishers" {
  lbrokerconnect                = false
  name                          = "npa_publisher_1"
  publisher_upgrade_profiles_id = 1
}