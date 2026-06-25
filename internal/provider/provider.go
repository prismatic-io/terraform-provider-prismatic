package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = (*prismaticProvider)(nil)

// New returns the Prismatic provider.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &prismaticProvider{version: version}
	}
}

type prismaticProvider struct {
	version string
}

type providerModel struct {
	Url          types.String `tfsdk:"url"`
	Token        types.String `tfsdk:"token"`
	RefreshToken types.String `tfsdk:"refresh_token"`
	TenantId     types.String `tfsdk:"tenant_id"`
}

func (p *prismaticProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "prismatic"
	resp.Version = p.version
}

func (p *prismaticProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Optional:    true,
				Description: "URL of the Prismatic stack to communicate with. Defaults to the value of the `PRISMATIC_URL` environment variable.",
			},
			"token": schema.StringAttribute{
				Optional:           true,
				Sensitive:          true,
				DeprecationMessage: "Access token use has been deprecated in favor of using refresh tokens. Please migrate provider configuration to use the new refresh_token attribute instead.",
				Description:        "An [access token obtained with Prism CLI](https://prismatic.io/docs/cli/prism/#metoken) of Prismatic API calls.",
			},
			"refresh_token": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "A [refresh token to use for headless authentication](https://prismatic.io/docs/cli/bash-scripting/#headless-prism-usage-for-cicd-pipelines) to the Prismatic API.",
			},
			"tenant_id": schema.StringAttribute{
				Optional:    true,
				Description: "The [tenant ID to authenticate against](https://prismatic.io/docs/cli/bash-scripting/#headless-prism-usage-for-cicd-pipelines) when a refresh token grants access to multiple tenants. If omitted, it is left out of the token exchange.",
			},
		},
	}
}

func (p *prismaticProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config providerModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Each value falls back from its configured attribute to the environment, then
	// to the documented default.
	baseUrl := stringWithEnvFallback(config.Url, "PRISMATIC_URL", "https://app.prismatic.io")
	token := stringWithEnvFallback(config.Token, "PRISMATIC_TOKEN", "")
	refreshToken := stringWithEnvFallback(config.RefreshToken, "PRISMATIC_REFRESH_TOKEN", "")
	tenantId := stringWithEnvFallback(config.TenantId, "PRISMATIC_TENANT_ID", "")

	if baseUrl == "" {
		resp.Diagnostics.AddError("Unable to create a Prismatic client", "Unable to create a Prismatic client without a url.")
	}
	if token == "" && refreshToken == "" {
		resp.Diagnostics.AddError("Unable to create a Prismatic client", "Unable to create a Prismatic client without an authorization token or a refresh token. Please either pass in an authorization token or a refresh_token to the Prismatic provider. Optionally, you can set a environment variable, PRISMATIC_TOKEN or PRISMATIC_REFRESH_TOKEN")
	}
	if resp.Diagnostics.HasError() {
		return
	}

	client, err := newGraphQLClient(baseUrl, token, refreshToken, tenantId)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create a Prismatic client", err.Error())
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client
}

func (p *prismaticProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource { return &componentResource{} },
		func() resource.Resource { return &integrationResource{} },
		func() resource.Resource { return &organizationSigningKeyResource{} },
		func() resource.Resource { return &organizationUserResource{} },
	}
}

func (p *prismaticProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		func() datasource.DataSource { return &authenticatedUserDataSource{} },
		func() datasource.DataSource { return &componentBundleDataSource{} },
		func() datasource.DataSource { return &componentsDataSource{} },
		func() datasource.DataSource { return &integrationsDataSource{} },
		func() datasource.DataSource { return &organizationRolesDataSource{} },
		func() datasource.DataSource { return &organizationSigningKeyDataSource{} },
		func() datasource.DataSource { return &usersDataSource{} },
	}
}

func stringWithEnvFallback(v types.String, env, fallback string) string {
	if !v.IsNull() && !v.IsUnknown() {
		return v.ValueString()
	}
	if e := os.Getenv(env); e != "" {
		return e
	}
	return fallback
}
