package cyberark_test

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/stretchr/testify/assert"
)

func TestGetToken(t *testing.T) {
	t.Run("GetToken", func(t *testing.T) {
		clientID := "test_client_id"
		clientSecret := []byte("test_client_secret")
		token := []byte("dummy_token")

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			// Read the request body
			body, _ := io.ReadAll(req.Body)
			req.Body.Close()

			assert.Equal(t, "application/x-www-form-urlencoded", req.Header.Get("Content-Type"))

			assert.Contains(t, string(body), fmt.Sprintf("client_id=%s", clientID))
			assert.Contains(t, string(body), fmt.Sprintf("client_secret=%s", clientSecret))

			rw.Write([]byte(fmt.Sprintf(`{"access_token": "%s"}`, token)))
		}))
		defer server.Close()

		// Create a new AuthApi instance with the test server's URL
		identityAuthAPI := cyberark.NewIdentityAuthAPI(server.URL)

		// Call GetToken and check the returned token and error
		resp, err := identityAuthAPI.GetToken(context.Background(), clientID, clientSecret)

		assert.NoError(t, err)
		assert.Equal(t, token, resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		identityAuthAPI := cyberark.NewIdentityAuthAPI(server.URL)

		resp, err := identityAuthAPI.GetToken(context.Background(), "test_client_id", []byte("test_client_secret"))

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		identityAuthAPI := cyberark.NewIdentityAuthAPI(server.URL)

		resp, err := identityAuthAPI.GetToken(context.Background(), "test_client_id", []byte("test_client_secret"))

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("NoAccessTokenInResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`{"no_access_token": "not_a_token"}`))
		}))
		defer server.Close()

		identityAuthAPI := cyberark.NewIdentityAuthAPI(server.URL)

		resp, err := identityAuthAPI.GetToken(context.Background(), "test_client_id", []byte("test_client_secret"))

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}
