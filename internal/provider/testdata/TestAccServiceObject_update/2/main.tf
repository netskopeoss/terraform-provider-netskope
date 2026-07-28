variable "name" {
  type = string
}

resource "netskope_service_object" "test" {
  name        = var.name
  description = "Updated by acceptance test"
  protocols = {
    tcp = ["80", "443"]
    udp = ["53"]
  }
}
