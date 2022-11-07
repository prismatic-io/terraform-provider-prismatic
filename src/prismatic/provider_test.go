package prismatic

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"prismatic": func() (*schema.Provider, error) { return Provider(), nil },
	}
	testAccProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PRISMATIC_URL"); v == "" {
		t.Fatal("PRISMATIC_URL must be set for acceptance tests")
	}

	if v := os.Getenv("PRISMATIC_TOKEN"); v == "" {
		t.Fatal("PRISMATIC_TOKEN must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
