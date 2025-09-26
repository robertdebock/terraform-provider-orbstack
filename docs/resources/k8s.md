# orbstack_k8s

Manages OrbStack Kubernetes cluster enable/disable and configuration.

## Example Usage

```hcl
resource "orbstack_k8s" "cluster" {
  enabled         = true
  expose_services = true
}
```

## Argument Reference

The following arguments are supported:

| Name | Type | Required | Default | Description |
|------|------|----------|---------|-------------|
| `enabled` | `bool` | Yes | - | Enable or disable Kubernetes cluster. |
| `expose_services` | `bool` | No | `true` | Expose Kubernetes services. |

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `id` | `string` | Unique identifier for the Kubernetes configuration. |
| `status` | `string` | Current status of the Kubernetes cluster (running, stopped, disabled). |
| `kubeconfig_path` | `string` | Path to the Kubernetes kubeconfig file. |

## Notes

- Enabling Kubernetes requires OrbStack to be restarted to apply changes.
- The kubeconfig file is located at `~/.orbstack/k8s/config.yml`.
- Use `kubectl` to manage Kubernetes resources within the cluster.
- The cluster context is named "orbstack" in kubectl.
