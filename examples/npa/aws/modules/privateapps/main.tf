resource "netskope_privateapps" "nsapp01" {
  app_name = "${var.project}-internal-ec2"
  host     = var.private_dns
  protocols {
    type = "tcp"
    port = "22"
  }

  publisher {
    publisher_id   = var.pub01_id
    publisher_name = var.pub01_name
  }

  publisher {
    publisher_id   = var.pub02_id
    publisher_name = var.pub02_name
  }


}