resource "orbstack_machine" "ubuntu_ci" {
  name       = "ubuntu-cloudinit-vm"
  image      = "ubuntu"
  cloud_init = <<-EOT
  #cloud-config
  hostname: ubuntu-cloudinit-vm
  users:
    - name: hank
      sudo: ALL=(ALL) NOPASSWD:ALL
      groups: users, admin
      shell: /bin/bash
      lock_passwd: true
      ssh_authorized_keys:
        - ${chomp(file("~/.ssh/id_rsa.pub"))}
  package_update: true
  packages:
    - curl
    - htop
  runcmd:
    - [ sh, -c, "echo 'Hello from cloud-init' > /etc/motd" ]
  EOT
}

