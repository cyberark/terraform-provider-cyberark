// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// SecretStore is an interface for interacting with SecretsHub's secret stores.
type SecretStore interface {
	AddAwsAsmSecretStore(ctx context.Context, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error)
	AddAzureAkvSecretStore(ctx context.Context, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error)
	GetAwsAsmSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AwsAsmData], error)
	GetAzureAkvSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AzureAkvData], error)
	GetAwsAsmSecretStores(ctx context.Context) (*SecretStoresOutput[AwsAsmData], error)
	GetAzureAkvSecretStores(ctx context.Context) (*SecretStoresOutput[AzureAkvData], error)
	UpdateAwsSecretStore(ctx context.Context, storeID string, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error)
	UpdateAzureAkvSecretStore(ctx context.Context, storeID string, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error)
	DeleteSecretStore(ctx context.Context, storeID string) error
	SetSecretStoreState(ctx context.Context, storeID string, action string) error
}

// SyncPolicy is an interface for interacting with SecretsHub's sync policies.
type SyncPolicy interface {
	AddSyncPolicy(ctx context.Context, pi PolicyInput) (*PolicyExternalOutput, error)
	GetSyncPolicy(ctx context.Context, policyID string) (*PolicyExternalOutput, error)
	GetSyncPolicies(ctx context.Context) (*SyncResponse, error)
	GetSecretFilter(ctx context.Context, storeID string, filterID string) (*SecretFilterOutput, error)
	DeleteSyncPolicy(ctx context.Context, policyID string) error
	UpdateSyncPolicy(ctx context.Context, policyID string, pi PolicyInput) (*PolicyExternalOutput, error)
}

// SecretsHubAPI is an interface for interacting with the SecretsHub APIs.
type SecretsHubAPI interface {
	SecretStore
	SyncPolicy
}

// secretsHubAPI is a client for interacting with the SecretsHub APIs.
type secretsHubAPI struct {
	client    *Client
	authToken []byte
}

