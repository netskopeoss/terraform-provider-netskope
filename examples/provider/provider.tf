terraform {
  required_providers {
    ns = {
      source  = "netskope/ns"
      version = "0.3.140"
    }
  }
}

provider "ns" {
  # Configuration options
}