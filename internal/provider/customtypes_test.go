package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func newDefinitionValue(s string) definitionStringValue {
	return definitionStringValue{StringValue: basetypes.NewStringValue(s)}
}

func newNormalizedValue(s string) normalizedStringValue {
	return normalizedStringValue{StringValue: basetypes.NewStringValue(s)}
}

// TestDefinitionStringValueSemanticEquals drives the method directly (receiver =
// canonical, arg = submitted). The superset case locks in the one-directional fix.
func TestDefinitionStringValueSemanticEquals(t *testing.T) {
	cases := []struct {
		name      string
		canonical string // receiver: API-normalized read-back
		submitted string // argument: value from config / prior state
		wantEqual bool
	}{
		{"submitted is a subset of the normalized read-back", integrationCanonicalV7, integrationSubmittedV7, true},
		{"identical definitions", integrationCanonicalV7, integrationCanonicalV7, true},
		{
			name:      "submitted superset with a dropped field is NOT suppressed",
			canonical: "name: X",
			submitted: "name: X\nsomeRealField: importantValue",
			wantEqual: false,
		},
		{
			name:      "changed scalar is a real diff",
			canonical: "version: 3",
			submitted: "version: 2",
			wantEqual: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			receiver := newDefinitionValue(tc.canonical)
			arg := newDefinitionValue(tc.submitted)

			got, diags := receiver.StringSemanticEquals(context.Background(), arg)
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %+v", diags)
			}
			if got != tc.wantEqual {
				t.Errorf("StringSemanticEquals = %v, want %v", got, tc.wantEqual)
			}
		})
	}
}

func TestNormalizedStringValueSemanticEquals(t *testing.T) {
	cases := []struct {
		name      string
		a         string
		b         string
		wantEqual bool
	}{
		{"whitespace-only differences are equal", "key\n", "  key  ", true},
		{"identical values are equal", "key", "key", true},
		{"different content is not equal", "key-a", "key-b", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, diags := newNormalizedValue(tc.a).StringSemanticEquals(context.Background(), newNormalizedValue(tc.b))
			if diags.HasError() {
				t.Fatalf("unexpected diagnostics: %+v", diags)
			}
			if got != tc.wantEqual {
				t.Errorf("StringSemanticEquals = %v, want %v", got, tc.wantEqual)
			}
		})
	}
}
