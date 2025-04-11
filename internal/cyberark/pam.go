// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Account is an interface for interacting with SecretsHub's accounts.
type Account interface {
	AddAccount(ctx context.Context, credential Credential) (*CredentialResponse, error)
	GetAccount(ctx context.Context, accountID string) (*CredentialResponse, error)
	FilterAccounts(ctx context.Context, search string, filter []string) (*CredentialSearchResponse, error)
	UpdateAccount(ctx context.Context, accountID string, credential Credential) (*CredentialResponse, error)
	DeleteAccount(ctx context.Context, accountID string) error
}

// Safe is an interface for interacting with SecretsHub's safes.
type Safe interface {
	AddSafe(ctx context.Context, safe SafeData) (*SafeData, error)
	GetSafe(ctx context.Context, safeID string) (*SafeData, error)
	UpdateSafe(ctx context.Context, safeID string, safe SafeData) (*SafeData, error)
	DeleteSafe(ctx context.Context, safeID string) error
}

// SafeMember is an interface for interacting with SecretsHub's safe members.
type SafeMember interface {
	AddSafeMember(ctx context.Context, safe SafeData) (*Member, error)
	GetSafeMember(ctx context.Context, safe SafeData) (*Member, error)
	UpdateSafeMember(ctx context.Context, safe SafeData) (*Member, error)
	DeleteSafeMember(ctx context.Context, safeName string, memberName string) error
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

// UpdateAccount updates an account in the SecretsHub using JSON Patch.
func (a *pamAPI) UpdateAccount(ctx context.Context, accountID string, credential Credential) (*CredentialResponse, error) {
	// Fetch the existing account to compare with the desired state
	existingAccount, err := a.GetAccount(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing account for update: %w", err)
	}

	// Generate the JSON Patch
	patch, err := generateAccountPatch(existingAccount, &credential)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JSON patch: %w", err)
	}

	// If there are no changes, return early
	if len(patch) == 0 {
		tflog.Info(ctx, "No changes detected, skipping account update")
		return existingAccount, nil
	}

	body, err := json.Marshal(patch)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON patch: %w", err)
	}

	response, err := a.client.DoRequest(
		ctx,
		"PATCH",
		fmt.Sprintf("/PasswordVault/API/Accounts/%s", accountID),
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		tflog.Error(ctx, fmt.Sprintf("failed to update account, got response: %s", response.Body))
		return nil, fmt.Errorf("failed to update account. Expected status code 200, got %d", response.StatusCode)
	}

	updatedAccount := CredentialResponse{}
	err = json.NewDecoder(response.Body).Decode(&updatedAccount)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully updated account [%s]: Name [%s] - ID [%v]",
		*updatedAccount.UserName, *updatedAccount.Name, *updatedAccount.CredID))

	return &updatedAccount, nil
}

