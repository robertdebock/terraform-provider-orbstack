terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = "~> 1.0"
    }
  }
}

provider "orbstack" {
  # orb_path = "/usr/local/bin/orb"  # Optional: path to orb CLI
}
