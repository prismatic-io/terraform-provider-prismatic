package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

func dataSourceUsers() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to list Prismatic Users.",
		ReadContext: dataSourceUsersRead,
		Schema: map[string]*schema.Schema{
			"customer_is_null": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Filter for users where customer is NULL. When true (default), returns only Organization users. When false, returns Customer users.",
			},
			"users": {
				Description: "List of users in Prismatic.",
				Type:        schema.TypeList,
				Computed:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The unique identifier of the user.",
						},
						"email": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The email address of the user.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the user.",
						},
						"role_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The ID of the role assigned to the user.",
						},
						"role_name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name of the role assigned to the user.",
						},
						"phone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The phone number of the user.",
						},
						"external_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The external ID for mapping to external systems.",
						},
						"avatar_url": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The URL of the user's avatar image.",
						},
						"created_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The timestamp when the user was created.",
						},
						"updated_at": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The timestamp when the user was last updated.",
						},
					},
				},
			},
		},
	}
}

func dataSourceUsersRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	customerIsNull := d.Get("customer_is_null").(bool)

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

	if err := client.Query(context.Background(), &query, variables); err != nil {
		return diag.FromErr(err)
	}

	count := len(query.Users.Nodes)
	users := make([]interface{}, count)
	for i, userNode := range query.Users.Nodes {
		user := make(map[string]interface{})
		user["id"] = userNode.Id.(string)
		user["email"] = string(userNode.Email)
		user["name"] = string(userNode.Name)
		user["phone"] = string(userNode.Phone)
		user["external_id"] = string(userNode.ExternalId)
		user["avatar_url"] = string(userNode.AvatarUrl)
		user["created_at"] = string(userNode.CreatedAt)
		user["updated_at"] = string(userNode.UpdatedAt)
		user["role_id"] = userNode.Role.Id.(string)
		user["role_name"] = string(userNode.Role.Name)
		users[i] = user
	}

	if err := d.Set("users", users); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.UniqueId())

	return diags
}
