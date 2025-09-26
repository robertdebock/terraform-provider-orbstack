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
| `name` | `string` | The name of the image |
| `tag` | `string` | The tag/version of the image |
| `size` | `string` | The size of the image |
| `created` | `string` | When the image was created |
| `description` | `string` | Description of the image |

## Example Output

```hcl
# Example of accessing image information
data "orbstack_images" "available" {}

resource "orbstack_machine" "vm" {
  name = "test-vm"
  image = data.orbstack_images.available.images[0].name
  tag  = data.orbstack_images.available.images[0].tag
}
```

## Notes

- This data source queries OrbStack for available images
- Images are returned in the order provided by OrbStack
- Use this to dynamically select images for machine creation
- The list may change between Terraform runs as images are updated or removed
