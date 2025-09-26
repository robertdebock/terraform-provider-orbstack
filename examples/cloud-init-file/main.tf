terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 1.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine" "vm" {
  name            = "cloudinit-file-vm"
  image           = "ubuntu"
  cloud_init_file = "cloud-init.yaml"
}

output "vm_ip" {
  value = orbstack_machine.vm.ip_address
}


