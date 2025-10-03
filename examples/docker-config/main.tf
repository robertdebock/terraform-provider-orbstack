terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {
  # Provider configuration
}

# Docker Engine Configuration
resource "orbstack_docker_config" "main" {
  set_context           = true
  expose_ports_to_lan  = true
  node_name            = "orbstack-dev"
}

# Output the status
output "docker_status" {
  value = orbstack_docker_config.main.status
}

output "docker_endpoint" {
  value = orbstack_docker_config.main.docker_endpoint
}

output "context_active" {
  value = orbstack_docker_config.main.context_active
}
