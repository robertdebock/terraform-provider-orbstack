# Resource: orbstack_machine

The `orbstack_machine` resource creates and manages Linux machines in OrbStack.

## Example Usage

```hcl
resource "orbstack_machine" "vm" {
  name            = "demo-vm"
  image           = "ubuntu:noble"  # Use OS:VERSION format for specific versions
  username        = "demo"           # Username for the default user
  arch            = "arm64"          # Architecture: amd64 or arm64
  default_machine = true             # Set this machine as the default
  
  cloud_init = <<-EOF
    #cloud-config
    users:
      - name: demo
        sudo: ALL=(ALL) NOPASSWD:ALL
        ssh_authorized_keys:
          - ssh-rsa AAAAB3NzaC1yc2E...
  EOF
}
```

## Argument Reference

The following arguments are supported:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `name` | `string` | Yes | - | The name of the machine |
| `image` | `string` | No | `"ubuntu"` | The base image/distribution. Use OS:VERSION format for specific versions (e.g., ubuntu:noble, debian:bookworm) |
| `username` | `string` | No | macOS username | Username for the default user |
| `arch` | `string` | No | - | Architecture: `amd64` or `arm64` |
| `cloud_init` | `string` | No | - | Cloud-init user data passed during creation |
| `cloud_init_file` | `string` | No | - | Path to a cloud-init user data file. Overrides `cloud_init` if both set |
| `validate_image` | `bool` | No | `false` | Validate image exists before create; fail fast if unknown |
| `power_state` | `string` | No | - | Desired power state: `running` or `stopped` |
| `default_machine` | `bool` | No | `false` | Set this machine as the default machine for OrbStack. Only one machine can be the default. |

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `id` | `string` | The unique identifier of the machine |
| `name` | `string` | The name of the machine |
| `image` | `string` | The base image used (may include OS:VERSION format) |
| `username` | `string` | The username for the default user |
| `arch` | `string` | The architecture of the machine |
| `status` | `string` | The current status of the machine |
| `ip_address` | `string` | The IP address of the machine |
| `ssh_host` | `string` | SSH host (usually same as ip_address) |
| `ssh_port` | `number` | SSH port |
| `created_at` | `string` | Creation time as reported by orb info |
| `power_state` | `string` | Current power state (running, stopped, etc.) |
| `default_machine` | `bool` | Whether this machine is the current default machine |

## Notes

- The machine is recreated if any immutable attributes change (image, cloud_init, cloud_init_file, username, arch)
- Cloud-init data is passed during machine creation and may not be applied if the image doesn't support it
- Use `validate_image = true` to ensure the image exists before attempting to create the machine
- The `cloud_init_file` argument takes precedence over `cloud_init` if both are specified
- **Architecture**: Use `arch = "arm64"` for Apple Silicon or `arch = "amd64"` for Intel-based systems. If an invalid architecture is specified, OrbStack will return an error during creation.
- **Username**: If not specified, defaults to your macOS username
- **Default Machine**: Only one machine can be set as the default at a time. Setting `default_machine = true` on one machine will automatically unset the default status from any other machine. The default machine is the one you connect to when running `orb` without specifying a machine name.
