package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

func dataSourceAuthenticatedUser() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to retrieve information about the currently authenticated user and the organization they belong to.",
		ReadContext: dataSourceAuthenticatedUserRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the authenticated user.",
			},
			"external_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The external identifier of the authenticated user.",
			},
			"email": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The email address of the authenticated user.",
			},
			"org_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the organization the authenticated user belongs to.",
			},
			"org_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the organization the authenticated user belongs to.",
			},
		},
	}
}

func dataSourceAuthenticatedUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

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

	if err := client.Query(context.Background(), &query, nil); err != nil {
		return diag.FromErr(err)
	}

	user := query.AuthenticatedUser

	if err := d.Set("external_id", string(user.ExternalId)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("email", string(user.Email)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("org_id", user.Org.Id.(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("org_name", string(user.Org.Name)); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(user.Id.(string))

	return diags
}
