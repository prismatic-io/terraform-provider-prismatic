package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	organizationRolesDataSourceName = "data.prismatic_organization_roles.test"
	organizationRolesConfig         = `
data "prismatic_organization_roles" "test" {}
`
)

func TestAccDataSourceOrganizationRoles_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: organizationRolesConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "roles.#"),
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "roles.0.id"),
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "roles.0.name"),
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "roles.0.description"),
					resource.TestCheckResourceAttrSet(organizationRolesDataSourceName, "roles.0.level"),
				),
			},
		},
	})
}
