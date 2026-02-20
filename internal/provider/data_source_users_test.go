package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	usersDataSourceName = "data.prismatic_users.test"
	usersConfig         = `
data "prismatic_users" "test" {
  customer_is_null = true
}
`
)

func TestAccDataSourceUsers_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: usersConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(usersDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(usersDataSourceName, "users.#"),
				),
			},
		},
	})
}
