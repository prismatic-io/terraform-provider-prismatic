package provider

import (
	"context"
	"errors"
	"fmt"
	"regexp"
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

func organizationUserConfig(email, name, roleRef, phone, externalID string) string {
	optionalFields := ""
	if phone != "" {
		optionalFields += fmt.Sprintf("\n  phone       = %q", phone)
	}
	if externalID != "" {
		optionalFields += fmt.Sprintf("\n  external_id = %q", externalID)
	}

	return fmt.Sprintf(`
data "prismatic_organization_roles" "roles" {}

locals {
  admin_role = [for r in data.prismatic_organization_roles.roles.roles : r if r.name == "Admin"][0]
}

resource "prismatic_organization_user" "test" {
  email = "%s"
  name  = "%s"
  role  = %s%s
}
`, email, name, roleRef, optionalFields)
}

func TestAccResourceOrganizationUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrganizationUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id", "", "EXT-TEST-001"),
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
				Config: organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id", "", "EXT-TEST-001"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(organizationUserResourceName, "name", testUserName),
				),
			},
			{
				Config: organizationUserConfig(testUserEmail, testUserUpdatedName, "local.admin_role.id", "", "EXT-TEST-001"),
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

func TestAccResourceOrganizationUser_invalidPhoneFormat(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id", "12345", ""),
				ExpectError: regexp.MustCompile(`Invalid phone number format`),
			},
			{
				Config:      organizationUserConfig(testUserEmail, testUserName, "local.admin_role.id", "+0123456789", ""),
				ExpectError: regexp.MustCompile(`Invalid phone number format`),
			},
		},
	})
}

func TestValidateE164Phone(t *testing.T) {
	testCases := []struct {
		name        string
		phone       string
		expectError bool
	}{
		{"valid US number", "+14155552671", false},
		{"valid min length", "+1234567", false},
		{"valid max length", "+123456789012345", false},
		{"empty string", "", false},
		{"missing plus", "14155552671", true},
		{"starts with zero", "+0123456789", true},
		{"too short", "+123456", true},
		{"too long", "+1234567890123456", true},
		{"only plus", "+", true},
		{"letters included", "+1415abc2671", true},
		{"spaces included", "+1 415 555 2671", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			diags := validateE164Phone(tc.phone, nil)
			hasError := diags.HasError()
			if hasError != tc.expectError {
				t.Errorf("validateE164Phone(%q): expected error=%v, got error=%v", tc.phone, tc.expectError, hasError)
			}
		})
	}
}
