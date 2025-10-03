terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = ">= 2.0.0"
    }
  }
}

provider "orbstack" {}

# Example: manage global settings
resource "orbstack_config" "main" {
  cpu              = 8
  memory_mib       = 8192
  start_at_login   = false
  pause_on_sleep   = true
  rosetta_enabled  = true
  setup_user_admin = true
}


