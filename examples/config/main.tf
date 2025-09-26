# See all keys with `orb config show`.

# Example: manage global memory (in MiB) and CPU count
resource "orbstack_config" "memory" {
  key   = "memory_mib"
  value = "8192"
}

resource "orbstack_config" "cpu" {
  key   = "cpu"
  value = "8"
}

# Example: toggle Rosetta usage
resource "orbstack_config" "rosetta" {
  key   = "rosetta"
  value = "true"
}


