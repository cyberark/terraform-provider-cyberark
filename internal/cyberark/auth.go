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
	GetToken(ctx context.Context, clientID string, clientSecret []byte) ([]byte, error)
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

	var tokenResponse IdentityToken
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
		client: NewClient(baseURL, false, true),
	}
}

type PVWAAuthAPIError struct {
	ErrorCode    string `json:"ErrorCode"`
	ErrorMessage string `json:"ErrorMessage"`
}

type PVWAAuthAPI struct {
	client      *Client
	loginMethod string
}

// GetToken fetches a PAM API token using the provided username and password.
func (a *PVWAAuthAPI) GetToken(ctx context.Context, clientID string, clientSecret []byte) ([]byte, error) {
	body := strings.NewReader(
		fmt.Sprintf(
			`{"username": "%s", "password": "%s"}`,
			clientID,
			string(clientSecret),
		),
	)
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	logonPath := fmt.Sprintf("/PasswordVault/API/auth/%s/Logon/", a.loginMethod)

	resp, err := a.client.DoRequest(ctx, "POST", logonPath, body, headers, map[string]string{})
	body = strings.NewReader("")
	body.Reset("")

	for i := range clientSecret {
		clientSecret[i] = 0
	}

	if err != nil {
		return []byte{}, err
	}

	if resp.StatusCode != 200 {
		var pvwaAuthError PVWAAuthAPIError
		if err := json.NewDecoder(resp.Body).Decode(&pvwaAuthError); err != nil {
			return []byte{}, fmt.Errorf("failed to decode PVWA Auth API error response: %w", err)
		}
		return []byte{}, fmt.Errorf("%s %s", pvwaAuthError.ErrorCode, pvwaAuthError.ErrorMessage)
	}
	defer resp.Body.Close()

	var token string
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return []byte{}, err
	}

	if token == "" {
		return []byte{}, fmt.Errorf("invalid token response")
	}

	return []byte(token), nil
}

func NewPVWAAuthAPI(baseURL string, loginMethod string) *PVWAAuthAPI {
	return &PVWAAuthAPI{
		client:      NewClient(baseURL, false, false),
		loginMethod: loginMethod,
	}
}
