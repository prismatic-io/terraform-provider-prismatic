package provider

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"testing"
)

const (
	organizationSigningKeyDataSourceName = "data.prismatic_organization_signing_key.organizations_signing_key"
	organizationSigningKeyConfig         = `
data "prismatic_organization_signing_key" "organization_signing_key" {}
`
)

func TestAccDataSourceOrganizationSigningKey_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: organizationSigningKeyConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(organizationSigningKeyDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(organizationSigningKeyDataSourceName, "public_key"),
					resource.TestCheckResourceAttrSet(organizationSigningKeyDataSourceName, "imported"),
					resource.TestCheckResourceAttrSet(organizationSigningKeyDataSourceName, "issued_at"),
				),
			},
		},
	})
}
