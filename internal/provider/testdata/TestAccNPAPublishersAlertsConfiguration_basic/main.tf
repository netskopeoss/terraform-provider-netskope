resource "netskope_npa_publishers_alerts_configuration" "test" {
  admin_users    = ["jharris@netskope.com"]
  event_types    = ["publisher_up", "publisher_down"]
  selected_users = "selected"
}
