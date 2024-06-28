terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.6.15"
    }
  }
}

provider "ns" {
  # Configuration options
}