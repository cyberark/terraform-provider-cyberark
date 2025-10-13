package cyberark

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockSecretsHubAPI struct {
	mock.Mock
}

func NewMockSecretsHubAPI() *MockSecretsHubAPI {
	return &MockSecretsHubAPI{}
}

func (m *MockSecretsHubAPI) AddAwsAsmSecretStore(ctx context.Context, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error) {
	args := m.Called(ctx, body)
	return args.Get(0).(*SecretStoreOutput[AwsAsmData]), args.Error(1)
}

func (m *MockSecretsHubAPI) AddAzureAkvSecretStore(ctx context.Context, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error) {
	args := m.Called(ctx, body)
	return args.Get(0).(*SecretStoreOutput[AzureAkvData]), args.Error(1)
}

func (m *MockSecretsHubAPI) AddGcpSecretStore(ctx context.Context, body SecretStoreInput[GcpData]) (*SecretStoreOutput[GcpData], error) {
	args := m.Called(ctx, body)
	return args.Get(0).(*SecretStoreOutput[GcpData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetAwsAsmSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AwsAsmData], error) {
	args := m.Called(ctx, storeID)
	return args.Get(0).(*SecretStoreOutput[AwsAsmData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetAzureAkvSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[AzureAkvData], error) {
	args := m.Called(ctx, storeID)
	return args.Get(0).(*SecretStoreOutput[AzureAkvData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetGcpSecretStore(ctx context.Context, storeID string) (*SecretStoreOutput[GcpData], error) {
	args := m.Called(ctx, storeID)
	return args.Get(0).(*SecretStoreOutput[GcpData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetAwsAsmSecretStores(ctx context.Context) (*SecretStoresOutput[AwsAsmData], error) {
	args := m.Called(ctx)
	return args.Get(0).(*SecretStoresOutput[AwsAsmData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetAzureAkvSecretStores(ctx context.Context) (*SecretStoresOutput[AzureAkvData], error) {
	args := m.Called(ctx)
	return args.Get(0).(*SecretStoresOutput[AzureAkvData]), args.Error(1)
}

func (m *MockSecretsHubAPI) GetGcpSecretStores(ctx context.Context) (*SecretStoresOutput[GcpData], error) {
	args := m.Called(ctx)
	return args.Get(0).(*SecretStoresOutput[GcpData]), args.Error(1)
}

func (m *MockSecretsHubAPI) UpdateAwsSecretStore(ctx context.Context, storeID string, body SecretStoreInput[AwsAsmData]) (*SecretStoreOutput[AwsAsmData], error) {
	args := m.Called(ctx, storeID, body)
	return args.Get(0).(*SecretStoreOutput[AwsAsmData]), args.Error(1)
}

func (m *MockSecretsHubAPI) UpdateAzureAkvSecretStore(ctx context.Context, storeID string, body SecretStoreInput[AzureAkvData]) (*SecretStoreOutput[AzureAkvData], error) {
	args := m.Called(ctx, storeID, body)
	return args.Get(0).(*SecretStoreOutput[AzureAkvData]), args.Error(1)
}

func (m *MockSecretsHubAPI) UpdateGcpSecretStore(ctx context.Context, storeID string, body SecretStoreInput[GcpData]) (*SecretStoreOutput[GcpData], error) {
	args := m.Called(ctx, storeID, body)
	return args.Get(0).(*SecretStoreOutput[GcpData]), args.Error(1)
}

func (m *MockSecretsHubAPI) DeleteSecretStore(ctx context.Context, storeID string) error {
	args := m.Called(ctx, storeID)
	return args.Error(0)
}

func (m *MockSecretsHubAPI) SetSecretStoreState(ctx context.Context, storeID string, action string) error {
	args := m.Called(ctx, storeID, action)
	return args.Error(0)
}

func (m *MockSecretsHubAPI) AddSyncPolicy(ctx context.Context, pi PolicyInput) (*PolicyExternalOutput, error) {
	args := m.Called(ctx, pi)
	return args.Get(0).(*PolicyExternalOutput), args.Error(1)
}

func (m *MockSecretsHubAPI) GetSyncPolicy(ctx context.Context, policyID string) (*PolicyExternalOutput, error) {
	args := m.Called(ctx, policyID)
	return args.Get(0).(*PolicyExternalOutput), args.Error(1)
}

func (m *MockSecretsHubAPI) GetSyncPolicies(ctx context.Context) (*SyncResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(*SyncResponse), args.Error(1)
}

func (m *MockSecretsHubAPI) GetSecretFilter(ctx context.Context, storeID string, filterID string) (*SecretFilterOutput, error) {
	args := m.Called(ctx, storeID, filterID)
	return args.Get(0).(*SecretFilterOutput), args.Error(1)
}

func (m *MockSecretsHubAPI) DeleteSyncPolicy(ctx context.Context, policyID string) error {
	args := m.Called(ctx, policyID)
	return args.Error(0)
}

func (m *MockSecretsHubAPI) UpdateSyncPolicy(ctx context.Context, policyID string, pi PolicyInput) (*PolicyExternalOutput, error) {
	args := m.Called(ctx, policyID, pi)
	return args.Get(0).(*PolicyExternalOutput), args.Error(1)
}
