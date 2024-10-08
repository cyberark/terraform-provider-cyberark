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
	GetToken(clientID string, clientSecret []byte) (string, error)
}

// IdentityAuthAPI provides methods for fetching identity tokens.
type IdentityAuthAPI struct {
	client *Client
}

// GetIdentityToken fetches an identity token using the provided client ID and client secret.
func (a *IdentityAuthAPI) GetToken(ctx context.Context, clientID string, clientSecret []byte) ([]byte, error) {
	body := strings.NewReader(fmt.Sprintf("client_id=%s&grant_type=client_credentials&client_secret=%s",
		url.QueryEscape(clientID),
		url.QueryEscape(string(clientSecret))))
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	resp, err := a.client.DoRequest(ctx, "POST", "/oauth2/platformtoken", body, headers, map[string]string{})
	body = strings.NewReader("")
	body.Reset("")

	for i := range clientSecret {
		clientSecret[i] = 0
	}

	if err != nil {
		return []byte{}, err
	}
	defer resp.Body.Close()

	var tokenResponse Token
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return []byte{}, err
	}

	if tokenResponse.AccessToken == nil {
		return []byte{}, fmt.Errorf("invalid token response: %v", tokenResponse)
	}

	return []byte(*tokenResponse.AccessToken), nil
}

// NewIdentityAuthAPI creates a new IdentityAuthAPI instance with the provided base URL.
func NewIdentityAuthAPI(baseURL string) *IdentityAuthAPI {
	return &IdentityAuthAPI{
		client: NewClient(baseURL, false),
	}
}
