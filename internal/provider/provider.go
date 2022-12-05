package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
	"net/url"
)

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"url": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("PRISMATIC_URL", "https://app.prismatic.io"),
					Description: "URL of the Prismatic stack to communicate with. Defaults to the value of the `PRISMATIC_URL` environment variable.",
				},
				"token": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("PRISMATIC_TOKEN", ""),
					Description: "An [access token to use for headless authentication](https://prismatic.io/docs/cli/cli-usage/#headless-prism-usage-for-cicd-pipelines) of Prismatic API calls.",
				},
			},
			ResourcesMap: map[string]*schema.Resource{
				"prismatic_component":                resourceComponent(),
				"prismatic_integration":              resourceIntegration(),
				"prismatic_organization_signing_key": resourceOrganizationSigningKey(),
			},
			DataSourcesMap: map[string]*schema.Resource{
				"prismatic_component_bundle":         dataSourceComponentBundle(),
				"prismatic_components":               dataSourceComponents(),
				"prismatic_integrations":             dataSourceIntegrations(),
				"prismatic_organization_signing_key": dataSourceOrganizationSigningKey(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)

		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		baseUrl := d.Get("url").(string)
		token := d.Get("token").(string)

		var diags diag.Diagnostics

		if baseUrl == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create a Prismatic client",
				Detail:   "Unable to create a Prismatic client without a url.",
			})
		}
		if token == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create a Prismatic client",
				Detail:   "Unable to create a Prismatic client without an authorization token. Please either pass in an authorization token to the Prismatic provider, or set an environment variable, PRISMATIC_TOKEN",
			})
		}

		if diags != nil && diags.HasError() {
			return nil, diags
		}

		u, err := url.Parse(baseUrl)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		u.Path = "api"
		apiUrl := u.String()

		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient := oauth2.NewClient(context.Background(), src)

		client := graphql.NewClient(apiUrl, httpClient)
		return client, diags
	}
}
