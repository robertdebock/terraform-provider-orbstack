output "machine_name" {
  description = "Name of the created machine"
  value       = orbstack_machine.demo.name
}

output "machine_username" {
  description = "Username for the default user"
  value       = orbstack_machine.demo.username
}

output "machine_arch" {
  description = "Architecture of the machine"
  value       = orbstack_machine.demo.arch
}

output "machine_ip" {
  description = "IP address of the machine"
  value       = orbstack_machine.demo.ip_address
}

output "ssh_connection" {
  description = "SSH connection string"
  value       = "ssh ${orbstack_machine.demo.username}@${orbstack_machine.demo.ip_address}"
}
