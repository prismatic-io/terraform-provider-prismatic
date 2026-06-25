package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
	"gopkg.in/yaml.v3"
)

var (
	_ resource.Resource                = (*integrationResource)(nil)
	_ resource.ResourceWithConfigure   = (*integrationResource)(nil)
	_ resource.ResourceWithImportState = (*integrationResource)(nil)
)

type integrationResource struct {
	client *graphql.Client
}

type integrationResourceModel struct {
	Id          types.String          `tfsdk:"id"`
	Definition  definitionStringValue `tfsdk:"definition"`
	Name        types.String          `tfsdk:"name"`
	Description types.String          `tfsdk:"description"`
}

type ImportIntegrationInput struct {
	Id         graphql.ID     `json:"integrationId"`
	Definition graphql.String `json:"definition"`
}

type DeleteIntegrationInput struct {
	Id graphql.ID `json:"id"`
}

func (r *integrationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_integration"
}

func (r *integrationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Import Integrations into Prismatic using YAML definitions.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Integration",
			},
			"definition": schema.StringAttribute{
				CustomType:  definitionStringType{},
				Required:    true,
				Description: "The YAML definition of the Integration",
			},
			"name": schema.StringAttribute{
				Computed:    true,
				Description: "The name of the Integration",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the Integration",
			},
		},
	}
}

func (r *integrationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *integrationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan integrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := r.importIntegration(ctx, "", plan.Definition.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.read(ctx, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state == nil {
		resp.Diagnostics.AddError("Unable to read integration after create", "The integration could not be found after import.")
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *integrationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state integrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updated := r.read(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if updated == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *integrationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan integrationResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var priorState integrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &priorState)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := priorState.Id.ValueString()

	r.importIntegration(ctx, id, plan.Definition.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	state := r.read(ctx, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *integrationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state integrationResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		DeleteIntegration struct {
			Integration struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"deleteIntegration (input: $input)"`
	}
	variables := map[string]interface{}{
		"input": DeleteIntegrationInput{
			Id: graphql.ID(state.Id.ValueString()),
		},
	}

	if err := r.client.Mutate(ctx, &mutation, variables); err != nil {
		resp.Diagnostics.AddError("Unable to delete integration", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.DeleteIntegration.Errors)...)
}

func (r *integrationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// importIntegration runs the importIntegration mutation and returns the resulting
// integration id. It returns "" and records diagnostics on failure.
func (r *integrationResource) importIntegration(ctx context.Context, id, definition string, diags *diag.Diagnostics) string {
	var mutation struct {
		ImportIntegration struct {
			Integration struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"importIntegration (input: $input)"`
	}
	variables := map[string]interface{}{
		"input": ImportIntegrationInput{
			Id:         graphql.ID(id),
			Definition: graphql.String(definition),
		},
	}

	if err := r.client.Mutate(ctx, &mutation, variables); err != nil {
		diags.AddError("Unable to import integration", err.Error())
		return ""
	}

	diags.Append(gqlErrorDiagnostics(mutation.ImportIntegration.Errors)...)
	if diags.HasError() {
		return ""
	}

	return mutation.ImportIntegration.Integration.Id.(string)
}

// read queries the integration by id and maps it to a model, returning nil if the
// integration no longer exists.
func (r *integrationResource) read(ctx context.Context, id string, diags *diag.Diagnostics) *integrationResourceModel {
	var query struct {
		Integration struct {
			Id          graphql.ID
			Name        graphql.String
			Description graphql.String
			Definition  graphql.String
		} `graphql:"integration(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.ID(id),
	}
	if err := r.client.Query(ctx, &query, variables); err != nil {
		if isRecordNotFound(err) {
			return nil
		}
		diags.AddError("Unable to read integration", err.Error())
		return nil
	}

	return &integrationResourceModel{
		Id:          types.StringValue(query.Integration.Id.(string)),
		Name:        types.StringValue(string(query.Integration.Name)),
		Description: types.StringValue(string(query.Integration.Description)),
		Definition:  definitionStringValue{StringValue: basetypes.NewStringValue(string(query.Integration.Definition))},
	}
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
