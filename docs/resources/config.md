# Resource: orbstack_config

The `orbstack_config` resource manages OrbStack configuration settings using the `orb config` command.

## Example Usage

```hcl
# Set memory allocation
resource "orbstack_config" "memory" {
  key   = "memory_mib"
  value = "8192"
}

# Set CPU allocation
resource "orbstack_config" "cpus" {
  key   = "cpus"
  value = "4"
}

# Set disk size
resource "orbstack_config" "disk" {
  key   = "disk_gb"
  value = "100"
}
```

## Argument Reference

The following arguments are supported:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `key` | `string` | Yes | - | The configuration key to set |
| `value` | `string` | Yes | - | The value to set for the configuration key |

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `id` | `string` | The unique identifier (same as key) |
| `key` | `string` | The configuration key |
| `value` | `string` | The configuration value |

## Common Configuration Keys

| Key | Description | Example Values |
|-----|-------------|----------------|
| `memory_mib` | Memory allocation in MiB | `"4096"`, `"8192"` |
| `cpus` | Number of CPUs | `"2"`, `"4"`, `"8"` |
| `disk_gb` | Disk size in GB | `"50"`, `"100"`, `"200"` |
| `network_mode` | Network configuration | `"bridge"`, `"host"` |

## Notes

- Configuration changes take effect immediately
- Some settings may require OrbStack to be restarted
- Use `orb config` command to see all available configuration options
- The resource will recreate if the key or value changes
