package cyberark

import (
	"context"
	"github.com/stretchr/testify/mock"
)

type MockPAMAPI struct {
	mock.Mock
}

func NewMockPAMAPI() *MockPAMAPI {
	return &MockPAMAPI{}
}

func (m *MockPAMAPI) AddAccount(ctx context.Context, credential Credential) (*CredentialResponse, error) {
	args := m.Called(ctx, credential)
	return args.Get(0).(*CredentialResponse), args.Error(1)
}

func (m *MockPAMAPI) GetAccount(ctx context.Context, accountID string) (*CredentialResponse, error) {
	args := m.Called(ctx, accountID)
	return args.Get(0).(*CredentialResponse), args.Error(1)
}

func (m *MockPAMAPI) FilterAccounts(ctx context.Context, search string, filter []string) (*CredentialSearchResponse, error) {
	args := m.Called(ctx, search, filter)
	return args.Get(0).(*CredentialSearchResponse), args.Error(1)
}

func (m *MockPAMAPI) UpdateAccount(ctx context.Context, accountID string, credential Credential) (*CredentialResponse, error) {
	args := m.Called(ctx, accountID, credential)
	return args.Get(0).(*CredentialResponse), args.Error(1)
}

func (m *MockPAMAPI) DeleteAccount(ctx context.Context, accountID string) error {
	args := m.Called(ctx, accountID)
	return args.Error(0)
}

func (m *MockPAMAPI) AddSafe(ctx context.Context, safe SafeData) (*SafeData, error) {
	args := m.Called(ctx, safe)
	return args.Get(0).(*SafeData), args.Error(1)
}

func (m *MockPAMAPI) GetSafe(ctx context.Context, safeID string) (*SafeData, error) {
	args := m.Called(ctx, safeID)
	return args.Get(0).(*SafeData), args.Error(1)
}

func (m *MockPAMAPI) UpdateSafe(ctx context.Context, safeID string, safe SafeData) (*SafeData, error) {
	args := m.Called(ctx, safeID, safe)
	return args.Get(0).(*SafeData), args.Error(1)
}

func (m *MockPAMAPI) DeleteSafe(ctx context.Context, safeID string) error {
	args := m.Called(ctx, safeID)
	return args.Error(0)
}

func (m *MockPAMAPI) AddSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	args := m.Called(ctx, safe)
	return args.Get(0).(*Member), args.Error(1)
}

func (m *MockPAMAPI) GetSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	args := m.Called(ctx, safe)
	return args.Get(0).(*Member), args.Error(1)
}

func (m *MockPAMAPI) UpdateSafeMember(ctx context.Context, safe SafeData) (*Member, error) {
	args := m.Called(ctx, safe)
	return args.Get(0).(*Member), args.Error(1)
}

func (m *MockPAMAPI) DeleteSafeMember(ctx context.Context, safeName string, memberName string) error {
	args := m.Called(ctx, safeName, memberName)
	return args.Error(0)
}
