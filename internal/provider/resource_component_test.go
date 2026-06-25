package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccResourceComponent_basic(t *testing.T) {
	resourceName := "prismatic_component.component"
	expectedKey := "componentKey"
	expectedLabel := "Component label"
	expectedDescription := "Component description"
	initial := `
data "prismatic_component_bundle" "bundle" {
    bundle_directory = "../../test/data/component/code"
    bundle_path = "../../test/data/component/bundle.zip"
}

resource "prismatic_component" "component" {
    bundle_directory = data.prismatic_component_bundle.bundle.bundle_directory
    bundle_path = data.prismatic_component_bundle.bundle.bundle_path
    signature = data.prismatic_component_bundle.bundle.signature
}`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: initial,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "signature"),
					resource.TestCheckResourceAttr(resourceName, "key", expectedKey),
					resource.TestCheckResourceAttr(resourceName, "label", expectedLabel),
					resource.TestCheckResourceAttr(resourceName, "description", expectedDescription),
				),
			},
		},
	})
}
