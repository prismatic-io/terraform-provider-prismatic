package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// TestUnitResourceOrganizationSigningKeyCreate verifies the signing-key resource
// sends the importOrganizationSigningKey mutation (NOT importIntegration), parses
// the returned id, and reads its attributes back — all against a mocked API, so
// it runs in CI with no credentials.
func TestUnitResourceOrganizationSigningKeyCreate(t *testing.T) {
	client, bodies := mockGraphQL(t, map[string]string{
		"importOrganizationSigningKey": `{"data":{"importOrganizationSigningKey":{"organizationSigningKey":{"id":"KEY-123"},"errors":[]}}}`,
		"signingKeys":                  `{"data":{"organization":{"signingKeys":{"nodes":[{"id":"KEY-123","publicKey":"PUBKEY","imported":true,"issuedAt":"2026-01-01T00:00:00Z"}]}}}}`,
	})

	d := schema.TestResourceDataRaw(t, resourceOrganizationSigningKey().Schema, map[string]interface{}{
		"public_key": "PUBKEY",
	})

	diags := resourceOrganizationSigningKeyCreate(context.Background(), d, client)
	if diags.HasError() {
		t.Fatalf("create returned diagnostics: %v", diags)
	}
	if d.Id() != "KEY-123" {
		t.Errorf("expected id %q, got %q", "KEY-123", d.Id())
	}
	if got := d.Get("public_key").(string); got != "PUBKEY" {
		t.Errorf("expected public_key %q, got %q", "PUBKEY", got)
	}

	sentMutation := false
	for _, b := range *bodies {
		if strings.Contains(b, "importOrganizationSigningKey") {
			sentMutation = true
		}
		if strings.Contains(b, "importIntegration") {
			t.Errorf("signing-key create unexpectedly sent an integration mutation: %s", b)
		}
	}
	if !sentMutation {
		t.Error("expected create to send importOrganizationSigningKey mutation")
	}
}
