terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.7.99"
    }
  }
}

provider "ns" {
  # Configuration options
}