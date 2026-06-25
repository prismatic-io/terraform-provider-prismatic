package util

import (
	"github.com/shurcooL/graphql"
)

// GqlErrors is the shape of the user-facing error list returned by Prismatic
// GraphQL mutations.
type GqlErrors []struct {
	Field    graphql.String
	Messages []graphql.String
}
