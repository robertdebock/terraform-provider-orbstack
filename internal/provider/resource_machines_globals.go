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

var _ resource.Resource = &MachinesGlobalsResource{}
var _ resource.ResourceWithConfigure = &MachinesGlobalsResource{}

func NewMachinesGlobalsResource() resource.Resource { return &MachinesGlobalsResource{} }

// MachinesGlobalsResource manages global defaults for all machines.
type MachinesGlobalsResource struct {
	client *ClientConfig
}

type MachinesGlobalsModel struct {
	ID               types.String `tfsdk:"id"`
	ExposePortsToLan types.Bool   `tfsdk:"expose_ports_to_lan"`
	ForwardPorts     types.Bool   `tfsdk:"forward_ports"`
	Status           types.String `tfsdk:"status"`
}

func (r *MachinesGlobalsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine_config"
}

func (r *MachinesGlobalsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages global defaults for all machines.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for the machines globals configuration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"expose_ports_to_lan": schema.BoolAttribute{
				Optional:    true,
				Description: "Expose machine ports to LAN by default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"forward_ports": schema.BoolAttribute{
				Optional:    true,
				Description: "Automatically forward ports for machines.",
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

func (r *MachinesGlobalsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MachinesGlobalsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data MachinesGlobalsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	data.ID = types.StringValue("orbstack-machines-globals")

	if data.ExposePortsToLan.IsNull() || data.ExposePortsToLan.IsUnknown() {
		data.ExposePortsToLan = types.BoolValue(true)
	}
	if data.ForwardPorts.IsNull() || data.ForwardPorts.IsUnknown() {
		data.ForwardPorts = types.BoolValue(true)
	}

	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply machines globals", err.Error())
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

func (r *MachinesGlobalsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data MachinesGlobalsModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.readConfig(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Failed to read machines globals", err.Error())
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

func (r *MachinesGlobalsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data MachinesGlobalsModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.applyConfig(ctx, data); err != nil {
		resp.Diagnostics.AddError("Failed to apply machines globals", err.Error())
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

func (r *MachinesGlobalsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.State.RemoveResource(ctx)
}

func (r *MachinesGlobalsResource) applyConfig(ctx context.Context, data MachinesGlobalsModel) error {
	if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "machines.expose_ports_to_lan", boolToString(data.ExposePortsToLan.ValueBool())); err != nil {
		return err
	}
	if _, _, err := runOrb(ctx, r.client.OrbPath, "config", "set", "machines.forward_ports", boolToString(data.ForwardPorts.ValueBool())); err != nil {
		return err
	}
	return nil
}

func (r *MachinesGlobalsResource) readConfig(ctx context.Context, data *MachinesGlobalsModel) error {
	out, _, err := runOrb(ctx, r.client.OrbPath, "config", "show")
	if err != nil {
		return err
	}
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "machines.expose_ports_to_lan:") {
			data.ExposePortsToLan = types.BoolValue(strings.TrimSpace(strings.TrimPrefix(line, "machines.expose_ports_to_lan:")) == "true")
		} else if strings.HasPrefix(line, "machines.forward_ports:") {
			data.ForwardPorts = types.BoolValue(strings.TrimSpace(strings.TrimPrefix(line, "machines.forward_ports:")) == "true")
		}
	}
	return nil
}

func (r *MachinesGlobalsResource) handleRestartIfNeeded(ctx context.Context) error {
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

func (r *MachinesGlobalsResource) getOrbStackStatus(ctx context.Context) (string, error) {
	stdout, _, err := runOrb(ctx, r.client.OrbPath, "status")
	if err != nil {
		return "stopped", nil
	}
	if strings.TrimSpace(stdout) == "Running" {
		return "running", nil
	}
	return "stopped", nil
}
