terraform {
  required_providers {
    netskope = {
      version = "0.1.0"
      source  = "github.com/netskopeoss/netskope"
    }
  }
}

provider "aws" {
  region  = var.aws_region
}

//Netskope Resources
//
//Create Publisher in Netskope
resource "netskope_publishers" "Publisher" {
  name = var.publisher
}

//Create Netskope Private App for Internal AWS Domains
resource "netskope_privateapps" "PrivateApp" {
  app_name = var.privateapp
  host     = var.privateapp_host

  //A single protocol is required but both TCP an UDP can be sent as shown
  protocols {
    type = "tcp"
    port = var.privateapp_tcp
  }

  protocols {
    type = "udp"
    port = var.privateapp_udp
  }

  publisher {
    publisher_id   = netskope_publishers.Publisher.id
    publisher_name = netskope_publishers.Publisher.name
  }

}

//AWS Resources
//
// Filter Netskope Publishers AMIs for the latests
data "aws_ami" "npa-publisher" {
  most_recent = true
  owners      = ["679593333241"]

  filter {
    name   = "name"
    values = ["Netskope Private Access Publisher*"]
  }
}

// Create EC2 Instance for the Publisher
resource "aws_instance" "NpaPublisher" {
  ami                         = data.aws_ami.npa-publisher.id
  associate_public_ip_address = true
  instance_type               = var.aws_instance_type
  key_name                    = var.aws_key_name
  subnet_id                   = var.aws_subnet
  vpc_security_group_ids      = [var.aws_security_group]
  user_data                   = netskope_publishers.Publisher.token

  tags = {
    "Name" = var.publisher
  }

}

//Publisher Details
output "publisher_details" {
  description = "Name/IP and Token for the Publisher"
  value = {
    name  = "${netskope_publishers.Publisher.name}"
    ip = "${aws_instance.NpaPublisher.public_ip}"
    token = "${netskope_publishers.Publisher.token}"
    
  }
}
