terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

data "orbstack_images" "debian" {
  filter = "debian"
}

locals {
  chosen = length(data.orbstack_images.debian.images) > 0 ? data.orbstack_images.debian.images[0].name : "debian"
}

resource "orbstack_machine" "vm" {
  name           = "validate-image-vm"
  arch           = "amd64"
  image          = local.chosen
  validate_image = true
}

output "vm_ip" {
  value = orbstack_machine.vm.ip_address
}



