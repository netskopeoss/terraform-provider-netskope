variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.name}-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = var.name
  private_app_hostname = "192.168.1.100,192.168.1.101"

  # IMPORTANT: Protocols must be in ascending port order to avoid drift
  # See docs/KNOWN_API_ISSUES.md - Issue #14
  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns          = true
  trust_self_signed_certs    = true
  clientless_access          = false
  is_user_portal_app         = false
  allow_unauthenticated_cors = false
}
