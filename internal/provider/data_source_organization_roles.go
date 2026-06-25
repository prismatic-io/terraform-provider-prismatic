package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*organizationRolesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*organizationRolesDataSource)(nil)
)

type organizationRolesDataSource struct {
	client *graphql.Client
}

type organizationRoleModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Level       types.Int64  `tfsdk:"level"`
}

type organizationRolesModel struct {
	Id    types.String            `tfsdk:"id"`
	Roles []organizationRoleModel `tfsdk:"roles"`
}

func (d *organizationRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_roles"
}

func (d *organizationRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to list Prismatic Organization Roles.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"roles": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of organization roles available in Prismatic.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the role.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the role.",
						},
						"description": schema.StringAttribute{
							Computed:    true,
							Description: "The description of the role.",
						},
						"level": schema.Int64Attribute{
							Computed:    true,
							Description: "The permission level of the role. Higher values indicate more permissions.",
						},
					},
				},
			},
		},
	}
}

func (d *organizationRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *organizationRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// organizationRoles returns [Role]! directly (not a Connection type with nodes)
	var query struct {
		OrganizationRoles []struct {
			Id          graphql.ID
			Name        graphql.String
			Description graphql.String
			Level       graphql.Int
		} `graphql:"organizationRoles"`
	}

	if err := d.client.Query(ctx, &query, nil); err != nil {
		resp.Diagnostics.AddError("Unable to read organization roles", err.Error())
		return
	}

	state := organizationRolesModel{
		Id:    types.StringValue("organization_roles"),
		Roles: make([]organizationRoleModel, 0, len(query.OrganizationRoles)),
	}
	for _, roleNode := range query.OrganizationRoles {
		state.Roles = append(state.Roles, organizationRoleModel{
			Id:          types.StringValue(roleNode.Id.(string)),
			Name:        types.StringValue(string(roleNode.Name)),
			Description: types.StringValue(string(roleNode.Description)),
			Level:       types.Int64Value(int64(roleNode.Level)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
