package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	authenticatedUserDataSourceName = "data.prismatic_authenticated_user.test"
	authenticatedUserConfig         = `
data "prismatic_authenticated_user" "test" {}
`
)

func TestAccDataSourceAuthenticatedUser_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: authenticatedUserConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(authenticatedUserDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(authenticatedUserDataSourceName, "email"),
					resource.TestCheckResourceAttrSet(authenticatedUserDataSourceName, "org_id"),
					resource.TestCheckResourceAttrSet(authenticatedUserDataSourceName, "org_name"),
				),
			},
		},
	})
}
