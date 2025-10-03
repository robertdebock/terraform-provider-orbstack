terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

resource "orbstack_machine" "cloudinit_inline" {
  name       = "ci-inline"
  image      = "ubuntu"
  cloud_init = <<-EOT
  #cloud-config
  hostname: ci-inline
  users:
    - name: hank
      sudo: ALL=(ALL) NOPASSWD:ALL
      groups: users, admin
      shell: /bin/bash
      lock_passwd: true
      ssh_authorized_keys:
        - ${chomp(file("~/.ssh/id_rsa.pub"))}
  EOT
}

resource "orbstack_machine" "cloudinit_file" {
  name       = "ci-file"
  image      = "ubuntu"
  cloud_init = file("${path.module}/../cloud-init-file/cloud-init.yaml")
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
