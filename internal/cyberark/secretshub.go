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
	AddAzureAkvSecretStore(ctx context.Context, body SecretStoreInput[CreateAzureAkvData]) (*SecretStoreOutput[CreateAzureAkvData], error)
	GetAwsAsmSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AwsAsmData], error)
	GetAzureAkvSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[CreateAzureAkvData], error)
	GetAwsAsmSecretStores(ctx context.Context) (*SecretStoresOutput[AwsAsmData], error)
	GetAzureAkvSecretStores(ctx context.Context) (*SecretStoresOutput[CreateAzureAkvData], error)
	UpdateSecretStore(ctx context.Context)
	DeleteSecretStore(ctx context.Context)
}

// ScanSecretStore is an interface for interacting with SecretsHub's secret store scans.
type ScanSecretStore interface {
	ScanDefinition(ctx context.Context, details TriggerScanInputBody) (TriggerScanOutput, error)
}

// SyncPolicy is an interface for interacting with SecretsHub's sync policies.
type SyncPolicy interface {
	AddSyncPolicy(ctx context.Context, pi PolicyInput) (*PolicyExternalOutput, error)
	GetSyncPolicy(ctx context.Context, policyID string) (*PolicyExternalOutput, error)
	GetSyncPolicies(ctx context.Context) (*SyncResponse, error)
	UpdateSyncPolicy(ctx context.Context)
	DeleteSyncPolicy(ctx context.Context)
}

// SecretsHubAPI is an interface for interacting with the SecretsHub APIs.
type SecretsHubAPI interface {
	SecretStore
	ScanSecretStore
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
	if len(output.ID) == 0 {
		return nil, fmt.Errorf("failed to get secret store ID")
	}
	return &output, nil
}

// AddAzureAkvSecretStore adds a new Azure AKS secret store to the SecretsHub.
func (a *secretsHubAPI) AddAzureAkvSecretStore(ctx context.Context, body SecretStoreInput[CreateAzureAkvData]) (*SecretStoreOutput[CreateAzureAkvData], error) {
	var output SecretStoreOutput[CreateAzureAkvData]
	err := a.addSecretStore(ctx, body, &output)
	if err != nil {
		return nil, err
	}
	if len(output.ID) == 0 {
		return nil, fmt.Errorf("failed to get secret store ID")
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

	if response.StatusCode == 409 {
		tflog.Info(ctx, "Secret Store already exists.")
		return nil
	} else if response.StatusCode != 201 {
		return fmt.Errorf("failed to add secret store, expected status code 201, got %d", response.StatusCode)
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
	if output.Data == nil {
		return nil, fmt.Errorf("secret store data is empty: %+v", output)
	}
	return &output, nil
}

// GetAzureAkvSecretStore retrieves a secret store from the Azure AKV SecretsHub.
func (a *secretsHubAPI) GetAzureAkvSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[CreateAzureAkvData], error) {
	var output SecretStoreOutput[CreateAzureAkvData]
	err := a.getSecretStore(ctx, storeID, &output)
	if err != nil {
		return nil, err
	}
	if output.Data == nil {
		return nil, fmt.Errorf("secret store data is empty: %+v", output)
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
		return fmt.Errorf("failed to get secret store, expected status code 200, got %d", response.StatusCode)
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
func (a *secretsHubAPI) GetAzureAkvSecretStores(ctx context.Context) (*SecretStoresOutput[CreateAzureAkvData], error) {
	var output SecretStoresOutput[CreateAzureAkvData]
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
		return fmt.Errorf("failed to get secret stores, expected status code 200, got %d", response.StatusCode)
	}

	err = json.NewDecoder(response.Body).Decode(output)
	if err != nil {
		return err
	}
	return nil

}

// UpdateSecretStore updates a secret store in the SecretsHub.
func (a *secretsHubAPI) UpdateSecretStore(_ context.Context) {
}

// DeleteSecretStore deletes a secret store from the SecretsHub.
func (a *secretsHubAPI) DeleteSecretStore(_ context.Context) {
}

// ScanDefinition triggers a scan for a secret store in the SecretsHub.
func (a *secretsHubAPI) ScanDefinition(ctx context.Context, details TriggerScanInputBody) (TriggerScanOutput, error) {
	body, err := json.Marshal(details)
	if err != nil {
		return TriggerScanOutput{}, err
	}
	// This REST APIs is a Beta version.
	headers := map[string]string{
		"Accept": "application/x.secretshub.beta+json",
	}

	response, err := a.client.DoRequest(
		ctx,
		"POST",
		"/api/scan-definitions/secret-store/default/scan",
		bytes.NewBuffer(body),
		headers,
		map[string]string{},
	)
	if err != nil {
		return TriggerScanOutput{}, err
	}

	// Doc shows that the response code can be 200 but API returns 202 so securing the other case.
	if response.StatusCode != 200 && response.StatusCode != 202 {
		return TriggerScanOutput{}, fmt.Errorf("failed to trigger scan, expected status code 200/202, got %d", response.StatusCode)
	}

	output := TriggerScanOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return TriggerScanOutput{}, err
	}

	return output, nil
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

	if response.StatusCode == 409 {
		tflog.Info(ctx, fmt.Sprintf("Sync policy [%s] already exists.", *pi.Name))
		return nil, nil
	} else if response.StatusCode != 201 {
		return nil, fmt.Errorf("failed to add sync policy, expected status code 201, got %d", response.StatusCode)
	}

	output := PolicyExternalOutput{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	tflog.Info(ctx, fmt.Sprintf("Sync policy created with ID: %s", *output.ID))

	return &output, nil
}

// GetSyncPolicy retrieves a sync policy from the SecretsHub.
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
		return nil, fmt.Errorf("failed to get sync policy, expected status code 200, got %d", response.StatusCode)
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
		return nil, fmt.Errorf("failed to get sync policies, expected status code 200, got %d", response.StatusCode)
	}

	output := SyncResponse{}
	err = json.NewDecoder(response.Body).Decode(&output)
	if err != nil {
		return nil, err
	}

	return &output, nil
}

// UpdateSyncPolicy updates a sync policy in the SecretsHub.
func (a *secretsHubAPI) UpdateSyncPolicy(_ context.Context) {
}

// DeleteSyncPolicy deletes a sync policy from the SecretsHub.
func (a *secretsHubAPI) DeleteSyncPolicy(_ context.Context) {
}

// NewSecretsHubAPI creates a new SecretsHubAPI client.
func NewSecretsHubAPI(baseURL string, authToken []byte) SecretsHubAPI {
	return &secretsHubAPI{
		client:    NewClientWithToken(baseURL, true, authToken),
		authToken: authToken,
	}
}
