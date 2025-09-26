output "machine_info" {
  value = {
    id         = data.orbstack_machine.example.id
    ip_address = data.orbstack_machine.example.ip_address
    status     = data.orbstack_machine.example.status
    ssh_host   = data.orbstack_machine.example.ssh_host
    ssh_port   = data.orbstack_machine.example.ssh_port
  }
}


