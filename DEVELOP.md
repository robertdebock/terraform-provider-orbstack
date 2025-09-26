# Development

## Build

   ```bash
   # Install Go 1.22+
   brew install go

   # Build and store locally
   go build -o ~/dev-plugins/orbstack/terraform-provider-orbstack ./cmd/terraform-provider-orbstack
   
   # Tell Terraform to use the locally built plugin
   cat << EOF > ~/.terraformrc
   provider_installation {
     dev_overrides {
       "robertdebock/orbstack" = "/Users/YOUR_USERNAME/dev-plugins/orbstack"
     }
    direct {}
  }
   ```
