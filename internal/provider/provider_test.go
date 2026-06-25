package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/shurcooL/graphql"
)

// testAccProtoV6ProviderFactories serves the framework provider for acceptance tests.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"prismatic": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PRISMATIC_URL"); v == "" {
		t.Fatal("PRISMATIC_URL must be set for acceptance tests")
	}

	if os.Getenv("PRISMATIC_TOKEN") == "" && os.Getenv("PRISMATIC_REFRESH_TOKEN") == "" {
		t.Fatal("Either PRISMATIC_TOKEN or PRISMATIC_REFRESH_TOKEN must be set for acceptance tests")
	}
}

// testAccGraphQLClient builds a client from the same environment variables the
// provider reads, for use in CheckDestroy functions (which run outside provider
// configuration).
func testAccGraphQLClient() (*graphql.Client, error) {
	return newGraphQLClient(
		os.Getenv("PRISMATIC_URL"),
		os.Getenv("PRISMATIC_TOKEN"),
		os.Getenv("PRISMATIC_REFRESH_TOKEN"),
		os.Getenv("PRISMATIC_TENANT_ID"),
	)
}
