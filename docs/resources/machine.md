# Resource: orbstack_machine

The `orbstack_machine` resource creates and manages Linux machines in OrbStack.

## Example Usage

```hcl
resource "orbstack_machine" "vm" {
  name   = "demo-vm"
  image  = "ubuntu"
  tag    = "24.04"
  
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
| `image` | `string` | No | `"ubuntu"` | The base image/distribution (e.g., ubuntu, debian, alpine) |
| `tag` | `string` | No | - | Optional image tag/version (if supported by OrbStack) |
| `cloud_init` | `string` | No | - | Cloud-init user data passed during creation |
| `cloud_init_file` | `string` | No | - | Path to a cloud-init user data file. Overrides `cloud_init` if both set |
| `validate_image` | `bool` | No | `false` | Validate image (and tag) exists before create; fail fast if unknown |

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

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

- The machine is recreated if any immutable attributes change (image, tag, cloud_init, cloud_init_file)
- Cloud-init data is passed during machine creation and may not be applied if the image doesn't support it
- Use `validate_image = true` to ensure the image exists before attempting to create the machine
- The `cloud_init_file` argument takes precedence over `cloud_init` if both are specified
