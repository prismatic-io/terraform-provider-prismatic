package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*componentsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*componentsDataSource)(nil)
)

type componentsDataSource struct {
	client *graphql.Client
}

type componentsModel struct {
	Id         types.String     `tfsdk:"id"`
	Components []componentModel `tfsdk:"components"`
}

type componentModel struct {
	ComponentId          types.String `tfsdk:"component_id"`
	ComponentKey         types.String `tfsdk:"component_key"`
	ComponentLabel       types.String `tfsdk:"component_label"`
	ComponentDescription types.String `tfsdk:"component_description"`
}

func (d *componentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_components"
}

func (d *componentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Data source to list Prismatic components",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"components": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"component_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Component",
						},
						"component_key": schema.StringAttribute{
							Computed:    true,
							Description: "The key of the Component",
						},
						"component_label": schema.StringAttribute{
							Computed:    true,
							Description: "The label of the Component",
						},
						"component_description": schema.StringAttribute{
							Computed:    true,
							Description: "The description of the Component",
						},
					},
				},
			},
		},
	}
}

func (d *componentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *componentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var query struct {
		Components struct {
			Nodes []struct {
				Id          string
				Key         string
				Label       string
				Description string
			}
		}
	}

	if err := d.client.Query(ctx, &query, nil); err != nil {
		resp.Diagnostics.AddError("Unable to read components", err.Error())
		return
	}

	state := componentsModel{
		Id:         types.StringValue("components"),
		Components: make([]componentModel, 0, len(query.Components.Nodes)),
	}

	for _, componentNode := range query.Components.Nodes {
		state.Components = append(state.Components, componentModel{
			ComponentId:          types.StringValue(componentNode.Id),
			ComponentKey:         types.StringValue(componentNode.Key),
			ComponentLabel:       types.StringValue(componentNode.Label),
			ComponentDescription: types.StringValue(componentNode.Description),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
