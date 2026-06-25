package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*organizationSigningKeyDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*organizationSigningKeyDataSource)(nil)
)

type organizationSigningKeyDataSource struct {
	client *graphql.Client
}

type organizationSigningKeyModel struct {
	Id        types.String `tfsdk:"id"`
	Imported  types.Bool   `tfsdk:"imported"`
	IssuedAt  types.String `tfsdk:"issued_at"`
	PublicKey types.String `tfsdk:"public_key"`
}

func (d *organizationSigningKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_signing_key"
}

func (d *organizationSigningKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to retrieve an Organization's signing key",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the signing key to fetch.",
			},
			"imported": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates if the signing key was imported or generated",
			},
			"issued_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp the signing key was issued at",
			},
			"public_key": schema.StringAttribute{
				Computed:    true,
				Description: "The public key of the signing key",
			},
		},
	}
}

func (d *organizationSigningKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *organizationSigningKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationSigningKeyModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var query struct {
		Organization struct {
			SigningKeys struct {
				Nodes []struct {
					Id        string
					Imported  bool
					IssuedAt  string
					PublicKey string
				}
			}
		}
	}

	targetID := config.Id.ValueString()

	if err := d.client.Query(ctx, &query, nil); err != nil {
		resp.Diagnostics.AddError("Unable to read organization signing key", err.Error())
		return
	}

	var found bool
	var state organizationSigningKeyModel
	for _, signingKey := range query.Organization.SigningKeys.Nodes {
		if signingKey.Id == targetID {
			state = organizationSigningKeyModel{
				Id:        types.StringValue(signingKey.Id),
				Imported:  types.BoolValue(signingKey.Imported),
				IssuedAt:  types.StringValue(signingKey.IssuedAt),
				PublicKey: types.StringValue(signingKey.PublicKey),
			}
			found = true
			break
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Organization signing key not found",
			"No organization signing key found with ID: "+targetID,
		)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
