variable "name" {
  type = string
}

resource "netskope_npa_private_app_public_host" "test" {
  host              = "${var.name}-updated.example.internal"
  real_host         = "192.168.1.200"
  clientless_access = true

  protocols = [
    {
      port     = "8443"
      protocol = "tcp"
    }
  ]
}
