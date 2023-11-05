resource "ns_npa_publisher_alerts" "my_npapublisheralerts" {
  admin_users = [
    ["admin1@abc.com", "admin2@abc.com"],
  ]
  event_types = [
    "UPGRADE_SUCCEEDED",
  ]
  selected_users = "abc@xyz.com,def@xyz.com"
}