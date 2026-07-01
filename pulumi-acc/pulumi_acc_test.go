// Package pulumiacc holds the Pulumi analog of the Terraform acceptance suite: it bridges the
// freshly-built provider through pulumi-terraform-bridge and runs a real create -> preview -> destroy
// round-trip against a live Prismatic API. See README.md.
package pulumiacc

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/pulumi/providertest/pulumitest"
	"github.com/pulumi/providertest/pulumitest/assertpreview"
	"github.com/pulumi/providertest/pulumitest/opttest"
)

// TestPulumiAcc drives create -> preview(no-diff) -> destroy through Pulumi against a live Prismatic
// API. Like the Terraform suite it self-skips unless its gate env var (PULUMI_ACC) is set, and it
// requires live credentials — the provider reads PRISMATIC_URL / PRISMATIC_TOKEN /
// PRISMATIC_REFRESH_TOKEN / PRISMATIC_TENANT_ID from the environment.
//
// pulumitest manages the local (file://) state backend, the config passphrase, `pulumi install`
// (which regenerates the bridged SDK from Pulumi.yaml), stack creation, and teardown — so this needs
// no Pulumi Cloud account.
func TestPulumiAcc(t *testing.T) {
	if os.Getenv("PULUMI_ACC") == "" {
		t.Skip("PULUMI_ACC not set; skipping the Pulumi acceptance test")
	}
	if os.Getenv("PRISMATIC_URL") == "" {
		t.Fatal("PRISMATIC_URL must be set for the Pulumi acceptance test")
	}
	if os.Getenv("PRISMATIC_TOKEN") == "" && os.Getenv("PRISMATIC_REFRESH_TOKEN") == "" {
		t.Fatal("Either PRISMATIC_TOKEN or PRISMATIC_REFRESH_TOKEN must be set for the Pulumi acceptance test")
	}

	// Build the provider where Pulumi.yaml's relative packages path expects it (../bin). `go test`
	// runs with the package directory as the working directory, so ".." is the repo root.
	bin, err := filepath.Abs(filepath.Join("..", "bin", "terraform-provider-prismatic"))
	if err != nil {
		t.Fatal(err)
	}
	build := exec.Command("go", "build", "-o", bin, "..")
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("building the provider binary: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}

	test := pulumitest.NewPulumiTest(t, wd, opttest.TestInPlace(), opttest.RequireYarnLinks(false))
	t.Cleanup(func() { test.Destroy(t) })

	// CREATE + READ. pulumitest fails the test if the deploy errors; Destroy on cleanup exercises
	// delete. The signing key's computed `id` must come back populated, proving the provider both
	// created the resource and read it back through the bridge.
	up := test.Up(t)
	if id, ok := up.Outputs["id"].Value.(string); !ok || id == "" {
		t.Fatalf("expected a non-empty signing key id output, got %#v", up.Outputs["id"].Value)
	}

	// A second plan must be empty — the provider round-trips its state cleanly through the bridge,
	// matching Terraform/OpenTofu's post-apply idempotency. (Relies on index.ts using a fixed key.)
	assertpreview.HasNoChanges(t, test.Preview(t))
}
