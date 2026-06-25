package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

var (
	_ datasource.DataSource              = (*integrationsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*integrationsDataSource)(nil)
)

type integrationsDataSource struct {
	client *graphql.Client
}

type integrationsModel struct {
	Id           types.String       `tfsdk:"id"`
	Integrations []integrationModel `tfsdk:"integrations"`
}

type integrationModel struct {
	IntegrationId         types.String `tfsdk:"integration_id"`
	IntegrationName       types.String `tfsdk:"integration_name"`
	IntegrationDefinition types.String `tfsdk:"integration_definition"`
}

func (d *integrationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integrations"
}

func (d *integrationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Identifier for this data source.",
			},
			"integrations": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Data source to list Prismatic Integrations",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"integration_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the Integration",
						},
						"integration_name": schema.StringAttribute{
							Computed:    true,
							Description: "The name of the Integration",
						},
						"integration_definition": schema.StringAttribute{
							Computed:    true,
							Description: "The YAML definition of the Integration",
						},
					},
				},
			},
		},
	}
}

func (d *integrationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (d *integrationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var query struct {
		Integrations struct {
			Nodes []struct {
				Id         string
				Name       string
				Definition string
			}
		}
	}

	if err := d.client.Query(ctx, &query, nil); err != nil {
		resp.Diagnostics.AddError("Unable to read integrations", err.Error())
		return
	}

	state := integrationsModel{
		Id:           types.StringValue("integrations"),
		Integrations: make([]integrationModel, 0, len(query.Integrations.Nodes)),
	}
	for _, integrationNode := range query.Integrations.Nodes {
		state.Integrations = append(state.Integrations, integrationModel{
			IntegrationId:         types.StringValue(integrationNode.Id),
			IntegrationName:       types.StringValue(integrationNode.Name),
			IntegrationDefinition: types.StringValue(integrationNode.Definition),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
