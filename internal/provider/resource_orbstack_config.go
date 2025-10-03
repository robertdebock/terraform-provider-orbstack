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

var _ resource.Resource = &OrbStackConfigResource{}
var _ resource.ResourceWithConfigure = &OrbStackConfigResource{}

func NewOrbStackConfigResource() resource.Resource { return &OrbStackConfigResource{} }

// OrbStackConfigResource manages global OrbStack application configuration.
type OrbStackConfigResource struct {
	client *ClientConfig
}

type OrbStackConfigModel struct {
	ID             types.String `tfsdk:"id"`
	CPU            types.Int64  `tfsdk:"cpu"`
	MemoryMib      types.Int64  `tfsdk:"memory_mib"`
	StartAtLogin   types.Bool   `tfsdk:"start_at_login"`
	PauseOnSleep   types.Bool   `tfsdk:"pause_on_sleep"`
	RosettaEnabled types.Bool   `tfsdk:"rosetta_enabled"`
	SetupUserAdmin types.Bool   `tfsdk:"setup_user_admin"`
	Status         types.String `tfsdk:"status"`
}

func (r *OrbStackConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *OrbStackConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages global OrbStack application configuration including CPU, memory, and system behavior settings.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the OrbStack configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
				Required:    true,
				Description: "Number of CPUs to allocate to OrbStack.",
			},
			"memory_mib": schema.Int64Attribute{
				Required:    true,
				Description: "Memory allocation in MiB.",
			},
			"start_at_login": schema.BoolAttribute{
				Optional:    true,
				Description: "Start OrbStack automatically at login.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"pause_on_sleep": schema.BoolAttribute{
				Optional:    true,
				Description: "Pause OrbStack when system goes to sleep.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"rosetta_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable Rosetta 2 for Apple Silicon compatibility.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"setup_user_admin": schema.BoolAttribute{
				Optional:    true,
				Description: "Use admin user for setup operations.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of OrbStack (running/stopped).",
			},
		},
	}
}

func (r *OrbStackConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrbStackConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data OrbStackConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set ID
	data.ID = types.StringValue("orbstack-config")

	// Set default values for optional fields
	if data.StartAtLogin.IsNull() || data.StartAtLogin.IsUnknown() {
		data.StartAtLogin = types.BoolValue(false)
	}
	if data.PauseOnSleep.IsNull() || data.PauseOnSleep.IsUnknown() {
		data.PauseOnSleep = types.BoolValue(true)
	}
	if data.RosettaEnabled.IsNull() || data.RosettaEnabled.IsUnknown() {
		data.RosettaEnabled = types.BoolValue(true)
	}
	if data.SetupUserAdmin.IsNull() || data.SetupUserAdmin.IsUnknown() {
		data.SetupUserAdmin = types.BoolValue(true)
	}

	// Apply configuration
	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply OrbStack configuration", err.Error())
		return
	}

	// Check if restart is needed and handle it
	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}

	// Get current status
	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get OrbStack status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrbStackConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data OrbStackConfigModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read current configuration
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to read OrbStack configuration", err.Error())
		return
	}

	// Get current status
	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get OrbStack status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrbStackConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data OrbStackConfigModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Apply configuration
	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply OrbStack configuration", err.Error())
		return
	}

	// Check if restart is needed and handle it
	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}

	// Get current status
	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get OrbStack status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *OrbStackConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// For now, we don't reset configuration on delete
	// This could be enhanced to reset to defaults if needed
	resp.State.RemoveResource(ctx)
}

// applyConfig applies all configuration settings
func (r *OrbStackConfigResource) applyConfig(ctx context.Context, data OrbStackConfigModel) error {
	configs := map[string]string{
		"cpu":                  fmt.Sprintf("%d", data.CPU.ValueInt64()),
		"memory_mib":           fmt.Sprintf("%d", data.MemoryMib.ValueInt64()),
		"app.start_at_login":   fmt.Sprintf("%t", data.StartAtLogin.ValueBool()),
		"power.pause_in_sleep": fmt.Sprintf("%t", data.PauseOnSleep.ValueBool()),
		"rosetta":              fmt.Sprintf("%t", data.RosettaEnabled.ValueBool()),
		"setup.use_admin":      fmt.Sprintf("%t", data.SetupUserAdmin.ValueBool()),
	}

	for key, value := range configs {
		if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// readConfig reads current configuration from OrbStack
func (r *OrbStackConfigResource) readConfig(ctx context.Context, data *OrbStackConfigModel) error {
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
	if val, ok := configs["cpu"]; ok {
		if cpu, err := parseInt64(val); err == nil {
			data.CPU = types.Int64Value(cpu)
		}
	}
	if val, ok := configs["memory_mib"]; ok {
		if mem, err := parseInt64(val); err == nil {
			data.MemoryMib = types.Int64Value(mem)
		}
	}
	if val, ok := configs["app.start_at_login"]; ok {
		data.StartAtLogin = types.BoolValue(val == "true")
	}
	if val, ok := configs["power.pause_in_sleep"]; ok {
		data.PauseOnSleep = types.BoolValue(val == "true")
	}
	if val, ok := configs["rosetta"]; ok {
		data.RosettaEnabled = types.BoolValue(val == "true")
	}
	if val, ok := configs["setup.use_admin"]; ok {
		data.SetupUserAdmin = types.BoolValue(val == "true")
	}

	return nil
}

// handleRestartIfNeeded checks if restart is needed and handles it
func (r *OrbStackConfigResource) handleRestartIfNeeded(ctx context.Context) error {
	// Check if OrbStack is running
	status, err := r.getOrbStackStatus(ctx)
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

// getOrbStackStatus gets the current status of OrbStack
func (r *OrbStackConfigResource) getOrbStackStatus(ctx context.Context) (string, error) {
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

// parseInt64 parses a string to int64
func parseInt64(s string) (int64, error) {
	// Simple implementation - could be enhanced with proper error handling
	var result int64
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
