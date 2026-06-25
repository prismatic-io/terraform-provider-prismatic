package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/shurcooL/graphql"
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
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResourceDestroy,
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
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// definition imports as the canonical form: semantically equal but not
				// textually equal to the submitted value in state, so skip literal compare.
				ImportStateVerifyIgnore: []string{"definition"},
			},
		},
	})
}

func TestAccResourceIntegration_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResourceDestroy,
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

func TestAccResourceIntegration_planStability(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIntegrationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: resourceWithDefinition(baseDefinition),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckIntegrationResourceDestroy(s *terraform.State) error {
	client, err := testAccGraphQLClient()
	if err != nil {
		return err
	}

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
