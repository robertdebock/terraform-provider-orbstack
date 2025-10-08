terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 3.1.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine" "example" {
  name            = "example-vm"
  image           = "ubuntu"
  default_machine = true
}

 

resource "orbstack_machine" "validate" {
  name           = "validate-image-vm"
  arch           = "amd64"
  image          = "ubuntu:noble"
  validate_image = false
}

resource "orbstack_machine" "ubuntu_tagged" {
  name  = "ubuntu-noble-vm"
  image = "ubuntu:noble"
}

resource "orbstack_machine" "username_arch" {
  name     = "demo-with-username"
  image    = "ubuntu:noble"
  username = "demo"
  arch     = "arm64"
}
