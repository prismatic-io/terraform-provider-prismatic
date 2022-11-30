package prismatic

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
)

func dataSourceOrganizationsSigningKey() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceOrganizationsSigningKeyRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"imported": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"issued_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"public_key": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceOrganizationsSigningKeyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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

	for key, value := range targetSigningKey {
		if err := d.Set(key, value); err != nil {
			return diag.FromErr(err)
		}
	}

	d.SetId(fmt.Sprintf("%v", targetSigningKey["id"]))

	return diags
}
