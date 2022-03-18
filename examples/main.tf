terraform {
  required_providers {
    netskope = {
      version = "0.1.6"
      source = "github.com/ns-sbrown/netskope"
    }
  }
}

//Provider block optional. Alternatively "NS_BaseURL" & "NS_ApiToken" Env Variables Can be used.
provider "netskope" {
    baseurl = "https://tenant-url.goskope.com"
    apitoken = "<api token>"
}

resource "netskope_publishers" "publisher_name" {
  name = "PublisherName"
}

output "publisher_details" {
  value = {
    name         = "${netskope_publisher.publisher_name.name}"
    token         = "${netskope_publisher.publisher_name.token}"
  }  
}


resource "netskope_publisher" "publisher_name" {
  name = "PublisherName"
}

output "publisher_details" {
  value = {
    name         = "${netskope_publisher.publisher_name.name}"
    token         = "${netskope_publisher.publisher_name.token}"
  }  
}