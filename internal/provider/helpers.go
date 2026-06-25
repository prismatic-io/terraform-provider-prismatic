package provider

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/prismatic-io/terraform-provider-prismatic/internal/util"
	"github.com/shurcooL/graphql"
)

// clientFromProviderData extracts the configured GraphQL client passed to a resource
// or data source during configuration. It returns nil before the provider has been
// configured (ProviderData is nil) and records a diagnostic if the data is an
// unexpected type.
func clientFromProviderData(providerData interface{}, diags *diag.Diagnostics) *graphql.Client {
	if providerData == nil {
		return nil
	}
	client, ok := providerData.(*graphql.Client)
	if !ok {
		diags.AddError(
			"Unexpected Provider Data Type",
			"Expected *graphql.Client. This is a bug in the provider; please report it.",
		)
		return nil
	}
	return client
}

// gqlErrorDiagnostics converts the user-facing field errors returned by a Prismatic
// GraphQL mutation into framework diagnostics, one per field.
func gqlErrorDiagnostics(errs util.GqlErrors) diag.Diagnostics {
	var diags diag.Diagnostics
	for _, e := range errs {
		messages := make([]string, 0, len(e.Messages))
		for _, m := range e.Messages {
			messages = append(messages, string(m))
		}
		diags.AddError("GraphQL error for field: "+string(e.Field), strings.Join(messages, "\n"))
	}
	return diags
}

// isRecordNotFound reports whether err is the Prismatic API's "Record not found"
// response, which the provider treats as the record having been deleted out of band.
func isRecordNotFound(err error) bool {
	return err != nil && strings.Contains(err.Error(), "Record not found")
}
