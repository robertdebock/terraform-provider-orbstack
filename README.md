# Terraform Provider for OrbStack

Manage OrbStack Linux machines and settings in Terraform using the `orb` CLI.

- Requires OrbStack 2.0.0+ and Terraform 1.6+ (recommended)
- Runs on macOS on the same host where OrbStack is installed

## Features (MVP)

- Resource `orbstack_machine`: create, read, delete a Linux machine
- Resource `orbstack_config`: manage a single OrbStack config key via `orb config`
- Recreate on change for immutable attributes

## Example

   ```hcl
   terraform {
     required_providers {
       orbstack = {
         source  = "robertdebock/orbstack"
         version = ">= 1.0.0"
       }
     }
   }

   provider "orbstack" {
     # orb_path = "/usr/local/bin/orb"
   }

   resource "orbstack_machine" "vm" {
     name   = "demo-vm"
     image  = "ubuntu"   # default: ubuntu
     # optional future: tag = "24.04"
   }

   # Manage OrbStack settings
   resource "orbstack_config" "memory" {
     key   = "memory_mib"
     value = "8192"
   }
   ```

Arguments for `orbstack_machine`:

| argument | required? | default value | description |
|----------|-----------|---------------|-------------|
| `name` | Yes | - | The name of the machine |
| `image` | No | ubuntu | The base image/distribution (e.g., ubuntu, debian, alpine) |
| `tag` | No | - | Optional image tag/version (if supported by OrbStack) |
| `cloud_init` | No | - | Cloud-init user data passed during creation (best-effort) |
| `cloud_init_file` | No | - | Path to a cloud-init user data file. Overrides cloud_init if both set |
| `validate_image` | No | - | Validate image (and tag) exists before create; fail fast if unknown |

### More examples

- `examples/ubuntu-cloudinit`: Ubuntu VM with a cloud-init user-data
- `examples/debian-tag`: Debian VM using a tagged image (e.g., `debian:bookworm`)
- `examples/alpine-minimal`: Minimal Alpine VM
- `examples/data-source-machine`: Read an existing machine by name
- `examples/machine-sizing`: Image discovery and sizing (cpus, memory, disk) with power_state
- `examples/cloud-init-file`: Use cloud_init_file and set_password
- `examples/validate-image`: Validate image exists before create
