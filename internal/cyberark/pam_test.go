package cyberark_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/stretchr/testify/assert"
)

var (
	token               = []byte("dummy_token")
	credID              = "123"
	name                = "user"
	safe                = "user_safe"
	owner               = "test_owner"
	ownerType           = "test_user"
	levelFull           = "full"
	levelRead           = "read"
	levelApprover       = "approver"
	levelManager        = "manager"
	number        int64 = 1234
	count               = 1
)

func TestAddAccount(t *testing.T) {
	t.Run("AddAccount", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			body, _ := io.ReadAll(req.Body)
			req.Body.Close()
			assert.Contains(t, string(body), `"name":"user"`)

			rw.WriteHeader(http.StatusCreated)
			resp := cyberark.CredentialResponse{
				CredID:   &credID,
				UserName: &name,
				Name:     &name,
			}

			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.AddAccount(context.Background(), credentials)

		expectedData := cyberark.CredentialResponse{
			CredID:   &credID,
			UserName: &name,
			Name:     &name,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("AccountAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Account already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.AddAccount(context.Background(), credentials)

		assert.Empty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.AddAccount(context.Background(), credentials)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.AddAccount(context.Background(), credentials)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetAccount(t *testing.T) {
	t.Run("GetAccount", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			resp := cyberark.CredentialResponse{
				CredID: &credID,
				Name:   &name,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()
		expectedData := cyberark.CredentialResponse{
			CredID: &credID,
			Name:   &name,
		}
		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetAccount(context.Background(), "test_account")

		assert.NoError(t, err)

		assert.Equal(t, expectedData, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetAccount(context.Background(), "test_account")

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetAccount(context.Background(), "test_account")

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestFilterAccounts_SearchAndFilter(t *testing.T) {
	t.Run("SearchAndFilter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("filter=safeName+eq+%s&search=%s", safe, name), req.URL.RawQuery)
			resp := cyberark.CredentialSearchResponse{
				Accounts: []*cyberark.CredentialResponse{
					{
						CredID: &credID,
						Name:   &name,
					},
				},
				Count: &count,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		expectedData := cyberark.CredentialSearchResponse{
			Accounts: []*cyberark.CredentialResponse{
				{
					CredID: &credID,
					Name:   &name,
				},
			},
			Count: &count,
		}
		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.FilterAccounts(context.Background(), name, []string{fmt.Sprintf("safeName eq %s", safe)})

		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("Search", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("search=%s", name), req.URL.RawQuery)
			resp := cyberark.CredentialSearchResponse{
				Accounts: []*cyberark.CredentialResponse{
					{
						CredID: &credID,
						Name:   &name,
					},
				},
				Count: &count,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		expectedData := cyberark.CredentialSearchResponse{
			Accounts: []*cyberark.CredentialResponse{
				{
					CredID: &credID,
					Name:   &name,
				},
			},
			Count: &count,
		}
		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.FilterAccounts(context.Background(), name, nil)

		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("Filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, fmt.Sprintf("filter=safeName+eq+%s", safe), req.URL.RawQuery)
			resp := cyberark.CredentialSearchResponse{
				Accounts: []*cyberark.CredentialResponse{
					{
						CredID: &credID,
						Name:   &name,
					},
				},
				Count: &count,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		expectedData := cyberark.CredentialSearchResponse{
			Accounts: []*cyberark.CredentialResponse{
				{
					CredID: &credID,
					Name:   &name,
				},
			},
			Count: &count,
		}
		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.FilterAccounts(context.Background(), "", []string{fmt.Sprintf("safeName eq %s", safe)})

		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.FilterAccounts(context.Background(), "", nil)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestUpdateAccount(t *testing.T) {
	t.Run("UpdateAccount", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "PUT", req.Method)
			assert.Contains(t, req.URL.Path, credID)

			body, _ := io.ReadAll(req.Body)
			req.Body.Close()
			assert.Contains(t, string(body), `"name":"user"`)

			rw.WriteHeader(http.StatusOK)
			resp := cyberark.CredentialResponse{
				CredID:   &credID,
				UserName: &name,
				Name:     &name,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.UpdateAccount(context.Background(), credID, credentials)

		expectedData := cyberark.CredentialResponse{
			CredID:   &credID,
			UserName: &name,
			Name:     &name,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.UpdateAccount(context.Background(), credID, credentials)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		credentials := cyberark.Credential{
			Name: &name,
		}

		resp, err := client.UpdateAccount(context.Background(), credID, credentials)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestDeleteAccount(t *testing.T) {
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Accounts/%s", credID), req.URL.Path)

			rw.WriteHeader(http.StatusNoContent) // 204 No Content
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteAccount(context.Background(), credID)

		assert.NoError(t, err)
	})

	t.Run("AccountNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Accounts/%s", credID), req.URL.Path)

			rw.WriteHeader(http.StatusNotFound) // 404 Not Found
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteAccount(context.Background(), credID)

		assert.Error(t, err) // Should return error for 404 response
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteAccount(context.Background(), credID)

		assert.Error(t, err)
	})
}

func TestAddSafe(t *testing.T) {
	t.Run("AddSafe", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusCreated)
			resp := cyberark.SafeData{
				Name:   &safe,
				URLID:  &safe,
				NUMBER: &number,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name:   &safe,
			URLID:  &safe,
			NUMBER: &number,
		}

		resp, err := client.AddSafe(context.Background(), safe)

		assert.NoError(t, err)
		assert.Equal(t, safe, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.AddSafe(context.Background(), safe)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("SafeAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Safe already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.AddSafe(context.Background(), safe)

		assert.Empty(t, resp)
		assert.NoError(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.AddSafe(context.Background(), safe)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestGetSafe(t *testing.T) {
	t.Run("GetSafe", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Contains(t, req.URL.Path, safe)
			resp := cyberark.SafeData{
				Name: &safe,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		expectedData := cyberark.SafeData{
			Name: &safe,
		}

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetSafe(context.Background(), safe)

		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetSafe(context.Background(), "test_safe_id")

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		resp, err := client.GetSafe(context.Background(), "test_safe_id")

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestUpdateSafe(t *testing.T) {
	t.Run("UpdateSafe", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "PUT", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Safes/%s", safe), req.URL.Path)

			body, _ := io.ReadAll(req.Body)
			req.Body.Close()
			assert.Contains(t, string(body), `"safeName":"user_safe"`)

			rw.WriteHeader(http.StatusOK)
			resp := cyberark.SafeData{
				Name:   &safe,
				URLID:  &safe,
				NUMBER: &number,
			}
			json.NewEncoder(rw).Encode(resp)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safeData := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.UpdateSafe(context.Background(), safe, safeData)

		expectedData := cyberark.SafeData{
			Name:   &safe,
			URLID:  &safe,
			NUMBER: &number,
		}
		assert.NoError(t, err)
		assert.Equal(t, expectedData, *resp)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safeData := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.UpdateSafe(context.Background(), safe, safeData)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})

	t.Run("InvalidJSONResponse", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			rw.Write([]byte(`invalid json`))
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safeData := cyberark.SafeData{
			Name: &safe,
		}

		resp, err := client.UpdateSafe(context.Background(), safe, safeData)

		assert.Empty(t, resp)
		assert.Error(t, err)
	})
}

func TestDeleteSafe(t *testing.T) {
	t.Run("SuccessfulDeletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Safes/%s", safe), req.URL.Path)

			rw.WriteHeader(http.StatusNoContent) // 204 No Content
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteSafe(context.Background(), safe)

		assert.NoError(t, err)
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteSafe(context.Background(), safe)

		assert.Error(t, err)
	})
}

func TestAddSafeMember(t *testing.T) {
	testCases := []struct {
		name  string
		level string
	}{
		{
			name:  "LevelFull",
			level: levelFull,
		},
		{
			name:  "LevelRead",
			level: levelRead,
		},
		{
			name:  "LevelApprover",
			level: levelApprover,
		},
		{
			name:  "LevelManager",
			level: levelManager,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				body, _ := io.ReadAll(req.Body)
				req.Body.Close()

				assert.Contains(t, req.URL.Path, name)
				assert.NotEmpty(t, body)

				rw.WriteHeader(http.StatusCreated)
			}))
			defer server.Close()

			client := cyberark.NewPAMAPI(server.URL, token, true)

			safe := cyberark.SafeData{
				Name:      &name,
				Owner:     &owner,
				OwnerType: &ownerType,
				Level:     &tc.level,
			}

			err := client.AddSafeMember(context.Background(), safe)

			assert.NoError(t, err)
			assert.Equal(t, nil, err)
		})
	}

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name:      &name,
			Owner:     &owner,
			OwnerType: &ownerType,
			Level:     &levelManager,
		}

		err := client.AddSafeMember(context.Background(), safe)

		assert.Error(t, err)
	})

	t.Run("MemberAlreadyExists", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Member already exists", http.StatusConflict)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name:      &name,
			Owner:     &owner,
			OwnerType: &ownerType,
			Level:     &levelManager,
		}

		err := client.AddSafeMember(context.Background(), safe)

		assert.NoError(t, err)
	})
}

func TestGetSafeMember(t *testing.T) {
	// Since GetSafeMember is a placeholder function that doesn't do anything,
	// we'll keep this as a basic verification but add a proper comment
	t.Run("Placeholder Function", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, req *http.Request) {
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		client.GetSafeMember(context.Background())
		// No assertions as the function currently does nothing
	})
}

func TestUpdateSafeMember(t *testing.T) {
	name := "TestSafe"
	owner := "testUser"
	ownerType := "user"

	// Test different permission levels
	testCases := []struct {
		name  string
		level string
	}{
		{"FullAdminPermissions", levelFull},
		{"ReadOnlyPermissions", levelRead},
		{"ApproverPermissions", levelApprover},
		{"ManagerPermissions", levelManager},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
				assert.Equal(t, "PUT", req.Method)
				assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", name, owner), req.URL.Path)

				rw.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client := cyberark.NewPAMAPI(server.URL, token, true)

			safe := cyberark.SafeData{
				Name:      &name,
				Owner:     &owner,
				OwnerType: &ownerType,
				Level:     &tc.level,
			}

			err := client.UpdateSafeMember(context.Background(), safe)

			assert.NoError(t, err)
		})
	}

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)

		safe := cyberark.SafeData{
			Name:      &name,
			Owner:     &owner,
			OwnerType: &ownerType,
			Level:     &levelManager,
		}

		err := client.UpdateSafeMember(context.Background(), safe)

		assert.Error(t, err)
	})
}

func TestDeleteSafeMember(t *testing.T) {
	safeName := "TestSafe"
	memberName := "testUser"

	t.Run("SuccessfulDeletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", safeName, memberName), req.URL.Path)

			rw.WriteHeader(http.StatusNoContent) // 204 No Content
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteSafeMember(context.Background(), safeName, memberName)

		assert.NoError(t, err)
	})

	t.Run("MemberNotFound", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			assert.Equal(t, "DELETE", req.Method)
			assert.Equal(t, fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", safeName, memberName), req.URL.Path)

			rw.WriteHeader(http.StatusNotFound) // 404 Not Found
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteSafeMember(context.Background(), safeName, memberName)

		assert.NoError(t, err) // Should return nil for 404 response
	})

	t.Run("ErrorStatusCode", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, _ *http.Request) {
			http.Error(rw, "Internal Server Error", http.StatusInternalServerError)
		}))
		defer server.Close()

		client := cyberark.NewPAMAPI(server.URL, token, true)
		err := client.DeleteSafeMember(context.Background(), safeName, memberName)

		assert.Error(t, err)
	})
}
