# Data Source: orbstack_machine

The `orbstack_machine` data source reads information about an existing OrbStack machine.

## Example Usage

```hcl
data "orbstack_machine" "existing" {
  name = "my-vm"
}

output "machine_ip" {
  value = data.orbstack_machine.existing.ip_address
}

output "machine_status" {
  value = data.orbstack_machine.existing.status
}

output "is_default_machine" {
  value = data.orbstack_machine.existing.default_machine
}
```

## Argument Reference

The following arguments are supported:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `name` | `string` | Yes | - | The name of the machine to read |

## Attributes Reference

The following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `id` | `string` | The unique identifier of the machine |
| `name` | `string` | The name of the machine |
| `image` | `string` | The base image used (may include OS:VERSION format) |
| `status` | `string` | The current status of the machine |
| `ip_address` | `string` | The IP address of the machine |
| `ssh_host` | `string` | SSH host (usually same as ip_address) |
| `ssh_port` | `number` | SSH port |
| `created_at` | `string` | Creation time as reported by orb info |
| `default_machine` | `bool` | Whether this machine is the current default machine |

## Notes

- This data source will fail if the machine does not exist
- Use this to reference existing machines in your Terraform configuration
- The machine must be managed by OrbStack to be discoverable
