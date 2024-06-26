// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// TokenFetcher is an interface for fetching identity tokens.
type TokenFetcher interface {
	GetIdentityToken(clientID, clientSecret string) (string, error)
}

// AuthAPI provides methods for fetching identity tokens.
type AuthAPI struct {
	client *Client
}

// GetIdentityToken fetches an identity token using the provided client ID and client secret.
func (a *AuthAPI) GetIdentityToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	body := strings.NewReader(fmt.Sprintf("client_id=%s&grant_type=client_credentials&client_secret=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(clientSecret)))
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}

	resp, err := a.client.DoRequest(ctx, "POST", "/oauth2/platformtoken", body, headers)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var tokenResponse Token
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", err
	}

	if tokenResponse.AccessToken == nil {
		return "", fmt.Errorf("invalid token response: %v", tokenResponse)
	}

	return *tokenResponse.AccessToken, nil
}

// NewAuthAPI creates a new AuthAPI instance with the provided base URL.
func NewAuthAPI(baseURL string) *AuthAPI {
	return &AuthAPI{
		client: NewClient(baseURL, false),
	}
}
