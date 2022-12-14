package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
	"os/exec"
	"path"
	"strings"
)

func resourceComponent() *schema.Resource {
	return &schema.Resource{
		Description:   "Publish a Component to Prismatic. Use the 'Component Bundle' data source to generate the bundle",
		CreateContext: resourceComponentCreate,
		ReadContext:   resourceComponentRead,
		UpdateContext: resourceComponentUpdate,
		DeleteContext: resourceComponentDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the Component",
			},
			"key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The key of the Component",
			},
			"label": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The label of the Component",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The description of the Component",
			},
			"bundle_directory": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Bundled directory. Reference the results of the 'Component Bundle' data source.",
			},
			"bundle_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Bundle path. Reference the results of the 'Component Bundle' data source.",
			},
			"signature": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Bundle signature. Reference the results of the 'Component Bundle' data source.",
			},
		},
	}
}

func resourceComponentCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	componentId, err := publishComponent(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(componentId)

	diags = append(diags, resourceComponentRead(ctx, d, m)...)

	return diags
}

func resourceComponentRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var query struct {
		Component struct {
			Id          graphql.ID
			Key         graphql.String
			Label       graphql.String
			Description graphql.String
		} `graphql:"component(id: $id)"`
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

	if err := d.Set("key", query.Component.Key); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("label", query.Component.Label); err != nil {
		return diag.FromErr(err)
	}
	if err := d.Set("description", query.Component.Description); err != nil {
		return diag.FromErr(err)
	}

	return diags
}

func resourceComponentUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	_, err := publishComponent(client, d)
	if err != nil {
		return diag.FromErr(err)
	}

	diags = append(diags, resourceComponentRead(ctx, d, m)...)

	return diags
}

func resourceComponentDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	// Only clear the Id since it seems unlikely that the consequences of a delete are
	// desired, particularly in resource "tainted" situations.
	d.SetId("")

	return diags
}

// Process the component bundle to capture its data for submission.
func readComponentBundle(path string) (*PublishComponentInput, error) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(nodePath, "-")
	cmd.Dir = path
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

func publishComponent(client *graphql.Client, d *schema.ResourceData) (string, error) {
	bundleDirectory := d.Get("bundle_directory").(string)
	bundle, err := readComponentBundle(bundleDirectory)
	if err != nil {
		return "", err
	}

	packagePath := d.Get("bundle_path").(string)

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

	if err := client.Mutate(context.Background(), &mutation, variables); err != nil {
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
