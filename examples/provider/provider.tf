terraform {
  required_providers {
    netskope = {
      source  = "netskopeoss/netskope"
      version = "0.4.6"
    }
  }
}

provider "netskope" {
  server_url = "..." # Optional - can use NETSKOPE_SERVER_URL environment variable
  tenant     = "..." # Optional
}