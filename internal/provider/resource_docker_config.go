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

var _ resource.Resource = &DockerConfigResource{}
var _ resource.ResourceWithConfigure = &DockerConfigResource{}

func NewDockerConfigResource() resource.Resource { return &DockerConfigResource{} }

// DockerConfigResource manages Docker engine configuration settings.
type DockerConfigResource struct {
	client *ClientConfig
}

type DockerConfigModel struct {
	ID               types.String `tfsdk:"id"`
	SetContext       types.Bool   `tfsdk:"set_context"`
	ExposePortsToLan types.Bool   `tfsdk:"expose_ports_to_lan"`
	NodeName         types.String `tfsdk:"node_name"`
	Status           types.String `tfsdk:"status"`
	DockerEndpoint   types.String `tfsdk:"docker_endpoint"`
	ContextActive    types.Bool   `tfsdk:"context_active"`
}

func (r *DockerConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_docker_config"
}

func (r *DockerConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages Docker engine configuration settings including context management and port exposure.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the Docker configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"set_context": schema.BoolAttribute{
				Optional:    true,
				Description: "Automatically set Docker context to orbstack.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"expose_ports_to_lan": schema.BoolAttribute{
				Optional:    true,
				Description: "Expose Docker ports to LAN.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"node_name": schema.StringAttribute{
				Optional:    true,
				Description: "Docker node name.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of Docker engine (running/stopped).",
			},
			"docker_endpoint": schema.StringAttribute{
				Computed:    true,
				Description: "Docker daemon endpoint.",
			},
			"context_active": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the orbstack Docker context is currently active.",
			},
		},
	}
}

func (r *DockerConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DockerConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data DockerConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID
	data.ID = types.StringValue("orbstack-docker-config")

	// Set default values for optional fields
	if data.SetContext.IsNull() || data.SetContext.IsUnknown() {
		data.SetContext = types.BoolValue(true)
	}
	if data.ExposePortsToLan.IsNull() || data.ExposePortsToLan.IsUnknown() {
		data.ExposePortsToLan = types.BoolValue(true)
	}
	if data.NodeName.IsNull() || data.NodeName.IsUnknown() {
		data.NodeName = types.StringValue("orbstack")
	}

	// Apply configuration
	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply Docker configuration", err.Error())
		return
	}

	// Check if restart is needed and handle it
	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}

	// Get current status and computed values
	status, err := r.getDockerStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	// Get Docker endpoint
	endpoint, err := r.getDockerEndpoint(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker endpoint", err.Error())
		return
	}
	data.DockerEndpoint = types.StringValue(endpoint)

	// Check if context is active
	contextActive, err := r.isContextActive(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Docker context", err.Error())
		return
	}
	data.ContextActive = types.BoolValue(contextActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data DockerConfigModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current configuration
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to read Docker configuration", err.Error())
		return
	}

	// Get current status and computed values
	status, err := r.getDockerStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	// Get Docker endpoint
	endpoint, err := r.getDockerEndpoint(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker endpoint", err.Error())
		return
	}
	data.DockerEndpoint = types.StringValue(endpoint)

	// Check if context is active
	contextActive, err := r.isContextActive(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Docker context", err.Error())
		return
	}
	data.ContextActive = types.BoolValue(contextActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data DockerConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply configuration
	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply Docker configuration", err.Error())
		return
	}

	// Check if restart is needed and handle it
	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}

	// Get current status and computed values
	status, err := r.getDockerStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	// Get Docker endpoint
	endpoint, err := r.getDockerEndpoint(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get Docker endpoint", err.Error())
		return
	}
	data.DockerEndpoint = types.StringValue(endpoint)

	// Check if context is active
	contextActive, err := r.isContextActive(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to check Docker context", err.Error())
		return
	}
	data.ContextActive = types.BoolValue(contextActive)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *DockerConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// For now, we don't reset configuration on delete
	// This could be enhanced to reset to defaults if needed
	resp.State.RemoveResource(ctx)
}

// applyConfig applies all Docker configuration settings
func (r *DockerConfigResource) applyConfig(ctx context.Context, data DockerConfigModel) error {
	configs := map[string]string{
		"docker.set_context":         fmt.Sprintf("%t", data.SetContext.ValueBool()),
		"docker.expose_ports_to_lan": fmt.Sprintf("%t", data.ExposePortsToLan.ValueBool()),
		"docker.node_name":           data.NodeName.ValueString(),
	}

	for key, value := range configs {
		if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// readConfig reads current Docker configuration from OrbStack
func (r *DockerConfigResource) readConfig(ctx context.Context, data *DockerConfigModel) error {
	// Get all config
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "config", "show")
	if err != nil {
		return err
	}

	// Parse the output
	configs := make(map[string]string)
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			configs[parts[0]] = strings.TrimSpace(parts[1])
		}
	}

	// Set values from config
	if val, ok := configs["docker.set_context"]; ok {
		data.SetContext = types.BoolValue(val == "true")
	}
	if val, ok := configs["docker.expose_ports_to_lan"]; ok {
		data.ExposePortsToLan = types.BoolValue(val == "true")
	}
	if val, ok := configs["docker.node_name"]; ok {
		data.NodeName = types.StringValue(val)
	}

	return nil
}

// handleRestartIfNeeded checks if restart is needed and handles it
func (r *DockerConfigResource) handleRestartIfNeeded(ctx context.Context) error {
	// Check if OrbStack is running
	status, err := r.getDockerStatus(ctx)
	if err != nil {
		return err
	}

	if status == "running" {
		// Stop OrbStack
		if _, _, err := runOrb(ctx, r.client.OrbPath, "stop"); err != nil {
			return fmt.Errorf("failed to stop OrbStack: %w", err)
		}

		// Start OrbStack
		if _, _, err := runOrb(ctx, r.client.OrbPath, "start"); err != nil {
			return fmt.Errorf("failed to start OrbStack: %w", err)
		}
	}

	return nil
}

// getDockerStatus gets the current status of Docker engine
func (r *DockerConfigResource) getDockerStatus(ctx context.Context) (string, error) {
	// Check if OrbStack is running (which means Docker is available)
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "status")
	if err != nil {
		return "stopped", nil
	}

	status := strings.TrimSpace(stdout)
	if status == "Running" {
		return "running", nil
	}
	return "stopped", nil
}

// getDockerEndpoint gets the Docker daemon endpoint
func (r *DockerConfigResource) getDockerEndpoint(ctx context.Context) (string, error) {
	// Check if we can get Docker info
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "run", "docker", "info", "--format", "{{.ServerVersion}}")
	if err != nil {
		return "unix:///Users/username/.orbstack/run/docker.sock", nil
	}

	if strings.TrimSpace(stdout) != "" {
		return "unix:///Users/username/.orbstack/run/docker.sock", nil
	}

	return "unix:///Users/username/.orbstack/run/docker.sock", nil
}

// isContextActive checks if the orbstack Docker context is currently active
func (r *DockerConfigResource) isContextActive(ctx context.Context) (bool, error) {
	// Check current Docker context
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "run", "docker", "context", "ls", "--format", "{{.Name}}")
	if err != nil {
		return false, nil
	}

	// Check if orbstack context is in the list and active
	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	for _, line := range lines {
		if strings.TrimSpace(line) == "orbstack" {
			return true, nil
		}
	}

	return false, nil
}
