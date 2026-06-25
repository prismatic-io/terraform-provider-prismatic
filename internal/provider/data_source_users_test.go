package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	usersDataSourceName = "data.prismatic_users.test"
	usersConfig         = `
data "prismatic_users" "test" {
  customer_is_null = true
}
`
	// usersConfigDefault omits customer_is_null to exercise the defaulted path.
	usersConfigDefault = `
data "prismatic_users" "test" {
}
`
)

func TestAccDataSourceUsers_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: usersConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(usersDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(usersDataSourceName, "users.#"),
				),
			},
			{
				Config: usersConfigDefault,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(usersDataSourceName, "users.#"),
					resource.TestCheckResourceAttr(usersDataSourceName, "customer_is_null", "true"),
				),
			},
		},
	})
}
