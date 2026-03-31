variable "name" {
  type = string
}

variable "publisher_suffix" {
  type = string
}

resource "netskope_npa_publisher" "test1" {
  publisher_name = "${var.name}-publisher-1"
}

resource "netskope_npa_publisher" "test2" {
  publisher_name = "${var.name}-publisher-2"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = var.name
  private_app_hostname = "192.168.1.100"

  protocols = [
    {
      port     = "443"
      protocol = "tcp"
    }
  ]

  publishers = [
    {
      publisher_id   = var.publisher_suffix == "1" ? tostring(netskope_npa_publisher.test1.publisher_id) : tostring(netskope_npa_publisher.test2.publisher_id)
      publisher_name = var.publisher_suffix == "1" ? netskope_npa_publisher.test1.publisher_name : netskope_npa_publisher.test2.publisher_name
    }
  ]

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
