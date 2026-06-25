package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/shurcooL/graphql"
	"golang.org/x/oauth2"
)

type RefreshTokenRequest struct {
	RefreshToken string  `json:"refresh_token"`
	TenantId     *string `json:"tenant_id,omitempty"`
}

// newGraphQLClient builds an authenticated GraphQL client for the Prismatic API.
// When a refresh token is supplied it is first exchanged for an access token.
func newGraphQLClient(baseUrl, token, refreshToken, tenantId string) (*graphql.Client, error) {
	u, err := url.Parse(baseUrl)
	if err != nil {
		return nil, err
	}

	if refreshToken != "" {
		var tid *string
		if tenantId != "" {
			tid = &tenantId
		}
		accessToken, err := refreshAccessToken(u, RefreshTokenRequest{RefreshToken: refreshToken, TenantId: tid})
		if err != nil {
			return nil, err
		}
		token = *accessToken
	}

	u.Path = "api"
	src := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	httpClient := oauth2.NewClient(context.Background(), src)
	return graphql.NewClient(u.String(), httpClient), nil
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
	defer func() { _ = resp.Body.Close() }()

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
