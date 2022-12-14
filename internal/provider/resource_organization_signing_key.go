package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
	"log"
	"strings"
)

func resourceOrganizationSigningKey() *schema.Resource {
	return &schema.Resource{
		Description:   "Import a public key into the Organization's Signing Keys",
		CreateContext: resourceOrganizationSigningKeyCreate,
		ReadContext:   resourceOrganizationSigningKeyRead,
		DeleteContext: resourceOrganizationSigningKeyDelete,
		Schema: map[string]*schema.Schema{
			"public_key": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressPubKeyWhitespaceDiff,
				ForceNew:         true,
				Description:      "Public key to import",
			},
			"imported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if signing key was imported or generated",
			},
			"issued_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp of when the signing key was issued",
			},
			"id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The ID of the signing key",
			},
		},
	}
}

func resourceOrganizationSigningKeyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var mutation struct {
		ImportOrganizationSigningKey struct {
			OrganizationSigningKey struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"importOrganizationSigningKey (input: {publicKey: $publicKey})"`
	}

	mutationVars := map[string]interface{}{
		"publicKey": graphql.String(d.Get("public_key").(string)),
	}

	if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
		return diag.FromErr(err)
	}

	if len(mutation.ImportOrganizationSigningKey.Errors) > 0 {
		return util.DiagFromGqlError(mutation.ImportOrganizationSigningKey.Errors)
	}

	d.SetId(mutation.ImportOrganizationSigningKey.OrganizationSigningKey.Id.(string))
	d.Set("public_key", d.Get("public_key"))

	// Update state from remote
	diags = append(diags, resourceOrganizationSigningKeyRead(ctx, d, m)...)

	return diags
}

func resourceOrganizationSigningKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

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

	targetID := fmt.Sprintf("%v", d.Id())

	if err := client.Query(context.Background(), &query, nil); err != nil {
		return diag.FromErr(err)
	}

	targetSigningKey := make(map[string]interface{})

	// API does not support filtering on this query so we loop through IDs until we find one.
	for _, signingKey := range query.Organization.SigningKeys.Nodes {
		if signingKey.Id == targetID {
			targetSigningKey["id"] = signingKey.Id
			targetSigningKey["public_key"] = signingKey.PublicKey
			break
		}
	}

	if !d.IsNewResource() && targetSigningKey["id"] == nil {
		log.Printf("[WARN] organization signing key (%s) not found, removing from state", targetID)
		d.SetId("")
		return nil
	}

	for key, value := range targetSigningKey {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	return diags
}

func resourceOrganizationSigningKeyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)
	var diags diag.Diagnostics

	var mutation struct {
		DeleteOrganizationSigningKey struct {
			OrganizationSigningKey struct {
				Id graphql.ID
			}
			Errors util.GqlErrors
		} `graphql:"deleteOrganizationSigningKey (input: {id: $id})"`
	}

	mutationVars := map[string]interface{}{
		"id": graphql.ID(d.Id()),
	}

	if err := client.Mutate(context.Background(), &mutation, mutationVars); err != nil {
		return diag.FromErr(err)
	}

	if len(mutation.DeleteOrganizationSigningKey.Errors) > 0 {
		return util.DiagFromGqlError(mutation.DeleteOrganizationSigningKey.Errors)
	}

	d.SetId("")

	return diags
}

func suppressPubKeyWhitespaceDiff(k, old, new string, d *schema.ResourceData) bool {
	return strings.TrimSpace(old) == strings.TrimSpace(new)
}
