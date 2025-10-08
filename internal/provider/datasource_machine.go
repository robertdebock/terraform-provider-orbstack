package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &MachineDataSource{}
var _ datasource.DataSourceWithConfigure = &MachineDataSource{}

func NewMachineDataSource() datasource.DataSource { return &MachineDataSource{} }

type MachineDataSource struct {
	client *ClientConfig
}

type MachineDataSourceModel struct {
	Name           types.String `tfsdk:"name"`
	ID             types.String `tfsdk:"id"`
	IPAddress      types.String `tfsdk:"ip_address"`
	Status         types.String `tfsdk:"status"`
	SSHHost        types.String `tfsdk:"ssh_host"`
	SSHPort        types.Int64  `tfsdk:"ssh_port"`
	CreatedAt      types.String `tfsdk:"created_at"`
	DefaultMachine types.Bool   `tfsdk:"default_machine"`
}

func (d *MachineDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (d *MachineDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Read an OrbStack machine by name.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Machine name.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal identifier (same as name).",
			},
			"ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "Machine IP address.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status.",
			},
			"ssh_host": schema.StringAttribute{
				Computed:    true,
				Description: "SSH host.",
			},
			"ssh_port": schema.Int64Attribute{
				Computed:    true,
				Description: "SSH port.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation timestamp.",
			},
			"default_machine": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether this machine is the current default machine.",
			},
		},
	}
}

func (d *MachineDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, _ := req.ProviderData.(*ClientConfig)
	d.client = cfg
}

func (d *MachineDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data MachineDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := d.client
	if cfg == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client configuration")
		return
	}

	name := data.Name.ValueString()
	model, diags := readUntilReady(ctx, cfg, name, cfg.CreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model == nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"machine not found",
			"No machine found with the given name.",
		)
		return
	}

	data.ID = types.StringValue(name)
	data.IPAddress = model.IPAddress
	data.Status = model.Status
	data.SSHHost = model.SSHHost
	data.SSHPort = model.SSHPort
	data.CreatedAt = model.CreatedAt

	// Check if this machine is the current default
	isDefault, diags := d.isDefaultMachine(ctx, cfg, name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.DefaultMachine = types.BoolValue(isDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// isDefaultMachine checks if the given machine is the current default
func (d *MachineDataSource) isDefaultMachine(ctx context.Context, cfg *ClientConfig, machineName string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	out, _, err := runOrb(ctx, cfg.OrbPath, "default")
	if err != nil {
		diags.AddError("orb default failed", err.Error())
		return false, diags
	}

	currentDefault := strings.TrimSpace(out)
	return currentDefault == machineName, diags
}
