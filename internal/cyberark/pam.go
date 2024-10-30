// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Account is an interface for interacting with SecretsHub's accounts.
type Account interface {
	AddAccount(ctx context.Context, credential Credential) (*CredentialResponse, error)
	GetAccount(ctx context.Context, accountID string) (*CredentialResponse, error)
	FilterAccounts(ctx context.Context, search string, filter []string) (*CredentialSearchResponse, error)
	UpdateAccount(ctx context.Context)
	DeleteAccount(ctx context.Context)
}

// Safe is an interface for interacting with SecretsHub's safes.
type Safe interface {
	AddSafe(ctx context.Context, safe SafeData) (*SafeData, error)
	GetSafe(ctx context.Context, safeID string) (*SafeData, error)
	UpdateSafe(ctx context.Context)
	DeleteSafe(ctx context.Context, safeID string) error
}

// SafeMember is an interface for interacting with SecretsHub's safe members.
type SafeMember interface {
	AddSafeMember(ctx context.Context, safe SafeData) error
	GetSafeMember(ctx context.Context)
	UpdateSafeMember(ctx context.Context)
	DeleteSafeMember(ctx context.Context)
}

// PAMAPI is an interface for interacting with the PAM APIs.
type PAMAPI interface {
	Account
	Safe
	SafeMember
}

// pamAPI is a client for interacting with the SecretsHub APIs.
type pamAPI struct {
	client    *Client
	authToken []byte
}

// AddAccount adds a new account to the SecretsHub.
func (a *pamAPI) AddAccount(ctx context.Context, credential Credential) (*CredentialResponse, error) {
	body, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		"/PasswordVault/API/Accounts",
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 409 {
		tflog.Info(ctx, fmt.Sprintf("Account [%s] already exists.", *credential.Name))
		return nil, nil
	} else if response.StatusCode != 201 {
		return nil, fmt.Errorf("failed to add account, expected status code 201, got %d", response.StatusCode)
	}

	createdAccount := CredentialResponse{}
	err = json.NewDecoder(response.Body).Decode(&createdAccount)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully added new account [%s]: Name [%s] - ID [%v]",
		*createdAccount.UserName, *createdAccount.Name, *createdAccount.CredID))

	return &createdAccount, nil
}

// GetAccount retrieves an account from the SecretsHub.
func (a *pamAPI) GetAccount(ctx context.Context, accountID string) (*CredentialResponse, error) {
	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/PasswordVault/API/Accounts/%s", accountID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get account. Expected status code 200, got %d", response.StatusCode)
	}

	account := CredentialResponse{}
	err = json.NewDecoder(response.Body).Decode(&account)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

// FilterAccounts searches for accounts in the SecretsHub.
func (a *pamAPI) FilterAccounts(ctx context.Context, search string, filter []string) (*CredentialSearchResponse, error) {
	params := a.filters(search, filter)

	response, err := a.client.DoRequest(
		ctx,
		"GET",
		"/PasswordVault/api/accounts",
		nil,
		map[string]string{},
		params,
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to filter accounts. Expected status code 200, got %d", response.StatusCode)
	}

	searchAccounts := CredentialSearchResponse{}
	err = json.NewDecoder(response.Body).Decode(&searchAccounts)
	if err != nil {
		return nil, err
	}

	return &searchAccounts, nil
}

func (a *pamAPI) filters(search string, filter []string) (query map[string]string) {
	query = make(map[string]string)

	if len(filter) > 0 {
		query["filter"] = strings.Join(filter, " AND ")
	}
	if len(search) > 0 {
		query["search"] = search
	}
	if len(query) == 0 {
		return map[string]string{}
	}
	return query
}

// UpdateAccount updates an account in the SecretsHub.
func (a *pamAPI) UpdateAccount(_ context.Context) {
}

// DeleteAccount deletes an account from the SecretsHub.
func (a *pamAPI) DeleteAccount(_ context.Context) {
}

