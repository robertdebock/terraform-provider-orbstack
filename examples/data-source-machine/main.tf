resource "orbstack_machine" "example" {
  name  = "example-ds-vm"
  image = "ubuntu"
}

data "orbstack_machine" "example" {
  name       = orbstack_machine.example.name
}


