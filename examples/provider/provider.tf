terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.3.180"
    }
  }
}

provider "ns" {
  # Configuration options
}