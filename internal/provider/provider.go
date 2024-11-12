package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
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
					Description: "An [access token to use for headless authentication](https://prismatic.io/docs/cli/cli-usage/#headless-prism-usage-for-cicd-pipelines) of Prismatic API calls. Refresh token parameter is not going to be used if token is provided.",
				},
				"refresh_token": {
					Type:        schema.TypeString,
					Optional:    true,
					DefaultFunc: schema.EnvDefaultFunc("PRISMATIC_REFRESH_TOKEN", ""),
					Description: "A [refresh token to use for headless authentication](https://prismatic.io/docs/cli/cli-usage/#headless-prism-usage-for-cicd-pipelines), of Prismatic API calls. Token parameter is not going to be used if refresh token is provided, a new access token will be requested using the refresh token provided.",
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

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (any, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (any, diag.Diagnostics) {
		baseUrl := d.Get("url").(string)
		token := d.Get("token").(string)
		refreshToken := d.Get("refresh_token").(string)

		var diags diag.Diagnostics

		if baseUrl == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create a Prismatic client",
				Detail:   "Unable to create a Prismatic client without a url.",
			})
		}
		if token == "" && refreshToken == "" {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Unable to create a Prismatic client",
				Detail:   "Unable to create a Prismatic client without an authorization token or a refresh token. Please either pass in an authorization token or a refresh_token to the Prismatic provider. Optionally, you can set a environment variable, PRISMATIC_TOKEN or PRISMATIC_REFRESH_TOKEN",
			})
		}

		if diags != nil && diags.HasError() {
			return nil, diags
		}

		u, err := url.Parse(baseUrl)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		if refreshToken != "" {
			accessToken, err := refreshAccessToken(u, RefreshTokenRequest{RefreshToken: refreshToken})
			if err != nil {
				return nil, diag.FromErr(err)
			}
			token = *accessToken
		}

		u.Path = "api"
		apiUrl := u.String()

		src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		httpClient := oauth2.NewClient(context.Background(), src)

		client := graphql.NewClient(apiUrl, httpClient)
		return client, diags
	}
}

func refreshAccessToken(baseUrl *url.URL, refreshToken RefreshTokenRequest) (*string, error) {
	baseUrl.Path = "/auth/refresh"
	apiUrl := baseUrl.String()

	body, err := json.Marshal(refreshToken)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to refresh access token: %s", resp.Status)
	}

	var result struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result.AccessToken, nil
}
