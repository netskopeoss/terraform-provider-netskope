variable "name" {
  type = string
}

resource "netskope_service_object" "test" {
  name        = var.name
  description = "Initial service object"
  protocols = {
    tcp = ["80"]
  }
}
