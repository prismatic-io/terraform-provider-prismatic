package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

var (
	_ resource.Resource              = (*componentResource)(nil)
	_ resource.ResourceWithConfigure = (*componentResource)(nil)
)

type componentResource struct {
	client *graphql.Client
}

type componentResourceModel struct {
	Id              types.String `tfsdk:"id"`
	Key             types.String `tfsdk:"key"`
	Label           types.String `tfsdk:"label"`
	Description     types.String `tfsdk:"description"`
	BundleDirectory types.String `tfsdk:"bundle_directory"`
	BundlePath      types.String `tfsdk:"bundle_path"`
	Signature       types.String `tfsdk:"signature"`
}

func (r *componentResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_component"
}

func (r *componentResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Publish a Component to Prismatic. Use the 'Component Bundle' data source to generate the bundle",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The ID of the Component",
			},
			"key": schema.StringAttribute{
				Computed:    true,
				Description: "The key of the Component",
			},
			"label": schema.StringAttribute{
				Computed:    true,
				Description: "The label of the Component",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "The description of the Component",
			},
			"bundle_directory": schema.StringAttribute{
				Required:    true,
				Description: "Bundled directory. Reference the results of the 'Component Bundle' data source.",
			},
			"bundle_path": schema.StringAttribute{
				Required:    true,
				Description: "Bundle path. Reference the results of the 'Component Bundle' data source.",
			},
			"signature": schema.StringAttribute{
				Required:    true,
				Description: "Bundle signature. Reference the results of the 'Component Bundle' data source.",
			},
		},
	}
}

func (r *componentResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.client = clientFromProviderData(req.ProviderData, &resp.Diagnostics)
}

