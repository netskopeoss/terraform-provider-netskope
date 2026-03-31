
variable "label_name" {
  type = string
}

variable "app_name" {
  type = string
}

resource "netskope_rbac_label" "test" {
  name  = var.label_name
  color = "#0294C9"
}

resource "netskope_npa_publisher" "test" {
  publisher_name = "${var.app_name}-publisher"
}

resource "netskope_npa_private_app" "test" {
  private_app_name     = var.app_name
  private_app_hostname = "192.168.1.100"
  label_ids            = [netskope_rbac_label.test.label_id]

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

  use_publisher_dns       = true
  trust_self_signed_certs = false
}
