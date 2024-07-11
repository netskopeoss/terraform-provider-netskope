resource "ns_npa_publishers_alerts_configuration_update" "my_npapublishersalertsconfigurationupdate" {
  admin_users = [
    "admin1@abc.com",
  ]
  event_types = [
    "UPGRADE_FAILED",
  ]
  selected_users = "admin1@abc.com"
}