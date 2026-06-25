package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/shurcooL/graphql"
)

func gqlStrPtr(s string) *graphql.String {
	v := graphql.String(s)
	return &v
}

func phonePtrEqual(got, want *graphql.String) bool {
	if got == nil || want == nil {
		return got == nil && want == nil
	}
	return *got == *want
}

// TestBuildUpdateUserInput guards the wipe regression: an untouched or unknown
// phone/external_id must not be sent (nil ptr), while "" clears and a value updates.
func TestBuildUpdateUserInput(t *testing.T) {
	const base = "+14155552671"

	cases := []struct {
		name           string
		plan           organizationUserResourceModel
		state          organizationUserResourceModel
		wantPhone      *graphql.String
		wantExternalID *graphql.String
	}{
		{
			name:           "unchanged values are not sent",
			plan:           organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			state:          organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			wantPhone:      nil,
			wantExternalID: nil,
		},
		{
			name:           "unknown planned values are not sent (the wipe regression)",
			plan:           organizationUserResourceModel{Phone: types.StringUnknown(), ExternalId: types.StringUnknown()},
			state:          organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			wantPhone:      nil,
			wantExternalID: nil,
		},
		{
			name:           "changed values are sent",
			plan:           organizationUserResourceModel{Phone: types.StringValue("+14155559999"), ExternalId: types.StringValue("ext-2")},
			state:          organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			wantPhone:      gqlStrPtr("+14155559999"),
			wantExternalID: gqlStrPtr("ext-2"),
		},
		{
			name:           "cleared to empty string is sent as empty (explicit clear)",
			plan:           organizationUserResourceModel{Phone: types.StringValue(""), ExternalId: types.StringValue("")},
			state:          organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			wantPhone:      gqlStrPtr(""),
			wantExternalID: gqlStrPtr(""),
		},
		{
			name:           "null planned value leaves the field untouched",
			plan:           organizationUserResourceModel{Phone: types.StringNull(), ExternalId: types.StringNull()},
			state:          organizationUserResourceModel{Phone: types.StringValue(base), ExternalId: types.StringValue("ext-1")},
			wantPhone:      nil,
			wantExternalID: nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.state.Id = types.StringValue("user-123")
			input := buildUpdateUserInput(tc.plan, tc.state)

			if graphql.ID("user-123") != input.Id {
				t.Errorf("Id = %v, want user-123", input.Id)
			}
			if !phonePtrEqual(input.Phone, tc.wantPhone) {
				t.Errorf("Phone = %v, want %v", input.Phone, tc.wantPhone)
			}
			if !phonePtrEqual(input.ExternalId, tc.wantExternalID) {
				t.Errorf("ExternalId = %v, want %v", input.ExternalId, tc.wantExternalID)
			}
		})
	}
}

// TestBuildUpdateUserInputUnknownName covers the name/avatar_url unknown branches.
func TestBuildUpdateUserInputUnknownName(t *testing.T) {
	plan := organizationUserResourceModel{Name: types.StringUnknown(), AvatarUrl: types.StringUnknown()}
	state := organizationUserResourceModel{
		Id:        types.StringValue("user-123"),
		Name:      types.StringValue("Ada"),
		AvatarUrl: types.StringValue("https://example.com/a.png"),
	}

	input := buildUpdateUserInput(plan, state)

	if input.Name != "" {
		t.Errorf("Name = %q, want empty (unknown must not be sent)", input.Name)
	}
	if input.AvatarUrl != "" {
		t.Errorf("AvatarUrl = %q, want empty (unknown must not be sent)", input.AvatarUrl)
	}
}

func TestValidateE164Phone(t *testing.T) {
	testCases := []struct {
		name        string
		phone       string
		expectError bool
	}{
		{"valid US number", "+14155552671", false},
		{"valid min length", "+1234567", false},
		{"valid max length", "+123456789012345", false},
		{"empty string", "", false},
		{"missing plus", "14155552671", true},
		{"starts with zero", "+0123456789", true},
		{"too short", "+123456", true},
		{"too long", "+1234567890123456", true},
		{"only plus", "+", true},
		{"letters included", "+1415abc2671", true},
		{"spaces included", "+1 415 555 2671", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.StringRequest{ConfigValue: types.StringValue(tc.phone)}
			resp := &validator.StringResponse{}
			e164PhoneValidator{}.ValidateString(context.Background(), req, resp)
			hasError := resp.Diagnostics.HasError()
			if hasError != tc.expectError {
				t.Errorf("validateE164Phone(%q): expected error=%v, got error=%v", tc.phone, tc.expectError, hasError)
			}
		})
	}
}
