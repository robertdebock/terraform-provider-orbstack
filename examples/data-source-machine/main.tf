terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine" "example" {
  name  = "example-ds-vm"
  image = "ubuntu"
}

data "orbstack_machine" "example" {
  name = orbstack_machine.example.name
}


