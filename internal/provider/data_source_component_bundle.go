package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*componentBundleDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*componentBundleDataSource)(nil)
)

type componentBundleDataSource struct {
	client *graphql.Client
}

type componentBundleModel struct {
	Id              types.String `tfsdk:"id"`
	BundleDirectory types.String `tfsdk:"bundle_directory"`
	BundlePath      types.String `tfsdk:"bundle_path"`
	Signature       types.String `tfsdk:"signature"`
}

func (d *componentBundleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component_bundle"
}

func (d *componentBundleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Generates a component bundle suitable for publishing",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"bundle_directory": schema.StringAttribute{
				Required:    true,
				Description: "Directory to bundle",
			},
			"bundle_path": schema.StringAttribute{
				Required:    true,
				Description: "Destination of the generated bundle",
			},
			"signature": schema.StringAttribute{
				Computed:    true,
				Description: "Signature of the bundle for detecting redundant publishes",
			},
		},
	}
}

func (d *componentBundleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *componentBundleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config componentBundleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bundleDirectory := config.BundleDirectory.ValueString()
	bundlePath := config.BundlePath.ValueString()

	_, packageSignature, err := util.GenerateBundleSignature(bundleDirectory, bundlePath)
	if err != nil {
		resp.Diagnostics.AddError("Unable to generate bundle signature", err.Error())
		return
	}

	state := componentBundleModel{
		Id:              types.StringValue(bundlePath),
		BundleDirectory: config.BundleDirectory,
		BundlePath:      config.BundlePath,
		Signature:       types.StringValue(packageSignature),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
