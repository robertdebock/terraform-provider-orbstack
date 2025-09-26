package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &K8sStatusDataSource{}
var _ datasource.DataSourceWithConfigure = &K8sStatusDataSource{}

func NewK8sStatusDataSource() datasource.DataSource { return &K8sStatusDataSource{} }

type K8sStatusDataSource struct {
	client *ClientConfig
}

type K8sStatusModel struct {
	Enabled        types.Bool   `tfsdk:"enabled"`
	ExposeServices types.Bool   `tfsdk:"expose_services"`
	Status         types.String `tfsdk:"status"`
	KubeconfigPath types.String `tfsdk:"kubeconfig_path"`
	Nodes          types.List   `tfsdk:"nodes"`
	Version        types.String `tfsdk:"version"`
}

func (d *K8sStatusDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s_status"
}

func (d *K8sStatusDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Get information about the current OrbStack Kubernetes cluster status.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Kubernetes is enabled in OrbStack configuration.",
			},
			"expose_services": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether Kubernetes services are exposed.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the Kubernetes cluster (running, stopped, disabled).",
			},
			"kubeconfig_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the Kubernetes kubeconfig file.",
			},
			"nodes": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "List of Kubernetes node names.",
			},
			"version": schema.StringAttribute{
				Computed:    true,
				Description: "Kubernetes cluster version.",
			},
		},
	}
}

func (d *K8sStatusDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ClientConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *ClientConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *K8sStatusDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data K8sStatusModel

	// Get Kubernetes configuration
	enabled, err := d.isK8sEnabled(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Kubernetes configuration", err.Error())
		return
	}
	data.Enabled = types.BoolValue(enabled)

	exposeServices, err := d.isK8sExposeServices(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Kubernetes expose services", err.Error())
		return
	}
	data.ExposeServices = types.BoolValue(exposeServices)

	// Get Kubernetes status
	status, err := d.getK8sStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kubernetes status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	// Set kubeconfig path
	data.KubeconfigPath = types.StringValue("~/.orbstack/k8s/config.yml")

	// Get nodes if Kubernetes is running
	if status == "running" {
		nodeList, err := d.getK8sNodes(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to get Kubernetes nodes", err.Error())
			return
		}

		// Convert string slice to attr.Value slice
		nodeElements := make([]attr.Value, len(nodeList))
		for i, node := range nodeList {
			nodeElements[i] = types.StringValue(node)
		}
		data.Nodes = types.ListValueMust(types.StringType, nodeElements)
	} else {
		// Empty list when Kubernetes is not running
		data.Nodes = types.ListValueMust(types.StringType, []attr.Value{})
	}

	// Get Kubernetes version
	version, err := d.getK8sVersion(ctx)
	if err != nil {
		// Version might not be available if Kubernetes is not running
		data.Version = types.StringValue("")
	} else {
		data.Version = types.StringValue(version)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *K8sStatusDataSource) isK8sEnabled(ctx context.Context) (bool, error) {
	stdout, _, err := runOrb(ctx, d.client.OrbPath, "config", "get", "k8s.enable")
	if err != nil {
		return false, err
	}

	// Parse the output to get the k8s.enable value
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "k8s.enable:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "k8s.enable:"))
			return value == "true", nil
		}
	}

	return false, nil
}

func (d *K8sStatusDataSource) isK8sExposeServices(ctx context.Context) (bool, error) {
	stdout, _, err := runOrb(ctx, d.client.OrbPath, "config", "get", "k8s.expose_services")
	if err != nil {
		return false, err
	}

	// Parse the output to get the k8s.expose_services value
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "k8s.expose_services:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "k8s.expose_services:"))
			return value == "true", nil
		}
	}

	return false, nil
}

func (d *K8sStatusDataSource) getK8sStatus(ctx context.Context) (string, error) {
	// Check if Kubernetes is running by trying to get nodes
	stdout, _, err := runOrb(ctx, d.client.OrbPath, "run", "kubectl", "get", "nodes", "--no-headers")
	if err != nil {
		// If kubectl fails, Kubernetes is likely not running
		return "stopped", nil
	}

	// If we get output, Kubernetes is running
	if strings.TrimSpace(stdout) != "" {
		return "running", nil
	}

	return "stopped", nil
}

func (d *K8sStatusDataSource) getK8sNodes(ctx context.Context) ([]string, error) {
	stdout, _, err := runOrb(ctx, d.client.OrbPath, "run", "kubectl", "get", "nodes", "--no-headers", "-o", "custom-columns=NAME:.metadata.name")
	if err != nil {
		return nil, err
	}

	var nodes []string
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			nodes = append(nodes, line)
		}
	}

	return nodes, nil
}

func (d *K8sStatusDataSource) getK8sVersion(ctx context.Context) (string, error) {
	stdout, _, err := runOrb(ctx, d.client.OrbPath, "run", "kubectl", "version", "--short", "--client")
	if err != nil {
		return "", err
	}

	// Parse version from output like "Client Version: v1.32.6+orb1"
	lines := strings.Split(stdout, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Client Version:") {
			version := strings.TrimSpace(strings.TrimPrefix(line, "Client Version:"))
			return version, nil
		}
	}

	return "", nil
}
