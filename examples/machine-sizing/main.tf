# Discover images (optional filter)
data "orbstack_images" "ubuntu" {
  filter = "ubuntu"
}

resource "orbstack_machine" "sized" {
  name = "sized-vm"
  # Pick the first ubuntu image.
  image = data.orbstack_images.ubuntu.images[0].name

  # Sizing parameters (create-time)
  cpus          = 2
  memory_mib    = 2048
  disk_size_gib = 40

  # Lifecycle state
  power_state = "running" # or "stopped"
}

