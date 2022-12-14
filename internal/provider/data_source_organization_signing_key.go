package provider

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
	"log"
)

func dataSourceOrganizationSigningKey() *schema.Resource {
	return &schema.Resource{
		Description: "Data source to retrieve an Organization's signing key",
		ReadContext: dataSourceOrganizationSigningKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The ID of the signing key to fetch.",
			},
			"imported": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Indicates if the signing key was imported or generated",
			},
			"issued_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The timestamp the signing key was issued at",
			},
			"public_key": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The public key of the signing key",
			},
		},
	}
}

func dataSourceOrganizationSigningKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var query struct {
		Organization struct {
			SigningKeys struct {
				Nodes []struct {
					Id        string
					Imported  bool
					IssuedAt  string
					PublicKey string
				}
			}
		}
	}

	targetID := fmt.Sprintf("%v", d.Get("id"))

	if err := client.Query(context.Background(), &query, nil); err != nil {
		return diag.FromErr(err)
	}

	targetSigningKey := make(map[string]interface{})

	for _, signingKey := range query.Organization.SigningKeys.Nodes {
		if signingKey.Id == targetID {
			targetSigningKey["id"] = signingKey.Id
			targetSigningKey["issued_at"] = signingKey.IssuedAt
			targetSigningKey["public_key"] = signingKey.PublicKey
			targetSigningKey["imported"] = signingKey.Imported
			break
		}
	}

	if !d.IsNewResource() && targetSigningKey["id"] == nil {
		log.Printf("organization signing key (%s) not found!", targetID)
		d.SetId("")
		return nil
	}

	for key, value := range targetSigningKey {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("%v", targetSigningKey["id"]))

	return diags
}