func (r *componentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan componentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	componentId, err := publishComponent(ctx, r.client, plan.BundleDirectory.ValueString(), plan.BundlePath.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to publish component", err.Error())
		return
	}

	// Publishing is processed asynchronously, so the Component is not queryable the
	// instant the mutation returns. Poll until it is available before reading it.
	if err := waitForComponent(ctx, r.client, componentId); err != nil {
		resp.Diagnostics.AddError("Unable to publish component", err.Error())
		return
	}

	state := r.read(ctx, componentId, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if state == nil {
		resp.Diagnostics.AddError("Unable to publish component", "Component was published but could not be found.")
		return
	}
	state.copyBundleInputsFrom(plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *componentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state componentResourceModel
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
	updated.copyBundleInputsFrom(state)

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *componentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan componentResourceModel
	var state componentResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// publishComponent upserts by key. Re-read by the prior id so the resource id
	// stays immutable across updates rather than adopting the id the publish returns.
	if _, err := publishComponent(ctx, r.client, plan.BundleDirectory.ValueString(), plan.BundlePath.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to publish component", err.Error())
		return
	}

	if err := waitForComponent(ctx, r.client, state.Id.ValueString()); err != nil {
		resp.Diagnostics.AddError("Unable to publish component", err.Error())
		return
	}

	updated := r.read(ctx, state.Id.ValueString(), &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	if updated == nil {
		resp.Diagnostics.AddError("Unable to publish component", "Component was published but could not be found.")
		return
	}
	updated.copyBundleInputsFrom(plan)

	resp.Diagnostics.Append(resp.State.Set(ctx, updated)...)
}

func (r *componentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Only clear state since it seems unlikely that the consequences of a delete are
	// desired, particularly in resource "tainted" situations.
	resp.State.RemoveResource(ctx)
}

// read queries the Component by id and maps its server-owned fields into a model,
// returning nil if the Component no longer exists. The bundle inputs are not part of
// the API response, so the caller restores them with copyBundleInputsFrom.
func (r *componentResource) read(ctx context.Context, id string, diags *diag.Diagnostics) *componentResourceModel {
	var query struct {
		Component struct {
			Id          graphql.ID
			Key         graphql.String
			Label       graphql.String
			Description graphql.String
		} `graphql:"component(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": graphql.ID(id),
	}
	if err := r.client.Query(ctx, &query, variables); err != nil {
		if isRecordNotFound(err) {
			return nil
		}
		diags.AddError("Unable to read component", err.Error())
		return nil
	}

	return &componentResourceModel{
		Id:          types.StringValue(query.Component.Id.(string)),
		Key:         types.StringValue(string(query.Component.Key)),
		Label:       types.StringValue(string(query.Component.Label)),
		Description: types.StringValue(string(query.Component.Description)),
	}
}

// copyBundleInputsFrom carries the configuration-only bundle inputs (which the API
// does not return) over from the plan or prior state into a freshly read model.
func (m *componentResourceModel) copyBundleInputsFrom(src componentResourceModel) {
	m.BundleDirectory = src.BundleDirectory
	m.BundlePath = src.BundlePath
	m.Signature = src.Signature
}

// waitForComponent polls until the Component with the given id is queryable,
// tolerating the brief "not found" window after a publish.
func waitForComponent(ctx context.Context, client *graphql.Client, id string) error {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		var query struct {
			Component struct {
				Id graphql.ID
			} `graphql:"component(id: $id)"`
		}
		variables := map[string]interface{}{"id": graphql.ID(id)}
		err := client.Query(ctx, &query, variables)
		if err == nil {
			return nil
		}
		if !isRecordNotFound(err) {
			return err
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("component %q not yet available after publish", id)
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}

// readComponentBundle runs the bundle through node to capture its definition and
// actions for submission to the publishComponent mutation.
func readComponentBundle(bundlePath string) (*PublishComponentInput, error) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(nodePath, "-")
	cmd.Dir = bundlePath
	cmd.Stdin = strings.NewReader("const bundle = require(\".\"); console.log(JSON.stringify(bundle.default));")
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		errStr := stderr.String()
		return nil, fmt.Errorf("node exec failed: %s (%s)", err, errStr)
	}

	var result map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &result)
	if err != nil {
		return nil, err
	}

	var actionInputs []interface{}
	if val, ok := result["actions"]; ok {
		actionsMap := val.(map[string]interface{})
		for _, v := range actionsMap {
			actionInputs = append(actionInputs, v)
		}
	}

	definition := result
	delete(definition, "actions")

	input := PublishComponentInput{
		Definition: definition,
		Actions:    actionInputs,
	}
	return &input, nil
}

func publishComponent(ctx context.Context, client *graphql.Client, bundleDirectory string, packagePath string) (string, error) {
	bundle, err := readComponentBundle(bundleDirectory)
	if err != nil {
		return "", err
	}

	var mutation struct {
		PublishComponent struct {
			PublishResult struct {
				Component struct {
					Id graphql.ID
				}
				IconUploadUrl    graphql.String
				PackageUploadUrl graphql.String
			}
		} `graphql:"publishComponent (input: $input)"`
	}
	variables := map[string]interface{}{
		"input": PublishComponentInput{
			Definition: bundle.Definition,
			Actions:    bundle.Actions,
		},
	}

	if err := client.Mutate(ctx, &mutation, variables); err != nil {
		return "", err
	}

	definitionDisplay := bundle.Definition["display"].(map[string]interface{})
	iconPath := path.Join(bundleDirectory, definitionDisplay["iconPath"].(string))
	if err := util.UploadFile(iconPath, string(mutation.PublishComponent.PublishResult.IconUploadUrl), "image/png"); err != nil {
		return "", err
	}

	if err := util.UploadFile(packagePath, string(mutation.PublishComponent.PublishResult.PackageUploadUrl), "application/zip"); err != nil {
		return "", err
	}

	return mutation.PublishComponent.PublishResult.Component.Id.(string), nil
}

type PublishComponentInput struct {
	Definition map[string]interface{} `json:"definition" graphql:"DefinitionInput!"`
	Actions    []interface{}          `json:"actions" graphql:"[ActionDefinitionInput]!"`
}
