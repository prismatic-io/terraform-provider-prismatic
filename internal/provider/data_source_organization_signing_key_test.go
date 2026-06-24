package provider

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	organizationSigningKeyDataSourceName = "data.prismatic_organization_signing_key.organization_signing_key"
)

func TestAccDataSourceOrganizationSigningKey_basic(t *testing.T) {
	// The data source requires an id, so provision a signing key and look it up by id.
	config := resourceWithPubkey(expectedPubKey) + fmt.Sprintf(`
data "prismatic_organization_signing_key" "organization_signing_key" {
  id = %s.id
}
`, "prismatic_organization_signing_key.key")

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: testAccProviders,
		CheckDestroy:      testAccCheckOrganizationSigningKeyResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
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
