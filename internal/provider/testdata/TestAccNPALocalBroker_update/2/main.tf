variable "name" {
  type = string
}

resource "netskope_npa_local_broker" "test" {
  local_broker_name = var.name
  city_name         = "San Francisco"
  region_name       = "CA"
  country_name      = "United States of America"
  country_code      = "US"
}