// AddSafe adds a new safe to the SecretsHub.
func (a *pamAPI) AddSafe(ctx context.Context, safe SafeData) (*SafeData, error) {
	body, err := json.Marshal(safe)
	if err != nil {
		return nil, err
	}

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		"/PasswordVault/API/Safes",
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode == 409 {
		tflog.Info(ctx, fmt.Sprintf("Safe [%s] already exists.", *safe.Name))
		return nil, nil
	} else if response.StatusCode != 201 {
		return nil, fmt.Errorf("failed to add safe, expected status code 201, got %d", response.StatusCode)
	}

	savedSafe := SafeData{}
	err = json.NewDecoder(response.Body).Decode(&savedSafe)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully added new safe [%s]: Name [%s] - ID [%v]",
		*savedSafe.Name, *savedSafe.URLID, *savedSafe.NUMBER))

	return &savedSafe, nil
}

// GetSafe retrieves a safe from the SecretsHub.
func (a *pamAPI) GetSafe(ctx context.Context, safeID string) (*SafeData, error) {
	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/PasswordVault/API/Safes/%s", safeID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get safe. Expected status code 200, got %d", response.StatusCode)
	}

	safe := SafeData{}
	err = json.NewDecoder(response.Body).Decode(&safe)
	if err != nil {
		return nil, err
	}

	return &safe, nil
}

// UpdateSafe updates a safe in the SecretsHub.
func (a *pamAPI) UpdateSafe(_ context.Context) {
}

// DeleteSafe deletes a safe from the SecretsHub.
func (a *pamAPI) DeleteSafe(ctx context.Context, safeID string) error {
	response, err := a.client.DoRequest(
		ctx,
		"DELETE",
		fmt.Sprintf("/PasswordVault/API/Safes/%s", safeID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return fmt.Errorf("failed to delete safe. Expected status code 204, got %d", response.StatusCode)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully deleted safe [%s]", safeID))
	return nil
}

// AddSafeMember adds a new member to a safe in the SecretsHub.
func (a *pamAPI) AddSafeMember(ctx context.Context, safe SafeData) error {
	tflog.Debug(ctx, fmt.Sprintf("Generating Permission %s.", *safe.Level))
	tflog.Debug(ctx, fmt.Sprintf("Ownership Properties: %s, %s, %s", *safe.Owner, *safe.OwnerType, *safe.Level))

	var block []byte
	var err error

	switch *safe.Level {
	case "full":
		block, err = FullAdmin(safe.OwnerType, safe.Owner)
	case "read":
		block, err = ReadOnly(safe.OwnerType, safe.Owner)
	case "approver":
		block, err = Approver(safe.OwnerType, safe.Owner)
	case "manager":
		block, err = Manager(safe.OwnerType, safe.Owner)
	}
	if err != nil {
		tflog.Error(ctx, "Error generating permissions block.")
	}

	tflog.Info(ctx, fmt.Sprintf("Generated permission block for: %s", *safe.Owner))
	tflog.Debug(ctx, string(block))

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		fmt.Sprintf("/PasswordVault/API/Safes/%s/Members", *safe.Name),
		bytes.NewBuffer(block),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode == 409 {
		tflog.Info(ctx, fmt.Sprintf("Safe [%s] already has member [%s].", *safe.Name, *safe.Owner))
		return nil
	} else if response.StatusCode != 201 {
		return fmt.Errorf("failed to add safe member, expected status code 201, got %d", response.StatusCode)
	}

	return nil
}

// GetSafeMember retrieves a safe member from the SecretsHub.
func (a *pamAPI) GetSafeMember(_ context.Context) {
}

// UpdateSafeMember updates a safe member in the SecretsHub.
func (a *pamAPI) UpdateSafeMember(_ context.Context) {
}

// DeleteSafeMember deletes a safe member from the SecretsHub.
func (a *pamAPI) DeleteSafeMember(_ context.Context) {
}

// NewPAMAPI creates a new PAMAPI client.
func NewPAMAPI(baseURL string, authToken []byte, withBearerToken bool) PAMAPI {
	return &pamAPI{
		client:    NewClientWithToken(baseURL, true, authToken, withBearerToken),
		authToken: authToken,
	}
}
