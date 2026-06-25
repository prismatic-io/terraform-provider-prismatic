package provider

import (
	"context"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

var (
	_ resource.Resource                = (*organizationUserResource)(nil)
	_ resource.ResourceWithConfigure   = (*organizationUserResource)(nil)
	_ resource.ResourceWithImportState = (*organizationUserResource)(nil)
)

type organizationUserResource struct {
	client *graphql.Client
}

type organizationUserResourceModel struct {
	Id         types.String `tfsdk:"id"`
	Email      types.String `tfsdk:"email"`
	Name       types.String `tfsdk:"name"`
	Role       types.String `tfsdk:"role"`
	Phone      types.String `tfsdk:"phone"`
	ExternalId types.String `tfsdk:"external_id"`
	AvatarUrl  types.String `tfsdk:"avatar_url"`
	CreatedAt  types.String `tfsdk:"created_at"`
	UpdatedAt  types.String `tfsdk:"updated_at"`
}

// e164PhoneValidator validates that a phone number is in E.164 format.
type e164PhoneValidator struct{}

func (v e164PhoneValidator) Description(ctx context.Context) string {
	return "Phone number must be in E.164 format (e.g., +14155552671). It must start with '+' followed by 7-15 digits, and the first digit after '+' cannot be 0."
}

func (v e164PhoneValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v e164PhoneValidator) ValidateString(ctx context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}

	value := req.ConfigValue.ValueString()

	// Allow empty string since the field is optional
	if value == "" {
		return
	}

	// E.164 format: + followed by 7-15 digits, first digit cannot be 0
	e164Regex := regexp.MustCompile(`^\+[1-9]\d{6,14}$`)
	if !e164Regex.MatchString(value) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid phone number format",
			"Phone number must be in E.164 format (e.g., +14155552671). It must start with '+' followed by 7-15 digits, and the first digit after '+' cannot be 0.",
		)
	}
}

func (r *organizationUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user"
}

func (r *organizationUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manage Organization Users in Prismatic.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier of the user.",
			},
			"email": schema.StringAttribute{
				Required:    true,
				Description: "The email address of the user. Changing this will recreate the user.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the user.",
				// Hold prior value when omitted so an unrelated update doesn't wipe it.
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"role": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the role to assign to the user.",
			},
			"phone": schema.StringAttribute{
				Optional: true,
				// Computed only to carry the Default (the framework requires it). The ""
				// default makes an omitted phone plan as a known "" rather than unknown,
				// so an unrelated update cannot wipe a previously set value.
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "The phone number of the user in E.164 format (e.g., +14155552671). Must start with '+' followed by 7-15 digits.",
				Validators: []validator.String{
					e164PhoneValidator{},
				},
			},
			"external_id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "An external ID for mapping to external systems.",
			},
			"avatar_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The URL of the user's avatar image.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the user was created.",
			},
			"updated_at": schema.StringAttribute{
				Computed:    true,
				Description: "The timestamp when the user was last updated.",
			},
		},
	}
}

