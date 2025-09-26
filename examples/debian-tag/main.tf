resource "orbstack_machine" "debian_tagged" {
  name  = "debian-12-vm"
  # Specify the tag as part of the image (e.g., bookworm)
  image = "debian:bookworm"
}


