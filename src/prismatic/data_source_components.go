package prismatic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
	"strconv"
	"time"
)

func dataSourceComponents() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComponentsRead,
		Schema: map[string]*schema.Schema{
			"components": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"component_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"component_key": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"component_label": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"component_description": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceComponentsRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*graphql.Client)

	var diags diag.Diagnostics

	var query struct {
		Components struct {
			Nodes []struct {
				Id          string
				Key         string
				Label       string
				Description string
			}
		}
	}

	if err := client.Query(context.Background(), &query, nil); err != nil {
		return diag.FromErr(err)
	}

	count := len(query.Components.Nodes)
	components := make([]interface{}, count, count)
	for i, componentNode := range query.Components.Nodes {
		component := make(map[string]interface{})
		component["component_id"] = componentNode.Id
		component["component_key"] = componentNode.Key
		component["component_label"] = componentNode.Label
		component["component_description"] = componentNode.Description
		components[i] = component
	}

	if err := d.Set("components", components); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
