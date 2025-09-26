# OrbStack Provider

The OrbStack provider is used to manage OrbStack Linux machines and settings using Terraform.

## Example Usage

```hcl
terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = "~> 1.0"
    }
  }
}

provider "orbstack" {
  # orb_path = "/usr/local/bin/orb"  # Optional: path to orb CLI
}

# Create a Linux machine
resource "orbstack_machine" "vm" {
  name   = "demo-vm"
  image  = "ubuntu:noble"  # Use OS:VERSION format for specific versions
}

# Manage OrbStack settings
resource "orbstack_config" "memory" {
  key   = "memory_mib"
  value = "8192"
}
```

## Requirements

- **OrbStack 2.0.0+** installed on macOS
- **Terraform 1.6+** (recommended)
- **macOS** (Apple Silicon or Intel)

## Provider Configuration

The provider supports the following configuration options:

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `orb_path` | `string` | `"orb"` | Path to the OrbStack CLI executable |

## Resources

- [`orbstack_machine`](resources/machine.md) - Create and manage Linux machines
- [`orbstack_config`](resources/config.md) - Manage OrbStack configuration settings

## Data Sources

- [`orbstack_machine`](data-sources/machine.md) - Read information about existing machines
- [`orbstack_images`](data-sources/images.md) - List available container images

## Examples

See the [examples/](../examples/) directory for complete usage examples:

- [Ubuntu with cloud-init](../examples/ubuntu-cloudinit/)
- [Debian with specific version](../examples/debian-tag/)
- [Machine sizing and validation](../examples/machine-sizing/)
- [Cloud-init from file](../examples/cloud-init-file/)
