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
  name  = "example-vm"
  image = "ubuntu"
}

data "orbstack_images" "debian" { filter = "debian" }

locals {
  chosen = length(data.orbstack_images.debian.images) > 0 ? data.orbstack_images.debian.images[0].name : "debian"
}

resource "orbstack_machine" "validate" {
  name           = "validate-image-vm"
  arch           = "amd64"
  image          = local.chosen
  validate_image = true
}

resource "orbstack_machine" "debian_tagged" {
  name  = "debian-12-vm"
  image = "debian:bookworm"
}

resource "orbstack_machine" "username_arch" {
  name     = "demo-with-username"
  image    = "ubuntu:noble"
  username = "demo"
  arch     = "arm64"
}
