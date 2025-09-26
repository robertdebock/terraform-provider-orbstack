# orbstack_k8s_status

Get information about the current OrbStack Kubernetes cluster status.

## Example Usage

```hcl
data "orbstack_k8s_status" "cluster" {}

output "k8s_status" {
  value = data.orbstack_k8s_status.cluster.status
}

output "k8s_nodes" {
  value = data.orbstack_k8s_status.cluster.nodes
}
```

## Attribute Reference

The following attributes are exported:

| Name | Type | Description |
|------|------|-------------|
| `enabled` | `bool` | Whether Kubernetes is enabled in OrbStack configuration. |
| `expose_services` | `bool` | Whether Kubernetes services are exposed. |
| `status` | `string` | Current status of the Kubernetes cluster (running, stopped, disabled). |
| `kubeconfig_path` | `string` | Path to the Kubernetes kubeconfig file. |
| `nodes` | `list(string)` | List of Kubernetes node names. |
| `version` | `string` | Kubernetes cluster version. |

## Notes

- The `nodes` list is only populated when the cluster is running.
- The `version` is only available when the cluster is running.
- Use this data source to check cluster status before deploying Kubernetes resources.
