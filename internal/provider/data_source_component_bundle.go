package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
)

func dataSourceComponentBundle() *schema.Resource {
	return &schema.Resource{
		Description: "Generates a component bundle suitable for publishing",
		ReadContext: dataSourceComponentBundleRead,
		Schema: map[string]*schema.Schema{
			"bundle_directory": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Directory to bundle",
			},
			"bundle_path": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Destination of the generated bundle",
			},
			"signature": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Signature of the bundle for detecting redundant publishes",
			},
		},
	}
}

func dataSourceComponentBundleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	bundleDirectory := d.Get("bundle_directory").(string)
	bundlePath := d.Get("bundle_path").(string)

	_, packageSignature, err := util.GenerateBundleSignature(bundleDirectory, bundlePath)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("signature", packageSignature); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(resource.UniqueId())

	return diags
}
