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

var _ resource.Resource = &NetworkConfigResource{}
var _ resource.ResourceWithConfigure = &NetworkConfigResource{}

func NewNetworkConfigResource() resource.Resource { return &NetworkConfigResource{} }

// NetworkConfigResource manages global OrbStack network configuration.
type NetworkConfigResource struct {
	client *ClientConfig
}

type NetworkConfigModel struct {
	ID            types.String `tfsdk:"id"`
	IPv4Subnet    types.String `tfsdk:"ipv4_subnet"`
	BridgeEnabled types.Bool   `tfsdk:"bridge_enabled"`
	ExposeSSHPort types.Bool   `tfsdk:"expose_ssh_port"`
	Status        types.String `tfsdk:"status"`
}

func (r *NetworkConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_config"
}

func (r *NetworkConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages OrbStack network configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the network configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ipv4_subnet": schema.StringAttribute{
				Optional:    true,
				Description: "IPv4 subnet in CIDR notation (e.g., 192.168.200.0/24).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bridge_enabled": schema.BoolAttribute{
				Optional:    true,
				Description: "Enable or disable network bridge.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"expose_ssh_port": schema.BoolAttribute{
				Optional:    true,
				Description: "Expose SSH port on the host.",
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

func (r *NetworkConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, ok := req.ProviderData.(*ClientConfig)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *ClientConfig, got: %T", req.ProviderData))
		return
	}
	r.client = cfg
}

func (r *NetworkConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data NetworkConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("orbstack-network-config")

	if data.BridgeEnabled.IsNull() || data.BridgeEnabled.IsUnknown() {
		data.BridgeEnabled = types.BoolValue(true)
	}
	if data.ExposeSSHPort.IsNull() || data.ExposeSSHPort.IsUnknown() {
		data.ExposeSSHPort = types.BoolValue(true)
	}

	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply network configuration", err.Error())
		return
	}

	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}

	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status", err.Error())
		return
	}
	data.Status = types.StringValue(status)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data NetworkConfigModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to read network configuration", err.Error())
		return
	}
	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status", err.Error())
		return
	}
	data.Status = types.StringValue(status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data NetworkConfigModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply network configuration", err.Error())
		return
	}
	if err := r.handleRestartIfNeeded(ctx); err != nil {
		resp.Diagnostics.AddError("Failed to restart OrbStack", err.Error())
		return
	}
	status, err := r.getOrbStackStatus(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to get status", err.Error())
		return
	}
	data.Status = types.StringValue(status)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *NetworkConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *NetworkConfigResource) applyConfig(ctx context.Context, data NetworkConfigModel) error {
	if s := strings.TrimSpace(data.IPv4Subnet.ValueString()); s != "" {
		if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "network.subnet4", s); err != nil {
			return err
		}
	}
	if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "network_bridge", boolToString(data.BridgeEnabled.ValueBool())); err != nil {
		return err
	}
	if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "ssh.expose_port", boolToString(data.ExposeSSHPort.ValueBool())); err != nil {
		return err
	}
	return nil
}

func (r *NetworkConfigResource) readConfig(ctx context.Context, data *NetworkConfigModel) error {
	out, _, err := runOrb(ctx, r.client.OrbPath, "config", "show")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "network.subnet4:") {
			data.IPv4Subnet = types.StringValue(strings.TrimSpace(strings.TrimPrefix(line, "network.subnet4:")))
		} else if strings.HasPrefix(line, "network_bridge:") {
			data.BridgeEnabled = types.BoolValue(strings.TrimSpace(strings.TrimPrefix(line, "network_bridge:")) == "true")
		} else if strings.HasPrefix(line, "ssh.expose_port:") {
			data.ExposeSSHPort = types.BoolValue(strings.TrimSpace(strings.TrimPrefix(line, "ssh.expose_port:")) == "true")
		}
	}
	return nil
}

func (r *NetworkConfigResource) handleRestartIfNeeded(ctx context.Context) error {
	// Same approach as other resources
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "status")
	if err != nil {
		return nil
	}
	if strings.TrimSpace(stdout) == "Running" {
		if _, _, err := runOrb(ctx, r.client.OrbPath, "stop"); err != nil {
			return err
		}
		if _, _, err := runOrb(ctx, r.client.OrbPath, "start"); err != nil {
			return err
		}
	}
	return nil
}

func (r *NetworkConfigResource) getOrbStackStatus(ctx context.Context) (string, error) {
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "status")
	if err != nil {
		return "stopped", nil
	}
	if strings.TrimSpace(stdout) == "Running" {
		return "running", nil
	}
	return "stopped", nil
}

func boolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}
