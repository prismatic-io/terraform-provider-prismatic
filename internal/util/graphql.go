package util

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/shurcooL/graphql"
	"strings"
)

type GqlErrors []struct {
	Field    graphql.String
	Messages []graphql.String
}

func DiagFromGqlError(errors GqlErrors) diag.Diagnostics {
	var diags diag.Diagnostics

	for _, err := range errors {
		var messages []string
		for _, m := range err.Messages {
			messages = append(messages, string(m))
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Error received from GraphQL operation for field: " + string(err.Field),
			Detail:   strings.Join(messages, "\n"),
		})
	}

	return diags
}
