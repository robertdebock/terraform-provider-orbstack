package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var _ provider.Provider = &OrbStackProvider{}

// New returns a new provider instance.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OrbStackProvider{version: version}
	}
}

// OrbStackProvider is the provider implementation.
type OrbStackProvider struct {
	version string
}

// OrbStackProviderModel maps provider schema data to a Go type.
// Fields are optional; defaults are applied where applicable.
type OrbStackProviderModel struct {
	OrbPath           types.String `tfsdk:"orb_path"`
	DefaultUser       types.String `tfsdk:"default_user"`
	DefaultSSHKeyPath types.String `tfsdk:"default_ssh_key_path"`
	CreateTimeout     types.String `tfsdk:"create_timeout"`
	DeleteTimeout     types.String `tfsdk:"delete_timeout"`
}

func (p *OrbStackProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "orbstack"
	resp.Version = p.version
}

func (p *OrbStackProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Terraform provider for managing OrbStack Linux machines using the orb CLI.",
		Attributes: map[string]schema.Attribute{
			"orb_path": schema.StringAttribute{
				Optional:    true,
				Description: "Path to the orb executable. If not set, uses orb in PATH.",
			},
			"default_user": schema.StringAttribute{
				Optional:    true,
				Description: "Default user for SSH metadata (read-only usage).",
			},
			"default_ssh_key_path": schema.StringAttribute{
				Optional:    true,
				Description: "Default SSH public key path for metadata/reporting.",
			},
			"create_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "Timeout for machine creation (e.g., 5m).",
			},
			"delete_timeout": schema.StringAttribute{
				Optional:    true,
				Description: "Timeout for machine deletion (e.g., 5m).",
			},
		},
	}
}

func (p *OrbStackProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data OrbStackProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := &ClientConfig{
		OrbPath:           stringOrDefault(data.OrbPath, "orb"),
		DefaultUser:       stringOrDefault(data.DefaultUser, ""),
		DefaultSSHKeyPath: stringOrDefault(data.DefaultSSHKeyPath, ""),
		CreateTimeout:     stringOrDefault(data.CreateTimeout, "5m"),
		DeleteTimeout:     stringOrDefault(data.DeleteTimeout, "5m"),
	}

	tflog.Debug(ctx, "orbstack provider configured", map[string]any{"orb_path": cfg.OrbPath})

	resp.DataSourceData = cfg
	resp.ResourceData = cfg
}

func (p *OrbStackProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewMachineResource,
		NewConfigResource,
		NewK8sResource,
	}
}

func (p *OrbStackProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewMachineDataSource,
		NewImagesDataSource,
		NewK8sStatusDataSource,
	}
}

// Helper to extract string value or default if null/unknown/empty
func stringOrDefault(v types.String, def string) string {
	if v.IsNull() || v.IsUnknown() {
		return def
	}
	if s := v.ValueString(); s != "" {
		return s
	}
	return def
}
