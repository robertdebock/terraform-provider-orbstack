terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 3.1.0"
    }
  }
}

provider "orbstack" {}

# Create a machine that will be set as the default
resource "orbstack_machine" "default_vm" {
  name            = "default-vm"
  image           = "ubuntu:noble"
  default_machine = true
}

# Create another machine (not default)
resource "orbstack_machine" "secondary_vm" {
  name  = "secondary-vm"
  image = "ubuntu:noble"
}

# Output the default machine name
output "default_machine_name" {
  value = orbstack_machine.default_vm.name
}

# Output whether the default machine is set
output "is_default_machine" {
  value = orbstack_machine.default_vm.default_machine
}
