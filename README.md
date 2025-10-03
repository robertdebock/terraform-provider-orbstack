# Terraform Provider for OrbStack

Manage OrbStack Linux machines and settings in Terraform using the `orb` CLI.

## v3.0.0

- Removed the `orbstack_images` data source. The modern OrbStack CLI no longer exposes an images list; specify tagged images explicitly (e.g., `ubuntu:noble`, `debian:bookworm`).
- Examples updated to use explicit tagged images and disable image validation.

## Note: Docker Support

This provider does **not** include Docker container management resources. OrbStack is fully compatible with the standard Docker API, so use the [kreuzwerker/docker provider](https://registry.terraform.io/providers/kreuzwerker/docker/latest) for managing Docker containers.

See the [Terraform registry](https://registry.terraform.io/providers/robertdebock/orbstack/latest) for more information.