// AddAwsAsmSecretStore adds a new AWS ASM secret store to the SecretsHub.
func (a *secretsHubAPI) AddAwsAsmSecretStore(ctx context.Context, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error) {
	var output SecretStoreOutput[AwsAsmData]
	err := a.addSecretStore(ctx, body, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// AddAzureAkvSecretStore adds a new Azure AKS secret store to the SecretsHub.
func (a *secretsHubAPI) AddAzureAkvSecretStore(ctx context.Context, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error) {
	var output SecretStoreOutput[AzureAkvData]
	err := a.addSecretStore(ctx, body, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (a *secretsHubAPI) addSecretStore(ctx context.Context, body interface{}, output interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		"/api/secret-stores",
		bytes.NewBuffer(bodyBytes),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 201 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return err
	}

	return nil
}

// GetAwsAsmSecretStore retrieves a secret store from the AWS ASM SecretsHub.
func (a *secretsHubAPI) GetAwsAsmSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AwsAsmData], error) {
	var output SecretStoreOutput[AwsAsmData]
	err := a.getSecretStore(ctx, storeID, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// GetAzureAkvSecretStore retrieves a secret store from the Azure AKV SecretsHub.
func (a *secretsHubAPI) GetAzureAkvSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AzureAkvData], error) {
	var output SecretStoreOutput[AzureAkvData]
	err := a.getSecretStore(ctx, storeID, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (a *secretsHubAPI) getSecretStore(ctx context.Context, storeID string, output interface{}) error {
	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/api/secret-stores/%s", storeID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return err
	}
	return nil
}

// GetAwsAsmSecretStores retrieves all AWS ASM secret stores from the SecretsHub.
func (a *secretsHubAPI) GetAwsAsmSecretStores(ctx context.Context) (*SecretStoresOutput[AwsAsmData], error) {
	var output SecretStoresOutput[AwsAsmData]

	err := a.getSecretStores(ctx, "AWS_ASM", &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// GetAzureAkvSecretStores retrieves all Azure AKS secret stores from the SecretsHub.
func (a *secretsHubAPI) GetAzureAkvSecretStores(ctx context.Context) (*SecretStoresOutput[AzureAkvData], error) {
	var output SecretStoresOutput[AzureAkvData]

	err := a.getSecretStores(ctx, "AZURE_AKV", &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (a *secretsHubAPI) getSecretStores(ctx context.Context, storeType string, output interface{}) error {
	params := map[string]string{
		"filter": fmt.Sprintf("type EQ %s", storeType),
	}

	response, err := a.client.DoRequest(
		ctx,
		"GET",
		"/api/secret-stores",
		nil,
		map[string]string{},
		params,
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return err
	}

	return nil
}

// UpdateSecretStore updates a secret store in the SecretsHub.
func (a *secretsHubAPI) UpdateAwsSecretStore(ctx context.Context, storeId string, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error) {
	var output SecretStoreOutput[AwsAsmData]

	err := a.updateSecretStore(ctx, storeId, body, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// UpdateAzureAkvSecretStore updates an Azure AKV secret store in the SecretsHub.
func (a *secretsHubAPI) UpdateAzureAkvSecretStore(ctx context.Context, storeId string, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error) {
	var output SecretStoreOutput[AzureAkvData]

	err := a.updateSecretStore(ctx, storeId, body, &output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

func (a *secretsHubAPI) updateSecretStore(ctx context.Context, storeId string, body interface{}, output interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	response, err := a.client.DoRequest(
		ctx,
		"PATCH",
		fmt.Sprintf("/api/secret-stores/%s", storeId),
		bytes.NewBuffer(bodyBytes),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return err
	}

	return nil
}

// DeleteSecretStore deletes a secret store from the SecretsHub.
func (a *secretsHubAPI) DeleteSecretStore(ctx context.Context, storeId string) error {
	response, err := a.client.DoRequest(
		ctx,
		"DELETE",
		fmt.Sprintf("/api/secret-stores/%s", storeId),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	return nil
}

// SetSecretStoreState sets the state of a secret store in the SecretsHub.
func (a *secretsHubAPI) SetSecretStoreState(ctx context.Context, storeID string, action string) error {
	body, err := json.Marshal(map[string]string{"action": action})
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	response, err := a.client.DoRequest(
		ctx,
		"PUT",
		fmt.Sprintf("/api/secret-stores/%s/state", storeID),
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 204 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	return nil
}

// AddSyncPolicy adds a new sync policy to the SecretsHub.
func (a *secretsHubAPI) AddSyncPolicy(ctx context.Context, pi PolicyInput) (*PolicyExternalOutput, error) {
	body, err := json.Marshal(pi)
	if err != nil {
		return nil, err
	}

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		"/api/policies",
		bytes.NewBuffer(body),
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 201 {
		return nil, APIErrorFromResponse(response.StatusCode, response.Body)
	}

	output := PolicyExternalOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// GetSyncPolicy retrieves a sync policy from SecretsHub.
func (a *secretsHubAPI) GetSyncPolicy(ctx context.Context, policyID string) (*PolicyExternalOutput, error) {
	params := map[string]string{
		"projection": "REGULAR",
	}

	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/api/policies/%s", policyID),
		nil,
		map[string]string{},
		params,
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, APIErrorFromResponse(response.StatusCode, response.Body)
	}

	output := PolicyExternalOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// GetSyncPolicies retrieves all sync policies from the SecretsHub.
func (a *secretsHubAPI) GetSyncPolicies(ctx context.Context) (*SyncResponse, error) {
	params := map[string]string{
		"projection": "REGULAR",
	}

	response, err := a.client.DoRequest(
		ctx,
		"GET",
		"/api/policies",
		nil,
		map[string]string{},
		params,
	)
	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, APIErrorFromResponse(response.StatusCode, response.Body)
	}

	output := SyncResponse{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// UpdateSyncPolicy updates a sync policy by deleting the existing one and creating a new one.
func (a *secretsHubAPI) UpdateSyncPolicy(ctx context.Context, policyID string, pi PolicyInput) (*PolicyExternalOutput, error) {
	// First delete the existing policy
	err := a.DeleteSyncPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing policy during update: %w", err)
	}

	// Then create a new policy with the updated parameters
	output, err := a.AddSyncPolicy(ctx, pi)
	if err != nil {
		return nil, fmt.Errorf("failed to create new policy during update: %w", err)
	}

	return output, nil
}

// DeleteSyncPolicy deletes a sync policy from the SecretsHub.
func (a *secretsHubAPI) DeleteSyncPolicy(ctx context.Context, policyID string) error {
	// First disable the policy
	disableBody, err := json.Marshal(map[string]string{"action": "disable"})
	if err != nil {
		return fmt.Errorf("failed to marshal disable policy request: %w", err)
	}

	disableResponse, err := a.client.DoRequest(
		ctx,
		"PUT",
		fmt.Sprintf("/api/policies/%s/state", policyID),
		bytes.NewBuffer(disableBody),
		map[string]string{},
		map[string]string{},
	)

	if err != nil {
		tflog.Warn(ctx, fmt.Sprintf("Failed to disable policy before deletion: %v", err))
		tflog.Info(ctx, "Attempting to delete the policy anyway...")
	} else if disableResponse.StatusCode != 200 {
		tflog.Warn(ctx, fmt.Sprintf("Failed to disable policy, expected status code 200, got %d", disableResponse.StatusCode))
		tflog.Info(ctx, "Attempting to delete the policy anyway...")
	} else {
		tflog.Info(ctx, fmt.Sprintf("Policy with ID %s disabled successfully before deletion", policyID))
	}

	response, err := a.client.DoRequest(
		ctx,
		"DELETE",
		fmt.Sprintf("/api/policies/%s", policyID),
		nil,
		map[string]string{},
		map[string]string{},
	)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return APIErrorFromResponse(response.StatusCode, response.Body)
	}

	return nil
}

// NewSecretsHubAPI creates a new SecretsHubAPI client.
func NewSecretsHubAPI(baseURL string, authToken []byte) SecretsHubAPI {
	return &secretsHubAPI{
		client:    NewClientWithToken(baseURL, true, authToken, true),
		authToken: authToken,
	}
}

func (a *secretsHubAPI) GetSecretFilter(ctx context.Context, storeID string, filterID string) (*SecretFilterOutput, error) {
	response, err := a.client.DoRequest(
		ctx,
		"GET",
		fmt.Sprintf("/api/secret-stores/%s/filters/%s", storeID, filterID),
		nil,
		map[string]string{},
		map[string]string{},
	)

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, APIErrorFromResponse(response.StatusCode, response.Body)
	}

	output := SecretFilterOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}
