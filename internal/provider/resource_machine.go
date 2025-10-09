package provider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &MachineResource{}
var _ resource.ResourceWithImportState = &MachineResource{}
var _ resource.ResourceWithConfigure = &MachineResource{}

func NewMachineResource() resource.Resource { return &MachineResource{} }

type MachineResource struct {
	client *ClientConfig
}

type MachineModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Image         types.String `tfsdk:"image"`
	CloudInit     types.String `tfsdk:"cloud_init"`
	CloudInitFile types.String `tfsdk:"cloud_init_file"`
	ValidateImage types.Bool   `tfsdk:"validate_image"`

	// User configuration
	Username types.String `tfsdk:"username"`

	// Machine configuration
	PowerState types.String `tfsdk:"power_state"`
	Arch       types.String `tfsdk:"arch"`

	// Default machine setting
	DefaultMachine types.Bool `tfsdk:"default_machine"`

	IPAddress types.String `tfsdk:"ip_address"`
	Status    types.String `tfsdk:"status"`
	SSHHost   types.String `tfsdk:"ssh_host"`
	SSHPort   types.Int64  `tfsdk:"ssh_port"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func (r *MachineResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_machine"
}

func (r *MachineResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage an OrbStack Linux machine.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Internal identifier (defaults to name).",
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Machine name (unique).",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"image": schema.StringAttribute{
				Optional:    true,
				Description: "Base image/distribution (e.g., ubuntu, debian, alpine). Use OS:VERSION format for specific versions (e.g., ubuntu:noble, debian:bookworm). Default ubuntu.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_init": schema.StringAttribute{
				Optional:    true,
				Description: "cloud-init user data passed during creation (best-effort).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"cloud_init_file": schema.StringAttribute{
				Optional:    true,
				Description: "Path to a cloud-init user data file. Overrides cloud_init if both set.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"validate_image": schema.BoolAttribute{
				Optional:    true,
				Description: "Validate image exists before create; fail fast if unknown.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Description: "Username for the default user (defaults to macOS username).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"power_state": schema.StringAttribute{
				Optional:    true,
				Description: "Desired power state: running or stopped.",
				Validators:  []validator.String{stringvalidator.OneOf("running", "stopped")},
			},
			"arch": schema.StringAttribute{
				Optional:    true,
				Description: "Architecture passed to orb (-a): amd64 or arm64.",
				Validators:  []validator.String{stringvalidator.OneOf("amd64", "arm64")},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"default_machine": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Set this machine as the default machine for OrbStack. Only one machine can be the default.",
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
			},
			"ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "Machine IP address.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status reported by orb info.",
			},
			"ssh_host": schema.StringAttribute{
				Computed:    true,
				Description: "SSH host (usually same as ip_address).",
			},
			"ssh_port": schema.Int64Attribute{
				Computed:    true,
				Description: "SSH port.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "Creation time as reported by orb info.",
			},
		},
	}
}

func (r *MachineResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cfg, _ := req.ProviderData.(*ClientConfig)
	r.client = cfg
}

func (r *MachineResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MachineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.client
	if cfg == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client configuration")
		return
	}

	name := plan.Name.ValueString()
	image := plan.Image.ValueString()
	if image == "" {
		image = "ubuntu"
	}

	args := []string{"create"}

	// cloud-init: file takes precedence over inline
	if f := strings.TrimSpace(plan.CloudInitFile.ValueString()); f != "" {
		// ensure the file exists and pass absolute path
		if _, err := os.Stat(f); err != nil {
			resp.Diagnostics.AddError("cloud_init_file not found", err.Error())
			return
		}
		abs, err := filepath.Abs(f)
		if err != nil {
			resp.Diagnostics.AddError("failed to resolve cloud_init_file path", err.Error())
			return
		}
		args = append(args, "-c", abs)
	} else if v := plan.CloudInit.ValueString(); strings.TrimSpace(v) != "" {
		tmpFile, err := os.CreateTemp("", "orbstack-cloudinit-*.yaml")
		if err != nil {
			resp.Diagnostics.AddError("failed to create temp file for cloud-init", err.Error())
			return
		}
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString(v); err != nil {
			resp.Diagnostics.AddError("failed to write cloud-init to temp file", err.Error())
			return
		}
		if err := tmpFile.Close(); err != nil {
			resp.Diagnostics.AddError("failed to close cloud-init temp file", err.Error())
			return
		}
		args = append(args, "-c", tmpFile.Name())
	}

	// set_password removed (interactive-only flag not supported by Terraform)

	// Architecture flag
	if v := strings.TrimSpace(plan.Arch.ValueString()); v != "" {
		args = append(args, "-a", v)
	}

	// Username flag
	if v := strings.TrimSpace(plan.Username.ValueString()); v != "" {
		args = append(args, "-u", v)
	}

	// Use image directly (may include OS:VERSION format)
	imageArg := image

	// Validate image if requested
	if plan.ValidateImage.ValueBool() {
		known, d := listAvailableImages(ctx, cfg)
		resp.Diagnostics.Append(d...)
		if resp.Diagnostics.HasError() {
			return
		}
		if _, ok := known[strings.ToLower(imageArg)]; !ok {
			resp.Diagnostics.AddError("unknown image", fmt.Sprintf("image '%s' not found by orb", imageArg))
			return
		}
	}

	args = append(args, imageArg, name)

	_, stderr, err := runOrb(ctx, cfg.OrbPath, args...)
	if err != nil {
		resp.Diagnostics.AddError("failed to create machine", fmt.Sprintf("orb error: %s", stderr))
		return
	}

	model, diags := readUntilReady(ctx, cfg, name, cfg.CreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model == nil {
		resp.Diagnostics.AddError("machine not found after creation", name)
		return
	}

	plan.ID = types.StringValue(name)
	plan.IPAddress = model.IPAddress
	plan.Status = model.Status
	plan.SSHHost = model.SSHHost
	plan.SSHPort = model.SSHPort
	plan.CreatedAt = model.CreatedAt

	// Enforce desired power state after creation
	desired := strings.TrimSpace(plan.PowerState.ValueString())
	if desired == "stopped" {
		_, stderr, err := runOrb(ctx, cfg.OrbPath, "stop", name)
		if err != nil {
			resp.Diagnostics.AddError("failed to stop machine after create", fmt.Sprintf("orb error: %s", stderr))
			return
		}
		// refresh
		model, diags = readUntilReady(ctx, cfg, name, cfg.CreateTimeout)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
		plan.Status = model.Status
		plan.IPAddress = model.IPAddress
	}

	// Set as default machine if requested
	if plan.DefaultMachine.ValueBool() {
		_, stderr, err := runOrb(ctx, cfg.OrbPath, "default", name)
		if err != nil {
			resp.Diagnostics.AddError("failed to set default machine", fmt.Sprintf("orb error: %s", stderr))
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MachineResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MachineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.client
	if cfg == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client configuration")
		return
	}

	name := state.Name.ValueString()

	model, diags := readMachine(ctx, cfg, name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.IPAddress = model.IPAddress
	state.Status = model.Status
	state.SSHHost = model.SSHHost
	state.SSHPort = model.SSHPort
	state.CreatedAt = model.CreatedAt

	// Check if this machine is the current default
	isDefault, diags := r.isDefaultMachine(ctx, cfg, name)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.DefaultMachine = types.BoolValue(isDefault)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *MachineResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Read both state and plan to detect name changes and then apply rename
	var plan MachineModel
	var state MachineModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.client
	if cfg == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client configuration")
		return
	}

	oldName := state.Name.ValueString()
	newName := plan.Name.ValueString()

	// If the name changed, perform a rename using orb CLI
	if oldName != "" && newName != "" && oldName != newName {
		args := []string{"rename", oldName, newName}
		_, stderr, err := runOrb(ctx, cfg.OrbPath, args...)
		if err != nil {
			resp.Diagnostics.AddError("failed to rename machine", fmt.Sprintf("orb error: %s", stderr))
			return
		}
	}

	// Power state changes
	desired := strings.TrimSpace(plan.PowerState.ValueString())
	if desired == "running" {
		_, _, _ = runOrb(ctx, cfg.OrbPath, "start", newName)
	} else if desired == "stopped" {
		_, _, _ = runOrb(ctx, cfg.OrbPath, "stop", newName)
	}

	// Handle default machine changes only when explicitly set in config
	oldDefault := state.DefaultMachine.ValueBool()
	if !plan.DefaultMachine.IsNull() && !plan.DefaultMachine.IsUnknown() {
		newDefault := plan.DefaultMachine.ValueBool()
		if newDefault && !oldDefault {
			// Set this machine as default
			_, stderr, err := runOrb(ctx, cfg.OrbPath, "default", newName)
			if err != nil {
				resp.Diagnostics.AddError("failed to set default machine", fmt.Sprintf("orb error: %s", stderr))
				return
			}
		} else if !newDefault && oldDefault {
			// Unset default machine
			_, stderr, err := runOrb(ctx, cfg.OrbPath, "default", "none")
			if err != nil {
				resp.Diagnostics.AddError("failed to unset default machine", fmt.Sprintf("orb error: %s", stderr))
				return
			}
		}
	}

	// Re-read machine to populate all computed attributes and ensure known values
	model, diags := readMachine(ctx, cfg, newName)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if model == nil {
		resp.Diagnostics.AddError("machine not found after update", newName)
		return
	}

	plan.ID = types.StringValue(newName)
	plan.IPAddress = model.IPAddress
	plan.Status = model.Status
	plan.SSHHost = model.SSHHost
	plan.SSHPort = model.SSHPort
	plan.CreatedAt = model.CreatedAt

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *MachineResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MachineModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := r.client
	if cfg == nil {
		resp.Diagnostics.AddError("provider not configured", "missing client configuration")
		return
	}

	name := state.Name.ValueString()

	args := []string{"delete", name}
	_, stderr, err := runOrb(ctx, cfg.OrbPath, args...)
	if err != nil {
		resp.Diagnostics.AddError("failed to delete machine", fmt.Sprintf("orb error: %s", stderr))
		return
	}
}

func (r *MachineResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}

func readMachine(ctx context.Context, cfg *ClientConfig, name string) (*MachineModel, diag.Diagnostics) {
	var diags diag.Diagnostics

	out, _, err := runOrb(ctx, cfg.OrbPath, "info", name)
	if err != nil {
		diags.AddError("orb info failed", err.Error())
		return nil, diags
	}

	model := &MachineModel{
		Name: types.StringValue(name),
		ID:   types.StringValue(name),
	}

	// Support multiple labels across OrbStack versions
	status := firstNonEmpty(
		findLineValue(out, "Status:"),
		findLineValue(out, "State:"),
	)
	ip := firstNonEmpty(
		findLineValue(out, "IP:"),
		findLineValue(out, "IPv4:"),
	)
	ssh := findLineValue(out, "SSH:")
	created := firstNonEmpty(
		findLineValue(out, "Created:"),
		findLineValue(out, "Creation:"),
	)

	if status != "" {
		model.Status = types.StringValue(strings.TrimSpace(status))
	}
	if ip != "" {
		model.IPAddress = types.StringValue(strings.TrimSpace(ip))
		model.SSHHost = model.IPAddress
		if model.SSHPort.IsNull() || model.SSHPort.IsUnknown() {
			model.SSHPort = types.Int64Value(22)
		}
	}
	if created != "" {
		model.CreatedAt = types.StringValue(strings.TrimSpace(created))
	}
	if ssh != "" {
		if port := parseSSHPort(ssh); port > 0 {
			model.SSHPort = types.Int64Value(int64(port))
		}
	}

	return model, diags
}

// readUntilReady polls orb info until core fields are populated or timeout elapses.
func readUntilReady(ctx context.Context, cfg *ClientConfig, name string, timeoutStr string) (*MachineModel, diag.Diagnostics) {
	var diags diag.Diagnostics
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil || timeout <= 0 {
		timeout = 30 * time.Second
	}
	deadline := time.Now().Add(timeout)
	var last *MachineModel
	for {
		select {
		case <-ctx.Done():
			return last, diags
		default:
		}
		m, d := readMachine(ctx, cfg, name)
		diags.Append(d...)
		if m != nil {
			last = m
			if isMachineReady(m) {
				return m, diags
			}
		}
		if time.Now().After(deadline) {
			return last, diags
		}
		time.Sleep(2 * time.Second)
	}
}

func isMachineReady(m *MachineModel) bool {
	hasIP := !m.IPAddress.IsNull() && !m.IPAddress.IsUnknown() && strings.TrimSpace(m.IPAddress.ValueString()) != ""
	hasStatus := !m.Status.IsNull() && !m.Status.IsUnknown() && strings.TrimSpace(m.Status.ValueString()) != ""
	return hasIP && hasStatus
}

func findLineValue(text, prefix string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, prefix) {
			return strings.TrimSpace(strings.TrimPrefix(line, prefix))
		}
	}
	return ""
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func parseSSHPort(sshLine string) int {
	re := regexp.MustCompile(`\-p\s+(\d+)`)
	m := re.FindStringSubmatch(sshLine)
	if len(m) == 2 {
		var p int
		fmt.Sscanf(m[1], "%d", &p)
		return p
	}
	return 22
}

// isDefaultMachine checks if the given machine is the current default
func (r *MachineResource) isDefaultMachine(ctx context.Context, cfg *ClientConfig, machineName string) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	out, _, err := runOrb(ctx, cfg.OrbPath, "default")
	if err != nil {
		diags.AddError("orb default failed", err.Error())
		return false, diags
	}

	currentDefault := strings.TrimSpace(out)
	return currentDefault == machineName, diags
}

// listAvailableImages returns a set of lowercased tokens of the form name or name:tag
func listAvailableImages(ctx context.Context, cfg *ClientConfig) (map[string]struct{}, diag.Diagnostics) {
	var diags diag.Diagnostics
	out, _, err := runOrb(ctx, cfg.OrbPath, "images")
	if err != nil || strings.TrimSpace(out) == "" {
		out, _, _ = runOrb(ctx, cfg.OrbPath, "image", "list")
	}
	tokens := make(map[string]struct{})
	if strings.TrimSpace(out) == "" {
		return tokens, diags
	}
	// reuse simple parsing similar to datasource
	lineTokens := strings.FieldsFunc(strings.ToLower(out), func(r rune) bool {
		return r == '\n' || r == ' ' || r == '\t' || r == ','
	})
	for _, t := range lineTokens {
		t = strings.Trim(t, ":,.;()[]{}<>\"'`")
		if t == "" {
			continue
		}
		if strings.Contains(t, "--") {
			continue
		}
		if strings.HasPrefix(t, "usage:") || strings.HasPrefix(t, "aliases:") || strings.HasPrefix(t, "examples:") || strings.HasPrefix(t, "flags:") {
			continue
		}
		// accept name or name:tag pattern
		// quick check: starts with letter
		if t[0] < 'a' || t[0] > 'z' {
			continue
		}
		tokens[t] = struct{}{}
	}
	return tokens, diags
}
