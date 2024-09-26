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
		input           = cyberark.SecretStoreInput[cyberark.CreateAzureAkvData]{
			Name: &secretStoreName,
		}
		body = cyberark.SecretStoreOutput[cyberark.CreateAzureAkvData]{
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
	t.Run("EmptyData", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body.Data = nil
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAwsAsmSecretStore(context.Background(), storeID)
		assert.Empty(t, resp)
		assert.Error(t, err)
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
		body  = cyberark.SecretStoreOutput[cyberark.CreateAzureAkvData]{
			ID:   "test_store_id",
			Data: &cyberark.CreateAzureAkvData{},
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
	t.Run("EmptyData", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body.Data = nil
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, token)

		resp, err := client.GetAzureAkvSecretStore(context.Background(), storeID)
		assert.Empty(t, resp)
		assert.Error(t, err)
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
		body = cyberark.SecretStoresOutput[cyberark.CreateAzureAkvData]{
			SecretStores: []*cyberark.SecretStoreOutput[cyberark.CreateAzureAkvData]{
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
	t.Run("VerifyNoOp", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		client.UpdateSecretStore(context.Background())
	})
}

func TestDeleteSecretStore(t *testing.T) {
	t.Run("VerifyNoOp", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		client.DeleteSecretStore(context.Background())
	})
}

func TestScanDefinition(t *testing.T) {
	var (
		input = cyberark.TriggerScanInputBody{
			Scope: cyberark.ScanScope{
				Scan: []string{"store_id"},
			},
		}
		body = cyberark.TriggerScanOutput{
			ScanIDs: []string{"scan_store_id"},
		}
	)

	t.Run("ScanDefinition", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "application/x.secretshub.beta+json", req.Header.Get("Accept"))
			json.NewEncoder(rw).Encode(body)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.ScanDefinition(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, body, resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		resp, err := client.ScanDefinition(context.Background(), input)
		assert.Empty(t, resp)
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
		assert.NoError(t, err)
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

func TestUpdateSyncPolicy(t *testing.T) {
	t.Run("VerifyNoOp", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		client.UpdateSyncPolicy(context.Background())
	})
}

func TestDeleteSyncPolicy(t *testing.T) {
	t.Run("VerifyNoOp", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		}))
		defer server.Close()

		client := cyberark.NewSecretsHubAPI(server.URL, []byte("dummy_token"))

		client.DeleteSyncPolicy(context.Background())
	})
}
