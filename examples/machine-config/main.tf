terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine_config" "main" {
  expose_ports_to_lan = true
  forward_ports       = true
}

output "machine_defaults_status" { value = orbstack_machine_config.main.status }

