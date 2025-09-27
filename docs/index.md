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
| `default_user` | `string` | - | Default user for SSH metadata (read-only usage) |
| `default_ssh_key_path` | `string` | - | Default SSH public key path for metadata/reporting |
| `create_timeout` | `string` | `"5m"` | Timeout for machine creation (e.g., 5m) |
| `delete_timeout` | `string` | `"5m"` | Timeout for machine deletion (e.g., 5m) |

## Resources

- [`orbstack_machine`](resources/machine.md) - Create and manage Linux machines
- [`orbstack_config`](resources/config.md) - Manage OrbStack configuration settings
- [`orbstack_k8s`](resources/k8s.md) - Enable/disable Kubernetes cluster

## Data Sources

- [`orbstack_machine`](data-sources/machine.md) - Read information about existing machines
- [`orbstack_images`](data-sources/images.md) - List available container images
- [`orbstack_k8s_status`](data-sources/k8s_status.md) - Get Kubernetes cluster status

## Docker Integration

**Note**: This provider does not include Docker container management resources. OrbStack is fully compatible with the standard Docker API, so you should use the [kreuzwerker/docker provider](https://registry.terraform.io/providers/kreuzwerker/docker/latest) for managing Docker containers.

## Examples

See the [examples/](../examples/) directory for complete usage examples:

- [Ubuntu with cloud-init](../examples/ubuntu-cloudinit/)
- [Debian with specific version](../examples/debian-tag/)
- [Machine sizing and validation](../examples/machine-sizing/)
- [Cloud-init from file](../examples/cloud-init-file/)
- [Kubernetes cluster management](../examples/k8s/)