// generateAccountPatch generates a JSON Patch for updating an account.
func generateAccountPatch(existing *CredentialResponse, desired *Credential) ([]map[string]interface{}, error) {
	patch := []map[string]interface{}{}

	if existing == nil || desired == nil {
		return patch, fmt.Errorf("existing and desired accounts must not be nil")
	}

	// Basic account properties
	if desired.Name != nil && existing.Name != nil && *existing.Name != *desired.Name {
		patch = append(patch, map[string]interface{}{
			"op":    "replace",
			"path":  "/name",
			"value": *desired.Name,
		})
	}

	if desired.Address != nil && existing.Address != nil && *existing.Address != *desired.Address {
		patch = append(patch, map[string]interface{}{
			"op":    "replace",
			"path":  "/address",
			"value": *desired.Address,
		})
	}

	if desired.UserName != nil && existing.UserName != nil && *existing.UserName != *desired.UserName {
		patch = append(patch, map[string]interface{}{
			"op":    "replace",
			"path":  "/userName",
			"value": *desired.UserName,
		})
	}

	if desired.Platform != nil && existing.Platform != nil && *existing.Platform != *desired.Platform {
		patch = append(patch, map[string]interface{}{
			"op":    "replace",
			"path":  "/platformId",
			"value": *desired.Platform,
		})
	}

	if desired.Props != nil {
		if existing.Props == nil {
			// If existing has no properties but desired does, add them all
			patch = append(patch, map[string]interface{}{
				"op":    "add",
				"path":  "/platformAccountProperties",
				"value": desired.Props,
			})
		} else if !reflect.DeepEqual(existing.Props, desired.Props) {
			// Only update if there are actual differences
			patch = append(patch, map[string]interface{}{
				"op":    "replace",
				"path":  "/platformAccountProperties",
				"value": desired.Props,
			})
		}
	}

	// Secret management properties
	if desired.SecretMgmt != nil {
		// Handle automaticManagementEnabled
		if desired.SecretMgmt.AutomaticManagement != nil {
			automaticManagementChanged := false

			if existing.SecretMgmt == nil || existing.SecretMgmt.AutomaticManagement == nil {
				automaticManagementChanged = true
			} else if *existing.SecretMgmt.AutomaticManagement != *desired.SecretMgmt.AutomaticManagement {
				automaticManagementChanged = true
			}

			if automaticManagementChanged {
				patch = append(patch, map[string]interface{}{
					"op":    "replace",
					"path":  "/secretManagement/automaticManagementEnabled",
					"value": *desired.SecretMgmt.AutomaticManagement,
				})
			}
		}

		// Handle manualManagementReason
		if desired.SecretMgmt.ManualManagementReason != nil {
			reasonChanged := false

			if existing.SecretMgmt == nil || existing.SecretMgmt.ManualManagementReason == nil {
				reasonChanged = true
			} else if *existing.SecretMgmt.ManualManagementReason != *desired.SecretMgmt.ManualManagementReason {
				reasonChanged = true
			}

			if reasonChanged {
				patch = append(patch, map[string]interface{}{
					"op":    "replace",
					"path":  "/secretManagement/manualManagementReason",
					"value": *desired.SecretMgmt.ManualManagementReason,
				})
			}
		}
	}

	return patch, nil
}

// DeleteAccount deletes an account from the SecretsHub.
func (a *pamAPI) DeleteAccount(ctx context.Context, accountID string) error {
	response, err := a.client.DoRequest(
		ctx,
		"DELETE",
		fmt.Sprintf("/PasswordVault/API/Accounts/%s", accountID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		tflog.Error(ctx, fmt.Sprintf("failed to delete account, got response: %s", response.Body))
		return fmt.Errorf("failed to delete account. Expected status code 204, got %d", response.StatusCode)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully deleted account with ID [%s]", accountID))
	return nil
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
func (a *pamAPI) UpdateSafe(ctx context.Context, safeID string, safe SafeData) (*SafeData, error) {
	body, err := json.Marshal(safe)
	if err != nil {
		return nil, err
	}

	response, err := a.client.DoRequest(
		ctx,
		"PUT",
		fmt.Sprintf("/PasswordVault/API/Safes/%s", safeID),
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		tflog.Error(ctx, fmt.Sprintf("failed to update safe, got response: %s", response.Body))
		return nil, fmt.Errorf("failed to update safe. Expected status code 200, got %d", response.StatusCode)
	}

	updatedSafe := SafeData{}
	err = json.NewDecoder(response.Body).Decode(&updatedSafe)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully updated safe [%s]: Name [%s] - ID [%v]",
		*updatedSafe.Name, *updatedSafe.URLID, *updatedSafe.NUMBER))

	return &updatedSafe, nil
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
		tflog.Error(ctx, fmt.Sprintf("failed to delete safe, got response: %s", response.Body))
		return fmt.Errorf("failed to delete safe. Expected status code 204, got %d", response.StatusCode)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully deleted safe with ID [%s]", safeID))
	return nil
}

// AddSafeMember adds a new member to a safe in the SecretsHub.
func (a *pamAPI) AddSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	tflog.Debug(ctx, fmt.Sprintf("Generating Permission %s.", *safe.Level))
	tflog.Debug(ctx, fmt.Sprintf("Ownership Properties: %s, %s, %s", *safe.Owner, *safe.OwnerType, *safe.Level))

	block, err := generateSafePermissions(&safe)
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
		return nil, err
	}

	if response.StatusCode == 409 {
		tflog.Info(ctx, fmt.Sprintf("Safe [%s] already has member [%s].", *safe.Name, *safe.Owner))
		return nil, nil
	} else if response.StatusCode != 201 {
		return nil, fmt.Errorf("failed to add safe member, expected status code 201, got %d", response.StatusCode)
	}

	safeMember := Member{}
	err = json.NewDecoder(response.Body).Decode(&safeMember)
	if err != nil {
		return nil, err
	}

	return &safeMember, nil
}

