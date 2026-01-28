package provider

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/shurcooL/graphql"
)

const (
	organizationUserResourceName = "prismatic_organization_user.test"
	testUserEmail                = "terraform-test-user@example.com"
	testUserName                 = "Terraform Test User"
	testUserUpdatedName          = "Terraform Test User Updated"
)

func organizationUserConfig(email, name, roleRef string) string {
	return fmt.Sprintf(`
data "prismatic_organization_roles" "roles" {}

locals {
  admin_role = [for r in data.prismatic_organization_roles.roles.roles : r if r.name == "Admin"][0]
}

resource "prismatic_organization_user" "test" {
  email       = "%s"
  name        = "%s"
  role        = %s
  external_id = "EXT-TEST-001"
}
`, email, name, roleRef)
}

func TestAccResourceOrganizationUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrganizationUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(organizationUserResourceName, "id"),
					resource.TestCheckResourceAttr(organizationUserResourceName, "email", testUserEmail),
					resource.TestCheckResourceAttr(organizationUserResourceName, "name", testUserName),
					resource.TestCheckResourceAttrSet(organizationUserResourceName, "role"),
					resource.TestCheckResourceAttr(organizationUserResourceName, "external_id", "EXT-TEST-001"),
					resource.TestCheckResourceAttrSet(organizationUserResourceName, "created_at"),
					resource.TestCheckResourceAttrSet(organizationUserResourceName, "updated_at"),
				),
			},
			// Test import
			{
				ResourceName:      organizationUserResourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccResourceOrganizationUser_update(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrganizationUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(organizationUserResourceName, "name", testUserName),
				),
			},
			{
				Config: organizationUserConfig(testUserEmail, testUserUpdatedName, "local.admin_role.id"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(organizationUserResourceName, "name", testUserUpdatedName),
				),
			},
		},
	})
}

func testAccCheckOrganizationUserDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*graphql.Client)

	var query struct {
		Users struct {
			TotalCount int
		} `graphql:"users(email: $email, customer_Isnull: true)"`
	}
	variables := map[string]interface{}{
		"email": graphql.String(testUserEmail),
	}

	if err := client.Query(context.Background(), &query, variables); err != nil {
		return err
	}

	if query.Users.TotalCount != 0 {
		return errors.New("found organization user that should have been deleted")
	}

	return nil
}
