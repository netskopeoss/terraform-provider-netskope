variable "name" {
  type = string
}

resource "netskope_aig_appliance" "test" {
  name = var.name
  host = "${var.name}.example.com"

  ports = {
    http = {
      enable = true
      port   = 80
    }
    https = {
      enable = true
      port   = 443
    }
  }
}
