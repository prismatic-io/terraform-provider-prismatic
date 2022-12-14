package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const (
	componentsDataSourceName = "data.prismatic_components.components"
	componentsConfig         = `
data "prismatic_components" "components" {}`
)

func TestAccDataSourceComponents_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: componentsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(componentsDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(componentsDataSourceName, "components.#"),
					resource.TestCheckResourceAttrSet(componentsDataSourceName, "components.0.component_key"),
					resource.TestCheckResourceAttrSet(componentsDataSourceName, "components.0.component_label"),
					resource.TestCheckResourceAttrSet(componentsDataSourceName, "components.0.component_description"),
				),
			},
		},
	})
}