func (r *organizationUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *organizationUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		CreateOrganizationUser struct {
			User struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"createOrganizationUser(input: $input)"`
	}

	input := CreateOrganizationUserInput{
		Email: graphql.String(plan.Email.ValueString()),
		Role:  graphql.ID(plan.Role.ValueString()),
	}

	if !plan.Name.IsNull() && !plan.Name.IsUnknown() {
		input.Name = graphql.String(plan.Name.ValueString())
	}
	if !plan.Phone.IsNull() && !plan.Phone.IsUnknown() {
		input.Phone = graphql.String(plan.Phone.ValueString())
	}
	if !plan.ExternalId.IsNull() && !plan.ExternalId.IsUnknown() {
		input.ExternalId = graphql.String(plan.ExternalId.ValueString())
	}

	variables := map[string]interface{}{
		"input": input,
	}

	if err := r.client.Mutate(ctx, &mutation, variables); err != nil {
		resp.Diagnostics.AddError("Unable to create organization user", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.CreateOrganizationUser.Errors)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id := mutation.CreateOrganizationUser.User.Id.(string)

	state := r.read(ctx, id, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state == nil {
		resp.Diagnostics.AddError("Unable to read organization user", "User was created but could not be found.")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *organizationUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := r.read(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if newState == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

// read fetches an organization user by id and maps it to a model, returning nil if
// the user no longer exists.
func (r *organizationUserResource) read(ctx context.Context, id string, diags *diag.Diagnostics) *organizationUserResourceModel {
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
		"id": graphql.ID(id),
	}

	if err := r.client.Query(ctx, &query, variables); err != nil {
		if isRecordNotFound(err) {
			return nil
		}
		diags.AddError("Unable to read organization user", err.Error())
		return nil
	}

	return &organizationUserResourceModel{
		Id:         types.StringValue(query.User.Id.(string)),
		Email:      types.StringValue(string(query.User.Email)),
		Name:       types.StringValue(string(query.User.Name)),
		Role:       types.StringValue(query.User.Role.Id.(string)),
		Phone:      types.StringValue(string(query.User.Phone)),
		ExternalId: types.StringValue(string(query.User.ExternalId)),
		AvatarUrl:  types.StringValue(string(query.User.AvatarUrl)),
		CreatedAt:  types.StringValue(string(query.User.CreatedAt)),
		UpdatedAt:  types.StringValue(string(query.User.UpdatedAt)),
	}
}

type CreateOrganizationUserInput struct {
	Email      graphql.String `json:"email"`
	Name       graphql.String `json:"name,omitempty"`
	Role       graphql.ID     `json:"role"`
	Phone      graphql.String `json:"phone,omitempty"`
	ExternalId graphql.String `json:"externalId,omitempty"`
}

// UpdateUserInput is the updateUser mutation input. Phone/ExternalId are
// non-omitempty pointers: a "" pointer clears them, nil leaves them untouched.
type UpdateUserInput struct {
	Id         graphql.ID      `json:"id"`
	Name       graphql.String  `json:"name,omitempty"`
	Role       graphql.ID      `json:"role,omitempty"`
	Phone      *graphql.String `json:"phone"`
	ExternalId *graphql.String `json:"externalId"`
	AvatarUrl  graphql.String  `json:"avatarUrl,omitempty"`
}

type DeleteUserInput struct {
	Id graphql.ID `json:"id"`
}

// buildUpdateUserInput includes only changed fields. The IsUnknown() guards stop
// an unknown (omitted Computed) plan value from serializing as a wiping "".
func buildUpdateUserInput(plan, state organizationUserResourceModel) UpdateUserInput {
	input := UpdateUserInput{
		Id: graphql.ID(state.Id.ValueString()),
	}

	if !plan.Name.Equal(state.Name) && !plan.Name.IsUnknown() {
		input.Name = graphql.String(plan.Name.ValueString())
	}
	if !plan.Role.Equal(state.Role) && !plan.Role.IsUnknown() {
		input.Role = graphql.ID(plan.Role.ValueString())
	}
	if !plan.Phone.Equal(state.Phone) && !plan.Phone.IsUnknown() {
		if plan.Phone.IsNull() {
			input.Phone = nil
		} else {
			phone := graphql.String(plan.Phone.ValueString())
			input.Phone = &phone
		}
	}
	if !plan.ExternalId.Equal(state.ExternalId) && !plan.ExternalId.IsUnknown() {
		if plan.ExternalId.IsNull() {
			input.ExternalId = nil
		} else {
			externalId := graphql.String(plan.ExternalId.ValueString())
			input.ExternalId = &externalId
		}
	}
	if !plan.AvatarUrl.Equal(state.AvatarUrl) && !plan.AvatarUrl.IsUnknown() {
		input.AvatarUrl = graphql.String(plan.AvatarUrl.ValueString())
	}

	return input
}

func (r *organizationUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationUserResourceModel
	var state organizationUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		UpdateUser struct {
			User struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"updateUser(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": buildUpdateUserInput(plan, state),
	}

	if err := r.client.Mutate(ctx, &mutation, variables); err != nil {
		resp.Diagnostics.AddError("Unable to update organization user", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.UpdateUser.Errors)...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState := r.read(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if newState == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, newState)...)
}

func (r *organizationUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var mutation struct {
		DeleteUser struct {
			User struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"deleteUser(input: $input)"`
	}

	variables := map[string]interface{}{
		"input": DeleteUserInput{
			Id: graphql.ID(state.Id.ValueString()),
		},
	}

	if err := r.client.Mutate(ctx, &mutation, variables); err != nil {
		resp.Diagnostics.AddError("Unable to delete organization user", err.Error())
		return
	}

	resp.Diagnostics.Append(gqlErrorDiagnostics(mutation.DeleteUser.Errors)...)
}

func (r *organizationUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
