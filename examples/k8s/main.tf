terraform {
  required_providers {
    orbstack = {
      source  = "robertdebock/orbstack"
      version = "~> 1.0"
    }
  }
}

provider "orbstack" {
  # orb_path = "/usr/local/bin/orb"  # Optional: path to orb CLI
}

# Enable Kubernetes cluster
resource "orbstack_k8s" "cluster" {
  enabled         = true
  expose_services = true
}

# Get Kubernetes status information
data "orbstack_k8s_status" "cluster" {}

# Output Kubernetes information
output "k8s_enabled" {
  description = "Whether Kubernetes is enabled"
  value       = data.orbstack_k8s_status.cluster.enabled
}

output "k8s_status" {
  description = "Current Kubernetes cluster status"
  value       = data.orbstack_k8s_status.cluster.status
}

output "k8s_nodes" {
  description = "List of Kubernetes nodes"
  value       = data.orbstack_k8s_status.cluster.nodes
}

output "k8s_version" {
  description = "Kubernetes version"
  value       = data.orbstack_k8s_status.cluster.version
}

output "kubeconfig_path" {
  description = "Path to kubeconfig file"
  value       = data.orbstack_k8s_status.cluster.kubeconfig_path
}
