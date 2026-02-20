package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

func dataSourceOrganizationRoles() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to list Prismatic Organization Roles.",
		ReadContext: dataSourceOrganizationRolesRead,
		Schema: map[string]*schema.Schema{
			"roles": {
				Description: "List of organization roles available in Prismatic.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the role.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the role.",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The description of the role.",
						},
						"level": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "The permission level of the role. Higher values indicate more permissions.",
						},
					},
				},
			},
		},
	}
}

func dataSourceOrganizationRolesRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	// organizationRoles returns [Role]! directly (not a Connection type with nodes)
	var query struct {
		OrganizationRoles []struct {
			Id          graphql.ID
			Name        graphql.String
			Description graphql.String
			Level       graphql.Int
		} `graphql:"organizationRoles"`
	}

	if err := client.Query(context.Background(), &query, nil); err != nil {
		return diag.FromErr(err)
	}

	count := len(query.OrganizationRoles)
	roles := make([]interface{}, count)
	for i, roleNode := range query.OrganizationRoles {
		role := make(map[string]interface{})
		role["id"] = roleNode.Id.(string)
		role["name"] = string(roleNode.Name)
		role["description"] = string(roleNode.Description)
		role["level"] = int(roleNode.Level)
		roles[i] = role
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.UniqueId())

	return diags
}
