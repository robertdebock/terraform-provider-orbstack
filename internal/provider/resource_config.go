package provider

import (
    "context"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/diag"
    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &ConfigResource{}
var _ resource.ResourceWithConfigure = &ConfigResource{}

func NewConfigResource() resource.Resource { return &ConfigResource{} }

// ConfigResource manages a single OrbStack configuration key via `orb config`.
type ConfigResource struct {
    client *ClientConfig
}

type ConfigModel struct {
    ID    types.String `tfsdk:"id"`
    Key   types.String `tfsdk:"key"`
    Value types.String `tfsdk:"value"`
}

func (r *ConfigResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_config"
}

func (r *ConfigResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "Manage a single OrbStack configuration option via `orb config get/set`.\nNote: delete does not reset to default; see README.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:    true,
                Description: "Internal ID, equals key.",
            },
            "key": schema.StringAttribute{
                Required:    true,
                Description: "Configuration key (see `orb config show`).",
            },
            "value": schema.StringAttribute{
                Required:    true,
                Description: "Desired value as string (e.g., '8192', 'true').",
            },
        },
    }
}

func (r *ConfigResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    cfg, _ := req.ProviderData.(*ClientConfig)
    r.client = cfg
}

func (r *ConfigResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan ConfigModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) 
    if resp.Diagnostics.HasError() {
        return
    }

    cfg := r.client
    if cfg == nil {
        resp.Diagnostics.AddError("provider not configured", "missing client configuration")
        return
    }

    key := strings.TrimSpace(plan.Key.ValueString())
    val := strings.TrimSpace(plan.Value.ValueString())

    if key == "" {
        resp.Diagnostics.AddAttributeError(path.Root("key"), "invalid key", "key must not be empty")
        return
    }

    _, stderr, err := runOrb(ctx, cfg.OrbPath, "config", "set", key, val)
    if err != nil {
        resp.Diagnostics.AddError("failed to set config", stderr)
        return
    }

    // Read back to confirm
    got, d := readConfig(ctx, cfg, key)
    resp.Diagnostics.Append(d...)
    if resp.Diagnostics.HasError() {
        return
    }

    plan.ID = types.StringValue(key)
    plan.Value = types.StringValue(got)

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state ConfigModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...) 
    if resp.Diagnostics.HasError() {
        return
    }

    cfg := r.client
    if cfg == nil {
        resp.Diagnostics.AddError("provider not configured", "missing client configuration")
        return
    }

    key := strings.TrimSpace(state.Key.ValueString())
    if key == "" {
        resp.State.RemoveResource(ctx)
        return
    }

    got, d := readConfig(ctx, cfg, key)
    resp.Diagnostics.Append(d...)
    if resp.Diagnostics.HasError() {
        return
    }

    state.ID = types.StringValue(key)
    state.Value = types.StringValue(got)

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *ConfigResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan ConfigModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...) 
    if resp.Diagnostics.HasError() {
        return
    }

    cfg := r.client
    if cfg == nil {
        resp.Diagnostics.AddError("provider not configured", "missing client configuration")
        return
    }

    key := strings.TrimSpace(plan.Key.ValueString())
    val := strings.TrimSpace(plan.Value.ValueString())

    _, stderr, err := runOrb(ctx, cfg.OrbPath, "config", "set", key, val)
    if err != nil {
        resp.Diagnostics.AddError("failed to set config", stderr)
        return
    }

    // Re-read to sync
    got, d := readConfig(ctx, cfg, key)
    resp.Diagnostics.Append(d...)
    if resp.Diagnostics.HasError() {
        return
    }
    plan.ID = types.StringValue(key)
    plan.Value = types.StringValue(got)
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *ConfigResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // There is no per-key reset in `orb config`. We avoid calling global reset.
    // Leave the current setting as-is and remove from state.
    resp.State.RemoveResource(ctx)
}

// readConfig reads a config value using `orb config get KEY` and returns the trimmed value.
func readConfig(ctx context.Context, cfg *ClientConfig, key string) (string, diag.Diagnostics) {
    var diags diag.Diagnostics
    out, stderr, err := runOrb(ctx, cfg.OrbPath, "config", "get", key)
    if err == nil {
        if v, ok := parseConfigValueFromText(out, key); ok {
            return v, diags
        }
        // Some versions may still print just the value; handle that gracefully if no colon
        if !strings.Contains(out, ":") {
            return strings.TrimSpace(out), diags
        }
        // Fall through to try show
    }

    // Fallback to show + parse "key: value" line (also used if get output couldn't be parsed)
    show, _, err2 := runOrb(ctx, cfg.OrbPath, "config", "show")
    if err2 != nil {
        diags.AddError("failed to read config", stderr)
        return "", diags
    }
    if v, ok := parseConfigValueFromText(show, key); ok {
        return v, diags
    }
    diags.AddError("config key not found", "could not find key in orb config output")
    return "", diags
}

// parseConfigValueFromText scans lines like "key: value" and returns the value for the given key.
func parseConfigValueFromText(text, key string) (string, bool) {
    for _, line := range strings.Split(text, "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, key+":") {
            return strings.TrimSpace(strings.TrimPrefix(line, key+":")), true
        }
    }
    return "", false
}


