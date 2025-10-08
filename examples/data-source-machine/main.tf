terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 3.1.0"
    }
  }
}

provider "orbstack" {}

# Create a machine and set it as default
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

# Use data source to read the default machine
data "orbstack_machine" "default_info" {
  name = "default-vm"
}

# Use data source to read the secondary machine
data "orbstack_machine" "secondary_info" {
  name = "secondary-vm"
}

# Output information about both machines
output "default_machine_info" {
  value = {
    name            = data.orbstack_machine.default_info.name
    ip_address      = data.orbstack_machine.default_info.ip_address
    status          = data.orbstack_machine.default_info.status
    is_default      = data.orbstack_machine.default_info.default_machine
  }
}

output "secondary_machine_info" {
  value = {
    name            = data.orbstack_machine.secondary_info.name
    ip_address      = data.orbstack_machine.secondary_info.ip_address
    status          = data.orbstack_machine.secondary_info.status
    is_default      = data.orbstack_machine.secondary_info.default_machine
  }
}
