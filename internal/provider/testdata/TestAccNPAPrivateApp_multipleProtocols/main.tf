variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.name}-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = var.name
  private_app_hostname = "192.168.1.100"

  # IMPORTANT: Protocols must be ordered to match API response ordering to avoid drift.
  # The API sorts by: protocol type (tcp before udp), then port number ascending.
  # See docs/KNOWN_API_ISSUES.md - Issue #14
  protocols = [
    {
      port     = "22"
      protocol = "tcp"
    },
    {
      port     = "443"
      protocol = "tcp"
    },
    {
      port     = "53"
      protocol = "udp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
