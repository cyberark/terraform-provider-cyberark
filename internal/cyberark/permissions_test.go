package cyberark_test

import (
	"encoding/json"
	"testing"

	"github.com/cyberark/terraform-provider-cyberark/internal/cyberark"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFullAdmin(t *testing.T) {
	userType := "User"
	user := "SomeUser"
	full := cyberark.Permission{
		ManageSafe:                             true,
		ManageSafeMembers:                      true,
		ViewSafeMembers:                        true,
		ViewAuditLog:                           true,
		UseAccounts:                            true,
		RetrieveAccounts:                       true,
		ListAccounts:                           true,
		AddAccounts:                            true,
		UpdateAccountContent:                   true,
		UpdateAccountProperties:                true,
		RenameAccounts:                         true,
		DeleteAccounts:                         true,
		UnlockAccounts:                         true,
		InitiateCPMAccountManagementOperations: true,
		SpecifyNextAccountContent:              true,
		BackupSafe:                             true,
		AccessWithoutConfirmation:              true,
		CreateFolders:                          true,
		DeleteFolders:                          true,
		MoveAccountsAndFolders:                 true,
		RequestsAuthorizationLevel1:            true,
		RequestsAuthorizationLevel2:            false,
	}

	t.Run("FullAdmin", func(t *testing.T) {
		resp, err := cyberark.FullAdmin(&userType, &user)
		require.NoError(t, err)

		var member cyberark.Member
		err = json.Unmarshal(resp, &member)

		assert.Equal(t, cyberark.Member{
			Member:     &user,
			MemberType: &userType,
			Perm:       full,
		}, member)

		assert.NoError(t, err)
	})

	t.Run("MissingUser", func(t *testing.T) {
		resp, err := cyberark.FullAdmin(nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestReadOnly(t *testing.T) {
	userType := "User"
	user := "SomeUser"
	read := cyberark.Permission{
		UseAccounts:      true,
		RetrieveAccounts: true,
		ListAccounts:     true,
	}

	t.Run("ReadOnly", func(t *testing.T) {
		resp, err := cyberark.ReadOnly(&userType, &user)
		require.NoError(t, err)

		var member cyberark.Member
		err = json.Unmarshal(resp, &member)

		assert.Equal(t, cyberark.Member{
			Member:     &user,
			MemberType: &userType,
			Perm:       read,
		}, member)

		assert.NoError(t, err)
	})

	t.Run("MissingUser", func(t *testing.T) {
		resp, err := cyberark.ReadOnly(nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestApprover(t *testing.T) {
	userType := "User"
	// file deepcode ignore NoHardcodedCredentials/test: This is a test file
	user := "SomeUser"
	approver := cyberark.Permission{
		UseAccounts:       true,
		RetrieveAccounts:  true,
		ListAccounts:      true,
		ViewSafeMembers:   true,
		ManageSafeMembers: true,
	}

	t.Run("Approver", func(t *testing.T) {
		resp, err := cyberark.Approver(&userType, &user)
		require.NoError(t, err)

		var member cyberark.Member
		err = json.Unmarshal(resp, &member)

		assert.Equal(t, cyberark.Member{
			Member:     &user,
			MemberType: &userType,
			Perm:       approver,
		}, member)

		assert.NoError(t, err)
	})

	t.Run("MissingUser", func(t *testing.T) {
		resp, err := cyberark.Approver(nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestManager(t *testing.T) {
	userType := "User"
	user := "SomeUser"
	manager := cyberark.Permission{
		ManageSafeMembers:                      true,
		ViewSafeMembers:                        true,
		ViewAuditLog:                           true,
		UseAccounts:                            true,
		RetrieveAccounts:                       true,
		ListAccounts:                           true,
		AddAccounts:                            true,
		UpdateAccountContent:                   true,
		UpdateAccountProperties:                true,
		RenameAccounts:                         true,
		DeleteAccounts:                         true,
		UnlockAccounts:                         true,
		InitiateCPMAccountManagementOperations: true,
		SpecifyNextAccountContent:              true,
		AccessWithoutConfirmation:              true,
	}

	t.Run("Manager", func(t *testing.T) {
		resp, err := cyberark.Manager(&userType, &user)
		require.NoError(t, err)

		var member cyberark.Member
		err = json.Unmarshal(resp, &member)

		assert.Equal(t, cyberark.Member{
			Member:     &user,
			MemberType: &userType,
			Perm:       manager,
		}, member)

		assert.NoError(t, err)
	})

	t.Run("MissingUser", func(t *testing.T) {
		resp, err := cyberark.Manager(nil, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestConjurSync(t *testing.T) {
	userType := "User"
	user := "ConjurSync"
	conjurSync := cyberark.Permission{
		UseAccounts:               true,
		RetrieveAccounts:          true,
		ListAccounts:              true,
		AccessWithoutConfirmation: true,
	}

	resp, err := cyberark.ConjurSync()
	require.NoError(t, err)

	var member cyberark.Member
	err = json.Unmarshal(resp, &member)

	assert.Equal(t, cyberark.Member{
		Member:     &user,
		MemberType: &userType,
		Perm:       conjurSync,
	}, member)

	assert.NoError(t, err)
}

func TestSecretsHub(t *testing.T) {
	userType := "User"
	user := "SecretsHub"
	conjurSync := cyberark.Permission{
		ViewSafeMembers:           true,
		RetrieveAccounts:          true,
		ListAccounts:              true,
		AccessWithoutConfirmation: true,
	}

	resp, err := cyberark.SecretsHub()
	require.NoError(t, err)

	var member cyberark.Member
	err = json.Unmarshal(resp, &member)

	assert.Equal(t, cyberark.Member{
		Member:     &user,
		MemberType: &userType,
		Perm:       conjurSync,
	}, member)

	assert.NoError(t, err)
}
