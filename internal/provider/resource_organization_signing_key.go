package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

var (
	_ resource.Resource                = (*organizationSigningKeyResource)(nil)
	_ resource.ResourceWithConfigure   = (*organizationSigningKeyResource)(nil)
	_ resource.ResourceWithImportState = (*organizationSigningKeyResource)(nil)
)

type organizationSigningKeyResource struct {
	client *graphql.Client
}

type organizationSigningKeyResourceModel struct {
	Id        types.String          `tfsdk:"id"`
	PublicKey normalizedStringValue `tfsdk:"public_key"`
	Imported  types.Bool            `tfsdk:"imported"`
	IssuedAt  types.String          `tfsdk:"issued_at"`
}

func (r *organizationSigningKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_signing_key"
}

func (r *organizationSigningKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Import a public key into the Organization's Signing Keys",
		Attributes: map[string]schema.Attribute{
			"public_key": schema.StringAttribute{
				CustomType:  normalizedStringType{},
				Required:    true,
				Description: "Public key to import",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"imported": schema.BoolAttribute{
				Computed:    true,
				Description: "Indicates if signing key was imported or generated",
			},
			"issued_at": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of when the signing key was issued",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the signing key",
			},
		},
	}
}

func (r *organizationSigningKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *organizationSigningKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationSigningKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		ImportOrganizationSigningKey struct {
			OrganizationSigningKey struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"importOrganizationSigningKey (input: {publicKey: $publicKey})"`
	}

	mutationVars := map[string]interface{}{
		"publicKey": graphql.String(plan.PublicKey.ValueString()),
	}

	if err := r.client.Mutate(ctx, &mutation, mutationVars); err != nil {
		resp.Diagnostics.AddError("Unable to import organization signing key", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.ImportOrganizationSigningKey.Errors)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := mutation.ImportOrganizationSigningKey.OrganizationSigningKey.Id.(string)

	// Read back from remote to populate state. The record is freshly created, so a
	// missing key here is unexpected and should not silently remove the resource.
	state := r.read(ctx, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state == nil {
		resp.Diagnostics.AddError("Unable to read organization signing key", "Signing key was imported but could not be found.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *organizationSigningKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationSigningKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	found := r.read(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if found == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, found)...)
}

// Update is intentionally a no-op. The only configurable attribute (public_key)
// forces replacement, so an in-place update is never requested.
func (r *organizationSigningKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationSigningKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// read scans the organization's signing keys for the given id and maps the match
// to a model, returning nil if none match. The API does not support filtering on
// this query, so all keys are fetched and scanned.
func (r *organizationSigningKeyResource) read(ctx context.Context, id string, diags *diag.Diagnostics) *organizationSigningKeyResourceModel {
	var query struct {
		Organization struct {
			SigningKeys struct {
				Nodes []struct {
					Id        string
					PublicKey string
					Imported  bool
					IssuedAt  string
				}
			}
		}
	}

	if err := r.client.Query(ctx, &query, nil); err != nil {
		diags.AddError("Unable to read organization signing key", err.Error())
		return nil
	}

	for _, signingKey := range query.Organization.SigningKeys.Nodes {
		if signingKey.Id == id {
			return &organizationSigningKeyResourceModel{
				Id:        types.StringValue(signingKey.Id),
				PublicKey: normalizedStringValue{StringValue: basetypes.NewStringValue(strings.TrimSpace(signingKey.PublicKey))},
				Imported:  types.BoolValue(signingKey.Imported),
				IssuedAt:  types.StringValue(signingKey.IssuedAt),
			}
		}
	}

	return nil
}

func (r *organizationSigningKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationSigningKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		DeleteOrganizationSigningKey struct {
			OrganizationSigningKey struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"deleteOrganizationSigningKey (input: {id: $id})"`
	}

	mutationVars := map[string]interface{}{
		"id": graphql.ID(state.Id.ValueString()),
	}

	if err := r.client.Mutate(ctx, &mutation, mutationVars); err != nil {
		resp.Diagnostics.AddError("Unable to delete organization signing key", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.DeleteOrganizationSigningKey.Errors)...)
}

func (r *organizationSigningKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
