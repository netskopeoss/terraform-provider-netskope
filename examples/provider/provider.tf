terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.5.16"
    }
  }
}

provider "ns" {
  # Configuration options
}