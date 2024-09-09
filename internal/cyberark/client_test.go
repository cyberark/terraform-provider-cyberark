package cyberark_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aharriscybr/terraform-provider-cybr-sh/internal/cyberark"
	"github.com/stretchr/testify/assert"
)

func TestDoRequest(t *testing.T) {
	t.Run("DoRequest", func(t *testing.T) {
		token := []byte("dummy_token")
		additionalHeaderKey := "Test-Header"
		additionalHeaderValue := "Test-Value"
		body := "test body"
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("Bearer %s", token), req.Header.Get("Authorization"))
			assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
			assert.Equal(t, additionalHeaderValue, req.Header.Get(additionalHeaderKey))

			requestBody, _ := io.ReadAll(req.Body)
			assert.Equal(t, body, string(requestBody))

			rw.Write([]byte(`{"response": "test response"}`))
		}))
		defer server.Close()

		client := cyberark.NewClientWithToken(server.URL, true, token)

		headers := map[string]string{
			additionalHeaderKey: additionalHeaderValue,
		}

		resp, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			strings.NewReader(body),
			headers,
			map[string]string{},
		)

		assert.NoError(t, err)
		responseBody, _ := io.ReadAll(resp.Body)
		assert.Equal(t, `{"response": "test response"}`, string(responseBody))
	})

	t.Run("WithoutAuthToken", func(t *testing.T) {
		body := `{"response": "test response"}`
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Empty(t, req.Header.Get("Authorization"))
			rw.Write([]byte(body))
		}))
		defer server.Close()

		client := cyberark.NewClient(server.URL, false)

		resp, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			nil,
			nil,
			map[string]string{},
		)

		assert.NoError(t, err)

		responseBody, _ := io.ReadAll(resp.Body)
		assert.Equal(t, body, string(responseBody))
	})

	t.Run("InvalidMethod", func(t *testing.T) {
		client := cyberark.NewClient("http://localhost:12345", false)

		_, err := client.DoRequest(
			context.Background(),
			"INVALID",
			"/test",
			nil,
			nil,
			map[string]string{},
		)

		assert.Error(t, err)
	})

	t.Run("ServerNotReachable", func(t *testing.T) {
		client := cyberark.NewClient("http://localhost:12345", false) // Non-existent server

		_, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			nil,
			nil,
			map[string]string{},
		)

		assert.Error(t, err)
	})

	t.Run("MissingScheme", func(t *testing.T) {
		client := cyberark.NewClient("invalid-url/api", false) // Invalid URL

		_, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			nil,
			nil,
			map[string]string{},
		)
		assert.Error(t, err)

	})

	t.Run("MissingHost", func(t *testing.T) {
		client := cyberark.NewClient("http://", false) // Invalid URL

		_, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			nil,
			nil,
			map[string]string{},
		)
		assert.Error(t, err)

	})
	t.Run("MissingUrl", func(t *testing.T) {
		client := cyberark.NewClient("", false) // Missing URL

		_, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test",
			nil,
			nil,
			map[string]string{},
		)
		assert.Error(t, err)

	})
	t.Run("MissingUrl", func(t *testing.T) {
		client := cyberark.NewClient("", false) // Missing URL

		_, err := client.DoRequest(
			context.Background(),
			"POST",
			"/test/a/b",
			nil,
			nil,
			map[string]string{},
		)
		assert.Error(t, err)

	})
}
