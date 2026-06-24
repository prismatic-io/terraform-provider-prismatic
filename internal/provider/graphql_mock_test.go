package provider

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/shurcooL/graphql"
)

// mockGraphQL stands up an httptest server that returns canned responses keyed by
// a substring of the request body (typically a GraphQL operation name), so
// resource CRUD can be unit tested without a live Prismatic API. It returns a
// graphql.Client pointed at the server and a pointer to the slice of request
// bodies it received, for asserting which operations were sent.
func mockGraphQL(t *testing.T, responses map[string]string) (*graphql.Client, *[]string) {
	t.Helper()
	var bodies []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		for marker, resp := range responses {
			if strings.Contains(string(body), marker) {
				io.WriteString(w, resp)
				return
			}
		}
		t.Errorf("unexpected GraphQL operation: %s", body)
		io.WriteString(w, `{"data":{}}`)
	}))
	t.Cleanup(server.Close)
	return graphql.NewClient(server.URL, server.Client()), &bodies
}
