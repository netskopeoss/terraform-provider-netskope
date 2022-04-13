terraform {
  required_providers {
    netskope = {
      version = "0.1.0"
      source  = "github.com/netskopeoss/netskope"
    }
  }
}

//Provider block optional. Alternatively "NS_BaseURL" & "NS_ApiToken" Env Variables Can be used.
//
//provider "netskope" {
//    baseurl = "https://tenant-url.goskope.com"
//    apitoken = "<api token>"
//}

//Create Primary Publisher
resource "netskope_publishers" "PrimaryPublisherName" {
  name = "PrimaryPublisherName"
}

//Create Secondary Publisher
resource "netskope_publishers" "SecondaryPublisherName" {
  name = "SecondaryPublisherName"
}

//Output Publisher Details
output "primary_publisher_details" {
  value = {
    name  = "${netskope_publishers.PrimaryPublisherName.name}"
    token = "${netskope_publishers.PrimaryPublisherName.token}"
  }
}

output "secondary_publisher_details" {
  value = {
    name  = "${netskope_publishers.SecondaryPublisherName.name}"
    token = "${netskope_publishers.SecondaryPublisherName.token}"
  }
}


//Create Netskope Private App
resource "netskope_privateapps" "PrivateAppName" {
  app_name = "PrivateAppName"
  host     = "www.myexample.com, ssh.example.com"
  protocols {
    type = "tcp"
    port = "22, 80-443"
  }

  protocols {
    type = "udp"
    port = "123, 8443"
  }

  publisher {
    publisher_id   = netskope_publishers.PrimaryPublisherName.id
    publisher_name = netskope_publishers.PrimaryPublisherName.name
  }

  publisher {
    publisher_id   = netskope_publishers.SecondaryPublisherName.id
    publisher_name = netskope_publishers.SecondaryPublisherName.name
  }

  //Optional Param
  use_publisher_dns = true

}