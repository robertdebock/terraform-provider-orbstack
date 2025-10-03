terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_network_config" "main" {
  ipv4_subnet     = "192.168.200.0/24"
  bridge_enabled  = true
  expose_ssh_port = true
}

output "network_status" { value = orbstack_network_config.main.status }
output "network_subnet" { value = orbstack_network_config.main.ipv4_subnet }

