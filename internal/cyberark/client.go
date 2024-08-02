// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Client is a client for interacting with the SecretsHub APIs.
type Client struct {
	httpClient  *http.Client
	baseURL     string
	AuthToken   string
	logResponse bool
}

// DoRequest sends an HTTP request to the CyberArk API.
func (c *Client) DoRequest(ctx context.Context, method string, path string, body io.Reader, headers map[string]string, params map[string]string) (*http.Response, error) {
	relativeURL, err := JoinURL(c.baseURL, path, params)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, relativeURL, body)
	if err != nil {
		return nil, err
	}

	if c.AuthToken != "" {
		// Set the Authorization header to include the auth token.
		req.Header.Set("Authorization", "Bearer "+c.AuthToken)
	}

	req.Header.Add("Content-Type", "application/json")

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	response, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if c.logResponse {
		responseBody, err := io.ReadAll(response.Body)
		if err == nil && len(responseBody) > 0 {
			tflog.Debug(
				ctx,
				"Response from CyberArk API",
				map[string]interface{}{
					"request_url":     req.URL,
					"method":          method,
					"response_status": response.Status,
					"response_body":   string(responseBody),
				},
			)
		}
		response.Body.Close()

		// Replace the response body with a new reader that contains the original data
		response.Body = io.NopCloser(bytes.NewBuffer(responseBody))
	}

	return response, nil
}

// NewClient creates a new Client instance with the provided base URL.
func NewClient(baseURL string, logResponse bool) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     baseURL,
		logResponse: logResponse,
	}
}

// NewClientWithToken creates a new Client instance with the provided base URL and auth token.
func NewClientWithToken(baseURL string, logResponse bool, authToken string) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:     baseURL,
		logResponse: logResponse,
		AuthToken:   authToken,
	}
}

// JoinURL constructs a URL by joining the base URL with the provided path segments.
func JoinURL(baseURL string, path string, params map[string]string) (string, error) {
	baseURI, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", err
	}

	relativePath, err := url.Parse(path)
	if err != nil {
		return "", err
	}

	finalURL := baseURI.JoinPath(relativePath.String())

	if len(params) != 0 {
		queryParams := url.Values{}
		for key, value := range params {
			queryParams.Add(key, value)
		}
		finalURL.RawQuery = queryParams.Encode()
	}

	return finalURL.String(), nil
}
