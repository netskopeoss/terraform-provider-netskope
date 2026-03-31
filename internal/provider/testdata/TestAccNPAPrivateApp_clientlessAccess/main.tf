variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.name}-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name           = var.name
  clientless_access          = true
  real_host                  = "browser.internal.test"
  private_app_protocol       = "http"

  protocols = [
    {
      port     = "80"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  allow_unauthenticated_cors = true
  use_publisher_dns          = false
  trust_self_signed_certs    = true
}
