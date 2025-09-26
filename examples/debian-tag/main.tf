resource "orbstack_machine" "debian_tagged" {
  name = "debian-12-vm"
  # Specify the OS:VERSION format directly in the image field
  image = "debian:bookworm"
}


