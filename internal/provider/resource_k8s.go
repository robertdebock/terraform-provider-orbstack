package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &K8sResource{}
var _ resource.ResourceWithConfigure = &K8sResource{}

func NewK8sResource() resource.Resource { return &K8sResource{} }

type K8sResource struct {
	client *ClientConfig
}

type K8sModel struct {
	ID             types.String `tfsdk:"id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	ExposeServices types.Bool   `tfsdk:"expose_services"`
	Status         types.String `tfsdk:"status"`
	KubeconfigPath types.String `tfsdk:"kubeconfig_path"`
}

func (r *K8sResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_k8s"
}

func (r *K8sResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OrbStack Kubernetes cluster enable/disable and configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the Kubernetes configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Enable or disable Kubernetes cluster.",
			},
			"expose_services": schema.BoolAttribute{
				Optional:    true,
				Description: "Expose Kubernetes services. Defaults to true.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the Kubernetes cluster (running, stopped, disabled).",
			},
			"kubeconfig_path": schema.StringAttribute{
				Computed:    true,
				Description: "Path to the Kubernetes kubeconfig file.",
			},
		},
	}
}

func (r *K8sResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*ClientConfig)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *ClientConfig, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *K8sResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data K8sModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	if data.ExposeServices.IsNull() || data.ExposeServices.IsUnknown() {
		data.ExposeServices = types.BoolValue(true)
	}

	// Set ID
	data.ID = types.StringValue("orbstack-k8s")

	// Configure Kubernetes settings
	if err := r.configureK8s(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to configure Kubernetes", err.Error())
		return
	}

	// Start or stop Kubernetes based on enabled setting
	if data.Enabled.ValueBool() {
		if err := r.startK8s(ctx); err != nil {
			resp.Diagnostics.AddError("Failed to start Kubernetes", err.Error())
			return
		}
		data.Status = types.StringValue("running")
	} else {
		if err := r.stopK8s(ctx); err != nil {
			resp.Diagnostics.AddError("Failed to stop Kubernetes", err.Error())
			return
		}
		data.Status = types.StringValue("stopped")
	}

	// Set kubeconfig path
	data.KubeconfigPath = types.StringValue("~/.orbstack/k8s/config.yml")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *K8sResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data K8sModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check current Kubernetes status
	status, err := r.getK8sStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Kubernetes status", err.Error())
		return
	}

	data.Status = types.StringValue(status)

	// Update enabled state based on current configuration
	enabled, err := r.isK8sEnabled(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Kubernetes configuration", err.Error())
		return
	}
	data.Enabled = types.BoolValue(enabled)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *K8sResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data K8sModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Configure Kubernetes settings
	if err := r.configureK8s(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to configure Kubernetes", err.Error())
		return
	}

	// Start or stop Kubernetes based on enabled setting
	if data.Enabled.ValueBool() {
		if err := r.startK8s(ctx); err != nil {
			resp.Diagnostics.AddError("Failed to start Kubernetes", err.Error())
			return
		}
		data.Status = types.StringValue("running")
	} else {
		if err := r.stopK8s(ctx); err != nil {
			resp.Diagnostics.AddError("Failed to stop Kubernetes", err.Error())
			return
		}
		data.Status = types.StringValue("stopped")
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *K8sResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data K8sModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Stop Kubernetes and disable it
	if err := r.stopK8s(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to stop Kubernetes", err.Error())
		return
	}

	// Disable Kubernetes
	if err := r.setK8sEnabled(ctx, false); err != nil {
		resp.Diagnostics.AddError("Failed to disable Kubernetes", err.Error())
		return
	}
}

func (r *K8sResource) configureK8s(ctx context.Context, data K8sModel) error {
	// Set k8s.enable
	if err := r.setK8sEnabled(ctx, data.Enabled.ValueBool()); err != nil {
		return err
	}

	// Set k8s.expose_services
	if err := r.setK8sExposeServices(ctx, data.ExposeServices.ValueBool()); err != nil {
		return err
	}

	return nil
}

func (r *K8sResource) setK8sEnabled(ctx context.Context, enabled bool) error {
	value := "false"
	if enabled {
		value = "true"
	}
	_, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "k8s.enable", value)
	return err
}

func (r *K8sResource) setK8sExposeServices(ctx context.Context, expose bool) error {
	value := "false"
	if expose {
		value = "true"
	}
	_, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "k8s.expose_services", value)
	return err
}

func (r *K8sResource) startK8s(ctx context.Context) error {
	_, _, err := runOrb(ctx, r.client.OrbPath, "start", "k8s")
	return err
}

func (r *K8sResource) stopK8s(ctx context.Context) error {
	_, _, err := runOrb(ctx, r.client.OrbPath, "stop", "k8s")
	return err
}

func (r *K8sResource) getK8sStatus(ctx context.Context) (string, error) {
	// Check if Kubernetes is running by trying to get nodes
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "run", "kubectl", "get", "nodes", "--no-headers")
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

func (r *K8sResource) isK8sEnabled(ctx context.Context) (bool, error) {
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "config", "get", "k8s.enable")
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
