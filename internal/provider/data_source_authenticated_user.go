package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*authenticatedUserDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*authenticatedUserDataSource)(nil)
)

type authenticatedUserDataSource struct {
	client *graphql.Client
}

type authenticatedUserModel struct {
	Id         types.String `tfsdk:"id"`
	ExternalId types.String `tfsdk:"external_id"`
	Email      types.String `tfsdk:"email"`
	OrgId      types.String `tfsdk:"org_id"`
	OrgName    types.String `tfsdk:"org_name"`
}

func (d *authenticatedUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_authenticated_user"
}

func (d *authenticatedUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to retrieve information about the currently authenticated user and the organization they belong to.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the authenticated user.",
			},
			"external_id": schema.StringAttribute{
				Computed:    true,
				Description: "The external identifier of the authenticated user.",
			},
			"email": schema.StringAttribute{
				Computed:    true,
				Description: "The email address of the authenticated user.",
			},
			"org_id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the organization the authenticated user belongs to.",
			},
			"org_name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the organization the authenticated user belongs to.",
			},
		},
	}
}

func (d *authenticatedUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *authenticatedUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var query struct {
		AuthenticatedUser struct {
			Id         graphql.ID
			ExternalId graphql.String
			Email      graphql.String
			Org        struct {
				Id   graphql.ID
				Name graphql.String
			}
		} `graphql:"authenticatedUser"`
	}

	if err := d.client.Query(ctx, &query, nil); err != nil {
		resp.Diagnostics.AddError("Unable to read authenticated user", err.Error())
		return
	}

	user := query.AuthenticatedUser
	state := authenticatedUserModel{
		Id:         types.StringValue(user.Id.(string)),
		ExternalId: types.StringValue(string(user.ExternalId)),
		Email:      types.StringValue(string(user.Email)),
		OrgId:      types.StringValue(user.Org.Id.(string)),
		OrgName:    types.StringValue(string(user.Org.Name)),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
