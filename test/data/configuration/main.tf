terraform {
  required_providers {
    prismatic = {
      source  = "prismatic-io/prismatic"
      version = "~> 0.1.1"
    }
  }
}

provider "prismatic" {}

data "prismatic_integrations" "integrations" {}

data "prismatic_components" "components" {}

resource "prismatic_integration" "pokemon" {
  definition = file("../integrations/pokemon.yml")
}

data "prismatic_component_bundle" "bundle" {
  bundle_directory = "../component/code"
  bundle_path      = "../component/bundle.zip"
}

resource "prismatic_component" "component" {
  bundle_directory = data.prismatic_component_bundle.bundle.bundle_directory
  bundle_path      = data.prismatic_component_bundle.bundle.bundle_path
  signature        = data.prismatic_component_bundle.bundle.signature
}

resource "prismatic_integration" "test" {
  depends_on = [prismatic_component.component]

  definition = file("../integrations/test.yml")
}

resource "prismatic_organization_signing_key" "signing_key" {
    public_key = file("../organization/public_key.pub")
}
