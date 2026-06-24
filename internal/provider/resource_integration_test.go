package provider

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shurcooL/graphql"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	resourceName        = "prismatic_integration.integration"
	expectedName        = "Acceptance Test"
	expectedUpdatedName = "Acceptance Test Updated"
	expectedDescription = "Acceptance Test Integration"

	baseDefinition = `
definitionVersion: 7
name: Acceptance Test
description: Acceptance Test Integration
requiredConfigVars: []
flows:
  - name: Flow 1
    isSynchronous: false
    steps:
      - name: Integration Trigger
        isTrigger: true
        action:
          component:
            key: webhook-triggers
            version: LATEST
            isPublic: true
          key: webhook
        inputs: {}`
	updateDefinition = `
definitionVersion: 7
name: Acceptance Test Updated
description: Acceptance Test Integration
requiredConfigVars: []
flows:
  - name: Flow 1
    isSynchronous: false
    steps:
      - name: Integration Trigger
        isTrigger: true
        action:
          component:
            key: webhook-triggers
            version: LATEST
            isPublic: true
          key: webhook
        inputs: {}`
)

func resourceWithDefinition(definition string) string {
	return fmt.Sprintf(`
resource "prismatic_integration" "integration" {
  definition = <<EOF
%s
EOF
}`, definition)
}

func TestAccResourceIntegration_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceWithDefinition(baseDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestCheckResourceAttr(resourceName, "name", expectedName),
					resource.TestCheckResourceAttr(resourceName, "description", expectedDescription),
				),
			},
		},
	})
}

func TestAccResourceIntegration_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceWithDefinition(baseDefinition),
			},
			{
				Config: resourceWithDefinition(updateDefinition),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "id"),
					resource.TestCheckResourceAttrSet(resourceName, "definition"),
					resource.TestCheckResourceAttr(resourceName, "name", expectedUpdatedName),
					resource.TestCheckResourceAttr(resourceName, "description", expectedDescription),
				),
			},
			{
				Config: resourceWithDefinition(baseDefinition),
			},
		},
	})
}

func testAccCheckIntegrationResourceDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*graphql.Client)

	var query struct {
		Integrations struct {
			TotalCount int
		} `graphql:"integrations(name_Icontains: $name)"`
	}
	variables := map[string]interface{}{
		"name": graphql.String(expectedName),
	}
	if err := client.Query(context.Background(), &query, variables); err != nil {
		return err
	}

	if query.Integrations.TotalCount != 0 {
		return errors.New("found integration that should have been deleted")
	}

	return nil
}

// FIXME: Generative test to ensure positive/negative behaviors?
func TestSuppressDiffIntegrationDefinition_same(t *testing.T) {
	first := `description: Acceptance Test Integration
isSynchronous: false
name: Acceptance Test
requiredConfigVars: []
steps: []
trigger:
  description: ""
  name: 'Integration Trigger'
  schedule: null`
	second := `"description": Acceptance Test Integration
isSynchronous: false
trigger:
  description: ''
  name: Integration Trigger
  schedule: !!null null
name: Acceptance Test
requiredConfigVars: []
steps: []`
	result := suppressDiffIntegrationDefinition("foo", first, second, nil)
	if !result {
		t.Fatalf("Did not suppress diff for logically identical definitions")
	}
}
