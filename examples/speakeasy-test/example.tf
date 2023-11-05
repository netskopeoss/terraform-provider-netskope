terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.0.1"
    }
  }
}

variable "api_key" {}
variable "server_url" {}

provider "ns" {
  api_key    = var.api_key
  server_url = var.server_url
}

# Note: error: You cannot consume this service. Testing to be done with higher access creds
#data "ns_npa_publisher_alerts" "alerts" {}

resource "ns_publisher_token" "my_publishertoken" {
  publisher_id = ns_npa_publisher.my_npapublisher.id
}

resource "ns_npa_publisher" "my_npapublisher" {
  name = "Speakeasy Terraform Provider Test"
}

# Note: got a 503 one time? Does this need a retry?
data "ns_npa_publishers" "npa_publishers" {
  depends_on = [ns_npa_publisher.my_npapublisher]
}


output "npa_publisher_token" {
  value = ns_publisher_token.my_publishertoken.token
}
