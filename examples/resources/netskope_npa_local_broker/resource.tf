resource "netskope_npa_local_broker" "my_npalocalbroker" {
  access_via_public_ip = "ON_OFF_PREM"
  city_name            = "Cupertino"
  country_code         = "US"
  country_name         = "United States of America"
  custom_private_ip    = "192.168.19.119"
  custom_public_ip     = "203.0.113.42"
  label_ids = [
    "696067da-3890-4fe3-a784-8e682e9b7122"
  ]
  latitude          = 37.323
  local_broker_name = "...my_local_broker_name..."
  longitude         = -122.032
  region_name       = "CA"
}