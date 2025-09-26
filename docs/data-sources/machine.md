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
| `image` | `string` | The base image used |
| `tag` | `string` | The image tag used |
| `status` | `string` | The current status of the machine |
| `ip_address` | `string` | The IP address of the machine |
| `cpus` | `number` | Number of CPUs allocated |
| `memory_mib` | `number` | Memory allocated in MiB |
| `disk_gb` | `number` | Disk size in GB |
| `power_state` | `string` | Current power state (running, stopped, etc.) |

## Notes

- This data source will fail if the machine does not exist
- Use this to reference existing machines in your Terraform configuration
- The machine must be managed by OrbStack to be discoverable