// GetSafeMember retrieves a safe member
func (a *pamAPI) GetSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", *safe.Name, *safe.Owner),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, fmt.Errorf("failed to get safe member. Expected status code 200, got %d", response.StatusCode)
	}

	safeMember := Member{}
	err = json.NewDecoder(response.Body).Decode(&safeMember)
	if err != nil {
		return nil, err
	}

	return &safeMember, nil
}

// UpdateSafeMember updates a safe member
func (a *pamAPI) UpdateSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	tflog.Debug(ctx, fmt.Sprintf("Updating permission for member %s to level %s.", *safe.Owner, *safe.Level))
	tflog.Debug(ctx, fmt.Sprintf("Ownership Properties: %s, %s, %s", *safe.Owner, *safe.OwnerType, *safe.Level))

	block, err := generateSafePermissions(&safe)
	if err != nil {
		tflog.Error(ctx, "Error generating permissions block.")
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Generated updated permission block for: %s", *safe.Owner))
	tflog.Debug(ctx, string(block))

	response, err := a.client.DoRequest(
		ctx,
		"PUT",
		fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", *safe.Name, *safe.Owner),
		bytes.NewBuffer(block),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	updatedSafeMember := Member{}
	err = json.NewDecoder(response.Body).Decode(&updatedSafeMember)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		tflog.Error(ctx, fmt.Sprintf("failed to update safe member, got response: %s", response.Body))
		return nil, fmt.Errorf("failed to update safe member, expected status code 200, got %d", response.StatusCode)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully updated member [%s] permissions in safe [%s]", *safe.Owner, *safe.Name))
	return &updatedSafeMember, nil
}

// DeleteSafeMember deletes a safe member from the SecretsHub.
func (a *pamAPI) DeleteSafeMember(ctx context.Context, safeName string, memberName string) error {
	tflog.Debug(ctx, fmt.Sprintf("Attempting to delete member [%s] from safe [%s]", memberName, safeName))

	response, err := a.client.DoRequest(
		ctx,
		"DELETE",
		fmt.Sprintf("/PasswordVault/API/Safes/%s/Members/%s", safeName, memberName),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode == 404 {
		tflog.Info(ctx, fmt.Sprintf("Member [%s] not found in safe [%s]", memberName, safeName))
		return nil
	} else if response.StatusCode != 204 {
		return fmt.Errorf("failed to delete safe member, expected status code 204, got %d", response.StatusCode)
	}

	tflog.Info(ctx, fmt.Sprintf("Successfully removed member [%s] from safe [%s]", memberName, safeName))
	return nil
}

func generateSafePermissions(safe *SafeData) ([]byte, error) {
	switch *safe.Level {
	case "full":
		return FullAdmin(safe.OwnerType, safe.Owner)
	case "read":
		return ReadOnly(safe.OwnerType, safe.Owner)
	case "approver":
		return Approver(safe.OwnerType, safe.Owner)
	case "manager":
		return Manager(safe.OwnerType, safe.Owner)
	}

	return []byte{}, errors.New("invalid permission level")
}

// NewPAMAPI creates a new PAMAPI client.
func NewPAMAPI(baseURL string, authToken []byte, withBearerToken bool) PAMAPI {
	return &pamAPI{
		client:    NewClientWithToken(baseURL, true, authToken, withBearerToken),
		authToken: authToken,
	}
}
