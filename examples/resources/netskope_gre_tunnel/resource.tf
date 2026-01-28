resource "netskope_gre_tunnel" "my_gretunnel" {
  bandwidth = 0
  enabled   = false
  notes     = "...my_notes..."
  options = {
    xff = {
      xff_enabled = false
      xff_ip_list = [
        "..."
      ]
    }
  }
  pop_names = [
    "..."
  ]
  site        = "...my_site..."
  source_ip   = "...my_source_ip..."
  source_type = "Machine"
  template    = "...my_template..."
  vendor      = "...my_vendor..."
}