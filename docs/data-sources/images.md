# Data Source: orbstack_images

The `orbstack_images` data source lists available container images that can be used with OrbStack machines.

## Example Usage

```hcl
data "orbstack_images" "available" {}

output "ubuntu_images" {
  value = [
    for img in data.orbstack_images.available.images : img.name
    if startswith(img.name, "ubuntu")
  ]
}

output "all_images" {
  value = data.orbstack_images.available.images
}
```

## Attributes Reference

The following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `images` | `list(object)` | List of available images |

Each image in the `images` list has the following attributes:

| Name | Type | Description |
|------|------|-------------|
| `name` | `string` | The name of the image (may include OS:VERSION format) |
| `display_name` | `string` | Display name if provided by OrbStack |
| `default` | `bool` | Whether this image is the default (if detectable) |

## Example Output

```hcl
# Example of accessing image information
data "orbstack_images" "available" {}

resource "orbstack_machine" "vm" {
  name = "test-vm"
  image = data.orbstack_images.available.images[0].name
}
```

## Notes

- This data source queries OrbStack for available images
- Images are returned in the order provided by OrbStack
- Use this to dynamically select images for machine creation
- The list may change between Terraform runs as images are updated or removed
