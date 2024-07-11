resource "ns_npa_publishers_alerts_configuration_update" "my_npapublishersalertsconfigurationupdate" {
  admin_users = [
    "...",
  ]
  event_types = [
    "UPGRADE_FAILED",
  ]
  selected_users = "...my_selected_users..."
}