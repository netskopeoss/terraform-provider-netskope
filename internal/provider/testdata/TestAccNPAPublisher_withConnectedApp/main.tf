variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = var.name
}

# Assign a private app to the publisher so that the publisher's GET-by-ID
# response includes a non-empty connected_apps array.
# This exercises the path fixed in BUG-017: the API returns connected_apps
# as an array of objects, not strings, and the SDK must not fail to deserialize it.
resource "netskope_npa_private_app" "test" {
  private_app_name     = var.name
  private_app_hostname = "test-bug017.internal"

  protocols = [
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

  use_publisher_dns       = false
  trust_self_signed_certs = false
}
