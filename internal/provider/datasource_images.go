package provider

import (
    "context"
    "regexp"
    "sort"
    "strings"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ImagesDataSource{}
var _ datasource.DataSourceWithConfigure = &ImagesDataSource{}

func NewImagesDataSource() datasource.DataSource { return &ImagesDataSource{} }

type ImagesDataSource struct {
    client *ClientConfig
}

type ImagesDataSourceModel struct {
    Filter types.String   `tfsdk:"filter"`
    Images []ImageItemDTO `tfsdk:"images"`
}

type ImageItemDTO struct {
    Name        types.String `tfsdk:"name"`
    Tag         types.String `tfsdk:"tag"`
    DisplayName types.String `tfsdk:"display_name"`
    Default     types.Bool   `tfsdk:"default"`
}

func (d *ImagesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_images"
}

func (d *ImagesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Description: "List available OrbStack images. Attempts `orb images`; may fall back to other commands.",
        Attributes: map[string]schema.Attribute{
            "filter": schema.StringAttribute{
                Optional:    true,
                Description: "Optional substring filter to match name or tag.",
            },
            "images": schema.ListNestedAttribute{
                Computed:    true,
                Description: "Discovered images.",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            Computed:    true,
                            Description: "Image name (e.g., ubuntu, debian, alpine).",
                        },
                        "tag": schema.StringAttribute{
                            Computed:    true,
                            Description: "Optional tag/version (e.g., 24.04, bookworm).",
                        },
                        "display_name": schema.StringAttribute{
                            Computed:    true,
                            Description: "Display name if provided by OrbStack.",
                        },
                        "default": schema.BoolAttribute{
                            Computed:    true,
                            Description: "Whether this image is the default (if detectable).",
                        },
                    },
                },
            },
        },
    }
}

func (d *ImagesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    cfg, _ := req.ProviderData.(*ClientConfig)
    d.client = cfg
}

func (d *ImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data ImagesDataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...) 
    if resp.Diagnostics.HasError() {
        return
    }

    cfg := d.client
    if cfg == nil {
        resp.Diagnostics.AddError("provider not configured", "missing client configuration")
        return
    }

    // Try a sequence of commands to discover images.
    // 1) orb images
    // 2) orb image list
    // 3) orb create --help (last resort heuristics)
    stdout, _, err := runOrb(ctx, cfg.OrbPath, "images")
    if err != nil || strings.TrimSpace(stdout) == "" {
        stdout, _, err = runOrb(ctx, cfg.OrbPath, "image", "list")
    }
    if (err != nil || strings.TrimSpace(stdout) == "") {
        // Last resort: attempt to parse help output for clues
        // This is best-effort; if empty, we'll return empty list gracefully.
        help, _, _ := runOrb(ctx, cfg.OrbPath, "create", "--help")
        if strings.TrimSpace(help) != "" {
            stdout = help
        }
    }

    // Extract plausible tokens from output (lines or words) that match image[:tag]
    tokenRe := regexp.MustCompile(`^[a-z][a-z0-9_-]*(?::[A-Za-z0-9._-]+)?$`)
    ignoreSet := map[string]struct{}{
        "orb": {}, "create": {}, "add": {}, "new": {}, "help": {},
        "usage": {}, "aliases": {}, "examples": {}, "flags": {},
        "for": {}, "the": {}, "machine_name": {}, "distro[:version]": {},
    }
    uniq := make(map[string]struct{})
    var tokens []string
    lines := strings.Split(stdout, "\n")
    for _, line := range lines {
        l := strings.ToLower(strings.TrimSpace(line))
        if l == "" { continue }
        if strings.HasPrefix(l, "usage:") || strings.HasPrefix(l, "aliases:") || strings.HasPrefix(l, "examples:") || strings.HasPrefix(l, "flags:") {
            continue
        }
        // split into words to avoid capturing full example lines
        for _, w := range strings.Fields(l) {
            w = strings.TrimSpace(strings.Trim(w, ",.;()[]{}<>\"'`"))
            if w == "" { continue }
            if _, skip := ignoreSet[w]; skip { continue }
            if tokenRe.MatchString(w) {
                if _, seen := uniq[w]; !seen {
                    uniq[w] = struct{}{}
                    tokens = append(tokens, w)
                }
            }
        }
    }
    sort.Strings(tokens)

    var items []ImageItemDTO
    for _, tok := range tokens {
        name, tag := parseImageToken(tok)
        if name == "" { continue }
        items = append(items, ImageItemDTO{
            Name:        types.StringValue(name),
            Tag:         types.StringValue(tag),
            DisplayName: types.StringNull(),
            Default:     types.BoolValue(false),
        })
    }

    // Apply filter if provided
    if !data.Filter.IsNull() && !data.Filter.IsUnknown() {
        f := strings.ToLower(strings.TrimSpace(data.Filter.ValueString()))
        if f != "" {
            filtered := make([]ImageItemDTO, 0, len(items))
            for _, it := range items {
                n := strings.ToLower(it.Name.ValueString())
                t := strings.ToLower(it.Tag.ValueString())
                if strings.Contains(n, f) || strings.Contains(t, f) {
                    filtered = append(filtered, it)
                }
            }
            items = filtered
        }
    }

    data.Images = items
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func parseImageToken(token string) (string, string) {
    // Accept formats like:
    //   ubuntu
    //   ubuntu:24.04
    //   debian:bookworm
    // Strip any leading bullet chars
    token = strings.TrimSpace(strings.TrimLeft(token, "-*•"))
    parts := strings.SplitN(token, ":", 2)
    if len(parts) == 2 {
        return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
    }
    return strings.TrimSpace(token), ""
}


