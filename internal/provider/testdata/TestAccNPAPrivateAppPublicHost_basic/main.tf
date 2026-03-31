variable "name" {
  type = string
}

resource "netskope_npa_private_app_public_host" "test" {
  host              = "${var.name}.example.internal"
  real_host         = "192.168.1.100"
  clientless_access = true

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]
}
