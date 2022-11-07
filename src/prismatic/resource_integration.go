package prismatic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
	"gopkg.in/yaml.v3"
	"reflect"
	"strings"
)

func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIntegrationCreate,
		ReadContext:   resourceIntegrationRead,
		UpdateContext: resourceIntegrationUpdate,
		DeleteContext: resourceIntegrationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"definition": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressDiffIntegrationDefinition,
			},
			"name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceIntegrationCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var mutation struct {
		ImportIntegration struct {
			Integration struct {
				Id graphql.ID
			}
		} `graphql:"importIntegration (input: $input)"`
	}
	type ImportIntegrationInput struct {
		Id         graphql.ID     `json:"integrationId"`
		Definition graphql.String `json:"definition"`
	}
	importVariables := map[string]interface{}{
		"input": ImportIntegrationInput{
			Id:         d.Id(),
			Definition: graphql.String(d.Get("definition").(string)),
		},
	}

	if err := client.Mutate(context.Background(), &mutation, importVariables); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(mutation.ImportIntegration.Integration.Id.(string))

	// Update state from remote
	diags = append(diags, resourceIntegrationRead(ctx, d, m)...)

	return diags
}

func resourceIntegrationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var query struct {
		Integration struct {
			Id          graphql.ID
			Name        graphql.String
			Description graphql.String
			Definition  graphql.String
		} `graphql:"integration(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}
	if err := client.Query(context.Background(), &query, variables); err != nil {
		if strings.Contains(err.Error(), "Record not found") {
			d.SetId("")
			return diags
		}
		return diag.FromErr(err)
	}

	if err := d.Set("name", query.Integration.Name); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", query.Integration.Description); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("definition", query.Integration.Definition); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	if d.HasChange("definition") {
		var importMutation struct {
			ImportIntegration struct {
				Integration struct {
					Id graphql.ID
				}
			} `graphql:"importIntegration (input: $input)"`
		}
		type ImportIntegrationInput struct {
			Id         graphql.ID     `json:"integrationId"`
			Definition graphql.String `json:"definition"'`
		}
		importVariables := map[string]interface{}{
			"input": ImportIntegrationInput{
				Id:         d.Id(),
				Definition: graphql.String(d.Get("definition").(string)),
			},
		}

		if err := client.Mutate(context.Background(), &importMutation, importVariables); err != nil {
			diags = append(diags, diag.FromErr(err)...)
		}
	}

	diags = append(diags, resourceIntegrationRead(ctx, d, m)...)

	return diags
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var mutation struct {
		DeleteIntegration struct {
			Integration struct {
				Id graphql.ID
			}
		} `graphql:"deleteIntegration (input: $input)"`
	}
	type DeleteIntegrationInput struct {
		Id graphql.ID `json:"id"`
	}
	variables := map[string]interface{}{
		"input": DeleteIntegrationInput{
			Id: d.Id(),
		},
	}

	if err := client.Mutate(context.Background(), &mutation, variables); err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}

// Suppress diff output for integration definitions if they're logically the same. We don't
// care if ordering is different only that they're representing the same definition.
func suppressDiffIntegrationDefinition(k, old, new string, d *schema.ResourceData) bool {
	var oldData map[string]interface{}
	var newData map[string]interface{}

	_ = yaml.Unmarshal([]byte(old), &oldData)
	_ = yaml.Unmarshal([]byte(new), &newData)

	return reflect.DeepEqual(oldData, newData)
}
