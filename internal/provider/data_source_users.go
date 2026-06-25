package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*usersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*usersDataSource)(nil)
)

type usersDataSource struct {
	client *graphql.Client
}

type usersModel struct {
	Id             types.String `tfsdk:"id"`
	CustomerIsNull types.Bool   `tfsdk:"customer_is_null"`
	Users          []userModel  `tfsdk:"users"`
}

type userModel struct {
	Id         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Name       types.String `tfsdk:"name"`
	RoleId     types.String `tfsdk:"role_id"`
	RoleName   types.String `tfsdk:"role_name"`
	Phone      types.String `tfsdk:"phone"`
	ExternalId types.String `tfsdk:"external_id"`
	AvatarUrl  types.String `tfsdk:"avatar_url"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

func (d *usersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *usersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to list Prismatic Users.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"customer_is_null": schema.BoolAttribute{
				Optional: true,
				// Computed so the default true written back on the omitted path is legal.
				Computed:    true,
				Description: "Filter for users where customer is NULL. When true (default), returns only Organization users. When false, returns Customer users.",
			},
			"users": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of users in Prismatic.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The unique identifier of the user.",
						},
						"email": schema.StringAttribute{
							Computed:    true,
							Description: "The email address of the user.",
						},
						"name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the user.",
						},
						"role_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the role assigned to the user.",
						},
						"role_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the role assigned to the user.",
						},
						"phone": schema.StringAttribute{
							Computed:    true,
							Description: "The phone number of the user.",
						},
						"external_id": schema.StringAttribute{
							Computed:    true,
							Description: "The external ID for mapping to external systems.",
						},
						"avatar_url": schema.StringAttribute{
							Computed:    true,
							Description: "The URL of the user's avatar image.",
						},
						"created_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the user was created.",
						},
						"updated_at": schema.StringAttribute{
							Computed:    true,
							Description: "The timestamp when the user was last updated.",
						},
					},
				},
			},
		},
	}
}

func (d *usersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *usersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config usersModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	customerIsNull := true
	if !config.CustomerIsNull.IsNull() && !config.CustomerIsNull.IsUnknown() {
		customerIsNull = config.CustomerIsNull.ValueBool()
	}

	// users returns UserConnection! (with nodes)
	var query struct {
		Users struct {
			Nodes []struct {
				Id         graphql.ID
				Email      graphql.String
				Name       graphql.String
				Phone      graphql.String
				ExternalId graphql.String
				AvatarUrl  graphql.String
				CreatedAt  graphql.String
				UpdatedAt  graphql.String
				Role       struct {
					Id   graphql.ID
					Name graphql.String
				}
			}
		} `graphql:"users(customer_Isnull: $customerIsNull)"`
	}

	variables := map[string]interface{}{
		"customerIsNull": graphql.Boolean(customerIsNull),
	}

	if err := d.client.Query(ctx, &query, variables); err != nil {
		resp.Diagnostics.AddError("Unable to read users", err.Error())
		return
	}

	state := usersModel{
		Id:             types.StringValue("users"),
		CustomerIsNull: types.BoolValue(customerIsNull),
		Users:          make([]userModel, 0, len(query.Users.Nodes)),
	}
	for _, userNode := range query.Users.Nodes {
		state.Users = append(state.Users, userModel{
			Id:         types.StringValue(userNode.Id.(string)),
			Email:      types.StringValue(string(userNode.Email)),
			Name:       types.StringValue(string(userNode.Name)),
			RoleId:     types.StringValue(userNode.Role.Id.(string)),
			RoleName:   types.StringValue(string(userNode.Role.Name)),
			Phone:      types.StringValue(string(userNode.Phone)),
			ExternalId: types.StringValue(string(userNode.ExternalId)),
			AvatarUrl:  types.StringValue(string(userNode.AvatarUrl)),
			CreatedAt:  types.StringValue(string(userNode.CreatedAt)),
			UpdatedAt:  types.StringValue(string(userNode.UpdatedAt)),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
