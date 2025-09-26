terraform {
  required_providers {
    orbstack = {
      source = "robertdebock/orbstack"
    }
  }
}

provider "orbstack" {
  # Configuration options
}

resource "orbstack_machine" "demo" {
  name     = "demo-with-username"
  image    = "ubuntu:noble"
  username = "demo"
  arch     = "arm64"
  
  cloud_init = <<-EOF
    #cloud-config
    users:
      - name: demo
        sudo: ALL=(ALL) NOPASSWD:ALL
        ssh_authorized_keys:
          - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7vbqajDhA...
  EOF
}
