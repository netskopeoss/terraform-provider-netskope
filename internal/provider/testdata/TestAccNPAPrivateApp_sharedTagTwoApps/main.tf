# Regression scenario for issue #120: two private apps created in the SAME apply,
# with NO depends_on between them, both carrying the same NEW tag_name.
# The API creates the tag row on first use; two concurrent creates raced on the
# (tag_name, tag_type) unique constraint and the second app failed with
# "Duplicate entry ... for key 'tag_unique'".
variable "name" {
  type = string
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.name}-publisher"
}

resource "netskope_npa_private_app" "a" {
  private_app_name     = "${var.name}-a"
  private_app_hostname = "192.0.2.1"

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

  # Unique per run so the tag row does not pre-exist on the tenant.
  tags = [
    {
      tag_name = "${var.name}-shared"
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}

resource "netskope_npa_private_app" "b" {
  private_app_name     = "${var.name}-b"
  private_app_hostname = "192.0.2.2"

  protocols = [
    {
      port     = "8443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = tostring(netskope_npa_publisher.test.publisher_id)
      publisher_name = netskope_npa_publisher.test.publisher_name
    }
  ]

  tags = [
    {
      tag_name = "${var.name}-shared"
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
