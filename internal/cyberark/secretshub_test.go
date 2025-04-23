package cyberark_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/stretchr/testify/assert"
)

func TestAddAwsAsmSecretStore(t *testing.T) {
	var (
		secretStoreName = "test_store"
		input           = cyberark.SecretStoreInput[cyberark.AwsAsmData]{
			Name: &secretStoreName,
		}
		body = cyberark.SecretStoreOutput[cyberark.AwsAsmData]{
			ID: "test_store_id",
		}
	)

	t.Run("AddAwsAsmSecretStore", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusCreated)
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAwsAsmSecretStore(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAwsAsmSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("SecretStoreAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Secret Store already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAwsAsmSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("MissingID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAwsAsmSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestAddAzureAkvSecretStore(t *testing.T) {
	var (
		secretStoreName = "test_store"
		input           = cyberark.SecretStoreInput[cyberark.AzureAkvData]{
			Name: &secretStoreName,
		}
		body = cyberark.SecretStoreOutput[cyberark.AzureAkvData]{
			ID: "test_store_id",
		}
	)

	t.Run("AddAzureAkvSecretStore", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusCreated)
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAzureAkvSecretStore(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAzureAkvSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("SecretStoreAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Secret Store already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAzureAkvSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("MissingID", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`{}`))
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddAzureAkvSecretStore(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetAwsAsmSecretStore(t *testing.T) {
	var (
		token = []byte("dummy_token")
		body  = cyberark.SecretStoreOutput[cyberark.AwsAsmData]{
			ID:   "test_store_id",
			Data: &cyberark.AwsAsmData{},
		}
		storeID = "test_store_id"
	)

	t.Run("GetAwsAsmSecretStore", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("/api/secret-stores/%s", storeID), req.URL.Path)
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAwsAsmSecretStore(context.Background(), storeID)
		assert.NoError(t, err)

		assert.Equal(t, &body, resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAwsAsmSecretStore(context.Background(), storeID)
		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetAzureAkvSecretStore(t *testing.T) {
	var (
		token = []byte("dummy_token")
		body  = cyberark.SecretStoreOutput[cyberark.AzureAkvData]{
			ID:   "test_store_id",
			Data: &cyberark.AzureAkvData{},
		}
		storeID = "test_store_id"
	)

	t.Run("GetAzureAkvSecretStore", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("/api/secret-stores/%s", storeID), req.URL.Path)
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAzureAkvSecretStore(context.Background(), storeID)
		assert.NoError(t, err)

		assert.Equal(t, &body, resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAzureAkvSecretStore(context.Background(), storeID)
		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetAwsAsmSecretStores(t *testing.T) {
	var (
		body = cyberark.SecretStoresOutput[cyberark.AwsAsmData]{
			SecretStores: []*cyberark.SecretStoreOutput[cyberark.AwsAsmData]{
				{
					ID: "test_store_id",
				},
			},
		}
	)

	t.Run("GetAwsAsmSecretStores", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "type EQ AWS_ASM", req.URL.Query().Get("filter"))
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetAwsAsmSecretStores(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetAwsAsmSecretStores(context.Background())

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetAzureAkvSecretStores(t *testing.T) {
	var (
		body = cyberark.SecretStoresOutput[cyberark.AzureAkvData]{
			SecretStores: []*cyberark.SecretStoreOutput[cyberark.AzureAkvData]{
				{
					ID: "test_store_id",
				},
			},
		}
	)

	t.Run("GetAzureAkvSecretStores", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "type EQ AZURE_AKV", req.URL.Query().Get("filter"))
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetAzureAkvSecretStores(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetAzureAkvSecretStores(context.Background())

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestUpdateSecretStore(t *testing.T) {
	t.Run("UpdateAwsSecretStore", func(t *testing.T) {
		var (
			storeID = "test-store-id"
			name    = "updated_aws_store"
			input   = cyberark.SecretStoreInput[cyberark.AwsAsmData]{
				Name: &name,
			}
			output = cyberark.SecretStoreOutput[cyberark.AwsAsmData]{
				ID:   storeID,
				Name: &name,
			}
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, fmt.Sprintf("/api/secret-stores/%s", storeID), req.URL.Path)
			json.NewEncoder(rw).Encode(output)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.UpdateAwsSecretStore(context.Background(), storeID, input)

		assert.NoError(t, err)
		assert.Equal(t, output, *resp)
	})

	t.Run("UpdateAwsSecretStoreError", func(t *testing.T) {
		var (
			storeID = "test-store-id"
			name    = "updated_aws_store"
			input   = cyberark.SecretStoreInput[cyberark.AwsAsmData]{
				Name: &name,
			}
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.UpdateAwsSecretStore(context.Background(), storeID, input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("UpdateAzureAkvSecretStore", func(t *testing.T) {
		var (
			storeID = "test-store-id"
			name    = "updated_azure_store"
			input   = cyberark.SecretStoreInput[cyberark.AzureAkvData]{
				Name: &name,
			}
			output = cyberark.SecretStoreOutput[cyberark.AzureAkvData]{
				ID:   storeID,
				Name: &name,
			}
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "PATCH", req.Method)
			assert.Equal(t, fmt.Sprintf("/api/secret-stores/%s", storeID), req.URL.Path)
			json.NewEncoder(rw).Encode(output)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.UpdateAzureAkvSecretStore(context.Background(), storeID, input)

		assert.NoError(t, err)
		assert.Equal(t, output, *resp)
	})

	t.Run("UpdateAzureAkvSecretStoreError", func(t *testing.T) {
		var (
			storeID = "test-store-id"
			name    = "updated_azure_store"
			input   = cyberark.SecretStoreInput[cyberark.AzureAkvData]{
				Name: &name,
			}
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.UpdateAzureAkvSecretStore(context.Background(), storeID, input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestDeleteSecretStore(t *testing.T) {
	t.Run("DeleteSecretStore", func(t *testing.T) {
		var (
			storeID = "test-store-id"
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, http.MethodDelete, req.Method)
			assert.Equal(t, fmt.Sprintf("/api/secret-stores/%s", storeID), req.URL.Path)
			rw.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		err := client.DeleteSecretStore(context.Background(), storeID)
		assert.NoError(t, err)
	})

	t.Run("DeleteSecretStoreError", func(t *testing.T) {
		var (
			storeID = "test-store-id"
		)

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		err := client.DeleteSecretStore(context.Background(), storeID)
		assert.Error(t, err)
	})
}

func TestAddSyncPolicy(t *testing.T) {
	var (
		policy   = "test_policy"
		policyID = "policy-62d19762-85d0-4cc0-ba44-9e0156a5c9c6"
		input    = cyberark.PolicyInput{
			Name: &policy,
		}
		body = cyberark.PolicyExternalOutput{
			ID:   &policyID,
			Name: &policy,
		}
	)
	t.Run("AddSyncPolicy", func(t *testing.T) {

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.WriteHeader(http.StatusCreated)
			json.NewEncoder(rw).Encode(body)

		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddSyncPolicy(context.Background(), input)

		assert.NoError(t, err)

		assert.Equal(t, body, *resp)
	})

	t.Run("PolicyAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Secret Store already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddSyncPolicy(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
		assert.Equal(t, "HTTP status code 409", err.Error())
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddSyncPolicy(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.AddSyncPolicy(context.Background(), input)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetSyncPolicy(t *testing.T) {
	var (
		policy   = "test_policy"
		policyID = "policy-62d19762-85d0-4cc0-ba44-9e0156a5c9c6"
		body     = cyberark.PolicyExternalOutput{
			Name: &policy,
		}
	)

	t.Run("GetSyncPolicy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("/api/policies/%s", policyID), req.URL.Path)
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetSyncPolicy(context.Background(), policyID)

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetSyncPolicy(context.Background(), policyID)
		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetSyncPolicies(t *testing.T) {
	t.Run("GetSyncPolicies", func(t *testing.T) {
		body := cyberark.SyncResponse{
			Policies: []*cyberark.PolicyExternalOutput{},
		}

		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetSyncPolicies(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, body, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.GetSyncPolicies(context.Background())
		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestDeleteSyncPolicy(t *testing.T) {
	var (
		policyID = "policy-62d19762-85d0-4cc0-ba44-9e0156a5c9c6"
	)

	t.Run("DeleteSyncPolicy", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch {
			case req.Method == http.MethodPut && req.URL.Path == fmt.Sprintf("/api/policies/%s/state", policyID):
				// Verify the disable request is sent with correct payload
				var requestBody map[string]string
				json.NewDecoder(req.Body).Decode(&requestBody)
				assert.Equal(t, "disable", requestBody["action"])
				rw.WriteHeader(http.StatusOK)
			case req.Method == http.MethodDelete && req.URL.Path == fmt.Sprintf("/api/policies/%s", policyID):
				// Verify the delete request - should return 200 OK per implementation
				rw.WriteHeader(http.StatusOK)
			default:
				http.Error(rw, "Unexpected request", http.StatusBadRequest)
			}
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		err := client.DeleteSyncPolicy(context.Background(), policyID)
		assert.NoError(t, err)
	})

	t.Run("DisableFailsButDeleteSucceeds", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch {
			case req.Method == http.MethodPut && req.URL.Path == fmt.Sprintf("/api/policies/%s/state", policyID):
				// Simulate disable failure
				http.Error(rw, "Failed to disable", http.StatusInternalServerError)
			case req.Method == http.MethodDelete && req.URL.Path == fmt.Sprintf("/api/policies/%s", policyID):
				// Delete still succeeds - should return 200 OK per implementation
				rw.WriteHeader(http.StatusOK)
			default:
				http.Error(rw, "Unexpected request", http.StatusBadRequest)
			}
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		err := client.DeleteSyncPolicy(context.Background(), policyID)
		assert.NoError(t, err) // Should still succeed as delete worked
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			switch {
			case req.Method == http.MethodPut:
				// Disable succeeds
				rw.WriteHeader(http.StatusOK)
			case req.Method == http.MethodDelete:
				// Delete fails
				http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
			}
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		err := client.DeleteSyncPolicy(context.Background(), policyID)
		assert.Error(t, err)
	})
}
