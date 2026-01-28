package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

func resourceOrganizationUser() *schema.Resource {
	return &schema.Resource{
		Description:   "Manage Organization Users in Prismatic.",
		CreateContext: resourceOrganizationUserCreate,
		ReadContext:   resourceOrganizationUserRead,
		UpdateContext: resourceOrganizationUserUpdate,
		DeleteContext: resourceOrganizationUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The unique identifier of the user.",
			},
			"email": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The email address of the user. Changing this will recreate the user.",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The name of the user.",
			},
			"role": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the role to assign to the user.",
			},
			"phone": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "The phone number of the user. Note: The API normalizes phone numbers to E.164 format.",
			},
			"external_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: "An external ID for mapping to external systems.",
			},
			"avatar_url": {
				Type:        schema.TypeString,
				Optional:    true,
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
	}
}

func resourceOrganizationUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var mutation struct {
		CreateOrganizationUser struct {
			User struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"createOrganizationUser(input: $input)"`
	}

	type CreateOrganizationUserInput struct {
		Email      graphql.String `json:"email"`
		Name       graphql.String `json:"name,omitempty"`
		Role       graphql.ID     `json:"role"`
		Phone      graphql.String `json:"phone,omitempty"`
		ExternalId graphql.String `json:"externalId,omitempty"`
	}

	input := CreateOrganizationUserInput{
		Email: graphql.String(d.Get("email").(string)),
		Role:  graphql.ID(d.Get("role").(string)),
	}

	if v, ok := d.GetOk("name"); ok {
		input.Name = graphql.String(v.(string))
	}
	if v, ok := d.GetOk("phone"); ok {
		input.Phone = graphql.String(v.(string))
	}
	if v, ok := d.GetOk("external_id"); ok {
		input.ExternalId = graphql.String(v.(string))
	}

	variables := map[string]interface{}{
		"input": input,
	}

	if err := client.Mutate(context.Background(), &mutation, variables); err != nil {
		return diag.FromErr(err)
	}

	if len(mutation.CreateOrganizationUser.Errors) > 0 {
		return util.DiagFromGqlError(mutation.CreateOrganizationUser.Errors)
	}

	d.SetId(mutation.CreateOrganizationUser.User.Id.(string))

	// Read back the created resource to populate computed fields
	diags = append(diags, resourceOrganizationUserRead(ctx, d, m)...)

	return diags
}

func resourceOrganizationUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var query struct {
		User struct {
			Id         graphql.ID
			Email      graphql.String
			Name       graphql.String
			Phone      graphql.String
			ExternalId graphql.String
			AvatarUrl  graphql.String
			CreatedAt  graphql.String
			UpdatedAt  graphql.String
			Role       struct {
				Id graphql.ID
			}
		} `graphql:"user(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		// If the user is not found, remove from state
		if strings.Contains(err.Error(), "Record not found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("email", string(query.User.Email)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("name", string(query.User.Name)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("role", query.User.Role.Id.(string)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("phone", string(query.User.Phone)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("external_id", string(query.User.ExternalId)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("avatar_url", string(query.User.AvatarUrl)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("created_at", string(query.User.CreatedAt)); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("updated_at", string(query.User.UpdatedAt)); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceOrganizationUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	if d.HasChanges("name", "role", "phone", "external_id", "avatar_url") {
		var mutation struct {
			UpdateUser struct {
				User struct {
					Id graphql.ID
				}
				Errors util.GqlErrors
			} `graphql:"updateUser(input: $input)"`
		}

		type UpdateUserInput struct {
			Id         graphql.ID     `json:"id"`
			Name       graphql.String `json:"name,omitempty"`
			Role       graphql.ID     `json:"role,omitempty"`
			Phone      graphql.String `json:"phone,omitempty"`
			ExternalId graphql.String `json:"externalId,omitempty"`
			AvatarUrl  graphql.String `json:"avatarUrl,omitempty"`
		}

		input := UpdateUserInput{
			Id: graphql.ID(d.Id()),
		}

		if d.HasChange("name") {
			input.Name = graphql.String(d.Get("name").(string))
		}
		if d.HasChange("role") {
			input.Role = graphql.ID(d.Get("role").(string))
		}
		if d.HasChange("phone") {
			input.Phone = graphql.String(d.Get("phone").(string))
		}
		if d.HasChange("external_id") {
			input.ExternalId = graphql.String(d.Get("external_id").(string))
		}
		if d.HasChange("avatar_url") {
			input.AvatarUrl = graphql.String(d.Get("avatar_url").(string))
		}

		variables := map[string]interface{}{
			"input": input,
		}

		if err := client.Mutate(context.Background(), &mutation, variables); err != nil {
			return diag.FromErr(err)
		}

		if len(mutation.UpdateUser.Errors) > 0 {
			return util.DiagFromGqlError(mutation.UpdateUser.Errors)
		}
	}

	// Read back the updated resource
	diags = append(diags, resourceOrganizationUserRead(ctx, d, m)...)

	return diags
}

func resourceOrganizationUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var mutation struct {
		DeleteUser struct {
			User struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"deleteUser(input: $input)"`
	}

	type DeleteUserInput struct {
		Id graphql.ID `json:"id"`
	}

	variables := map[string]interface{}{
		"input": DeleteUserInput{
			Id: graphql.ID(d.Id()),
		},
	}

	if err := client.Mutate(context.Background(), &mutation, variables); err != nil {
		return diag.FromErr(err)
	}

	if len(mutation.DeleteUser.Errors) > 0 {
		return util.DiagFromGqlError(mutation.DeleteUser.Errors)
	}

	d.SetId("")

	return diags
}
