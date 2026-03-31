variable "name" {
  type = string
}

resource "netskope_npa_local_broker" "test" {
  local_broker_name    = var.name
  city_name            = "Cupertino"
  region_name          = "CA"
  country_name         = "United States of America"
  country_code         = "US"
  latitude             = 37.323
  longitude            = -122.032
  custom_public_ip     = "203.0.113.42"
  custom_private_ip    = "192.168.19.119"
  access_via_public_ip = "NONE"
}
