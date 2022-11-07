package prismatic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"strconv"
	"terraform-provider/prismatic/util"
	"time"
)

func dataSourceComponentBundle() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceComponentBundleRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"bundle_directory": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bundle_path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"signature": {
				Type:     schema.TypeString,
				Computed: true,
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

	d.SetId(strconv.FormatInt(time.Now().Unix(), 10))

	return diags
}
