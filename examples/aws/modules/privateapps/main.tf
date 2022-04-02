resource "netskope_privateapps" "nsapp01" {
  app_name = "Private EC2 Instances"
  host     = var.private_dns
  protocols {
    type = "tcp"
    port = "22"
  }

  publisher {
    publisher_id   = var.pub_id
    publisher_name = var.pub_name
  }

}