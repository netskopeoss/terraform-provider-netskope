variable "name" {
  type = string
}

resource "netskope_service_object" "test" {
  name        = var.name
  description = "Acceptance test service object"
  protocols = {
    tcp = ["443"]
  }
}

data "netskope_service_object_list" "test" {
  depends_on = [netskope_service_object.test]
}
