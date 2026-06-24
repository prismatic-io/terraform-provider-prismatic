package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
	"gopkg.in/yaml.v3"
	"strings"
)

func resourceIntegration() *schema.Resource {
	return &schema.Resource{
		Description:   "Import Integrations into Prismatic using YAML definitions.",
		CreateContext: resourceIntegrationCreate,
		ReadContext:   resourceIntegrationRead,
		UpdateContext: resourceIntegrationUpdate,
		DeleteContext: resourceIntegrationDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Integration",
			},
			"definition": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressDiffIntegrationDefinition,
				Description:      "The YAML definition of the Integration",
			},
			"name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The name of the Integration",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Integration",
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
			Errors util.GqlErrors
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

	if len(mutation.ImportIntegration.Errors) > 0 {
		return util.DiagFromGqlError(mutation.ImportIntegration.Errors)
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
		var mutation struct {
			ImportIntegration struct {
				Integration struct {
					Id graphql.ID
				}
				Errors util.GqlErrors
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
			diags = append(diags, diag.FromErr(err)...)
		}

		if len(mutation.ImportIntegration.Errors) > 0 {
			return util.DiagFromGqlError(mutation.ImportIntegration.Errors)
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
			Errors util.GqlErrors
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

	if len(mutation.DeleteIntegration.Errors) > 0 {
		return util.DiagFromGqlError(mutation.DeleteIntegration.Errors)
	}

	d.SetId("")

	return diags
}

// suppressDiffIntegrationDefinition suppresses spurious diffs between the YAML
// definition in config and the heavily-normalized form the API returns on read
// (it resolves component versions, injects defaults, and drops empties). The
// config definition is treated as equivalent when it is a semantic subset of the
// canonical definition. See definitionsEquivalent.
func suppressDiffIntegrationDefinition(k, old, new string, d *schema.ResourceData) bool {
	// old = normalized definition stored in state; new = definition from config.
	return definitionsEquivalent(new, old)
}

// definitionsEquivalent reports whether the submitted definition is a semantic
// subset of the server's canonical (normalized) definition.
func definitionsEquivalent(submitted, canonical string) bool {
	var w, g interface{}
	if yaml.Unmarshal([]byte(submitted), &w) != nil || yaml.Unmarshal([]byte(canonical), &g) != nil {
		return false
	}
	return yamlSubset(w, g)
}

// yamlSubset reports whether want is contained in got: every map key in want must
// exist in got with a subset value (keys present only in got — server-injected
// defaults — are ignored); lists compare element-wise; scalars must be equal.
// Two carve-outs model the API's normalization: a "version" of "LATEST" matches
// any resolved version, and an empty want value matches an absent got key.
func yamlSubset(want, got interface{}) bool {
	switch w := want.(type) {
	case map[string]interface{}:
		g, ok := got.(map[string]interface{})
		if !ok {
			return false
		}
		for k, wv := range w {
			gv, present := g[k]
			if !present {
				if yamlEmpty(wv) {
					continue
				}
				return false
			}
			if k == "version" && fmt.Sprint(wv) == "LATEST" {
				continue
			}
			if !yamlSubset(wv, gv) {
				return false
			}
		}
		return true
	case []interface{}:
		g, ok := got.([]interface{})
		if !ok || len(w) != len(g) {
			return false
		}
		for i := range w {
			if !yamlSubset(w[i], g[i]) {
				return false
			}
		}
		return true
	default:
		return fmt.Sprint(want) == fmt.Sprint(got)
	}
}

// yamlEmpty reports whether v is nil, an empty string, an empty map, or an empty slice.
func yamlEmpty(v interface{}) bool {
	switch x := v.(type) {
	case nil:
		return true
	case string:
		return x == ""
	case map[string]interface{}:
		return len(x) == 0
	case []interface{}:
		return len(x) == 0
	default:
		return false
	}
}
