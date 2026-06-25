package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

const (
	integrationsDataSourceName = "data.prismatic_integrations.integrations"
	integrationsConfig         = `
data "prismatic_integrations" "integrations" {}
`
)

func TestAccDataSourceIntegrations_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: integrationsConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(integrationsDataSourceName, "id"),
					resource.TestCheckResourceAttrSet(integrationsDataSourceName, "integrations.#"),
					resource.TestCheckResourceAttrSet(integrationsDataSourceName, "integrations.0.integration_id"),
					resource.TestCheckResourceAttrSet(integrationsDataSourceName, "integrations.0.integration_name"),
					resource.TestCheckResourceAttrSet(integrationsDataSourceName, "integrations.0.integration_definition"),
				),
			},
		},
	})
}
