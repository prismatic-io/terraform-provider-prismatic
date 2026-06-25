package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// TestResourceSchemasValid runs ValidateImplementation over every resource schema.
// No credentials needed.
func TestResourceSchemasValid(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	for _, newResource := range p.Resources(ctx) {
		res := newResource()

		var md resource.MetadataResponse
		res.Metadata(ctx, resource.MetadataRequest{ProviderTypeName: "prismatic"}, &md)

		var resp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%s: Schema returned diagnostics: %+v", md.TypeName, resp.Diagnostics)
			continue
		}
		if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: invalid schema implementation: %+v", md.TypeName, diags)
		}
	}
}

// TestDataSourceSchemasValid is the data-source counterpart of TestResourceSchemasValid.
func TestDataSourceSchemasValid(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	for _, newDataSource := range p.DataSources(ctx) {
		ds := newDataSource()

		var md datasource.MetadataResponse
		ds.Metadata(ctx, datasource.MetadataRequest{ProviderTypeName: "prismatic"}, &md)

		var resp datasource.SchemaResponse
		ds.Schema(ctx, datasource.SchemaRequest{}, &resp)
		if resp.Diagnostics.HasError() {
			t.Errorf("%s: Schema returned diagnostics: %+v", md.TypeName, resp.Diagnostics)
			continue
		}
		if diags := resp.Schema.ValidateImplementation(ctx); diags.HasError() {
			t.Errorf("%s: invalid schema implementation: %+v", md.TypeName, diags)
		}
	}
}
