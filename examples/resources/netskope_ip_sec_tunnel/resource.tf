resource "netskope_ip_sec_tunnel" "my_ipsectunnel" {
  bandwidth  = 2
  enabled    = true
  encryption = "...my_encryption..."
  notes      = "...my_notes..."
  options = {
    reauth = true
    rekey  = false
    xff = {
      enabled = true
      iplist = [
        "..."
      ]
    }
  }
  pop_names = [
    "..."
  ]
  psk             = "...my_psk..."
  site            = "...my_site..."
  source_identity = "...my_source_identity..."
  source_ip       = "...my_source_ip..."
  source_type     = "...my_source_type..."
  template        = "...my_template..."
  vendor          = "...my_vendor..."
}