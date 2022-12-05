package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const (
	expectedLabel = "Component label"
)

func TestAccResourceComponent_basic(t *testing.T) {
	resourceName := "prismatic_component.component"
	expectedKey := "componentKey"
	expectedLabel := "Component label"
	expectedDescription := "Component description"
	initial := `
data "prismatic_component_bundle" "bundle" {
    bundle_directory = "test-fixtures/component/code"
    bundle_path = "test-fixtures/component/bundle.zip"
}

resource "prismatic_component" "component" {
    bundle_directory = data.prismatic_component_bundle.bundle.bundle_directory
    bundle_path = data.prismatic_component_bundle.bundle.bundle_path
    signature = data.prismatic_component_bundle.bundle.signature
}`

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
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

func TestReadComponentBundle(t *testing.T) {
	result, err := readComponentBundle("test-fixtures/component/code/")
	if err != nil {
		t.Fatalf("Failed to read component bundle: %s", err)
	}
	if result == nil {
		t.Fatalf("Received nil result from bundle read")
	}
}
