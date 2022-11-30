package prismatic

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const (
	organizationsSigningKeyDataSourceName = "data.prismatic_organizations_signing_key.organizations_signing_key"
	organizationsSigningKeyConfig         = `
data "prismatic_organizations_signing_key" "organizations_signing_key" {}
`
)

func TestAccDataSourceOrganizationsSigningKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: organizationsSigningKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(organizationsSigningKeyDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(organizationsSigningKeyDataSourceName, "public_key"),
					resource.TestCheckResourceAttrSet(organizationsSigningKeyDataSourceName, "imported"),
					resource.TestCheckResourceAttrSet(organizationsSigningKeyDataSourceName, "issued_at"),
				),
			},
		},
	})
}
