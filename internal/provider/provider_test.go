package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

var testAccProviders map[string]func() (*schema.Provider, error)
var testAccProvider *schema.Provider

func init() {
	testAccProviders = map[string]func() (*schema.Provider, error){
		"prismatic": func() (*schema.Provider, error) { return New("dev")(), nil },
	}
	testAccProvider = New("dev")()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("PRISMATIC_URL"); v == "" {
		t.Fatal("PRISMATIC_URL must be set for acceptance tests")
	}

	if os.Getenv("PRISMATIC_TOKEN") == "" && os.Getenv("PRISMATIC_REFRESH_TOKEN") == "" {
		t.Fatal("Either PRISMATIC_TOKEN or PRISMATIC_REFRESH_TOKEN must be set for acceptance tests")
	}

	err := testAccProvider.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if err != nil {
		t.Fatal(err)
	}
}

func TestProvider(t *testing.T) {
	if err := New("dev")().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
