terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine" "vm" {
  name           = "validate-image-vm"
  arch           = "amd64"
  image          = "debian:bookworm"
  validate_image = false
}

output "vm_ip" {
  value = orbstack_machine.vm.ip_address
}



