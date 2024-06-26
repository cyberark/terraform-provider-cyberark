// Package cyberark provides a client for interacting with the SecretsHub APIs.
package cyberark

import (
	"encoding/json"
	"errors"
)

/*
=========================================
* Functions
=========================================
*/

// FullAdmin gets Full Administrator Permissions
// intakes a user type string and user string to bundle permissions
func FullAdmin(userType *string, User *string) ([]byte, error) {
	Perm := Permission{
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

	userBlock := Member{
		Member:     User,
		MemberType: userType,
		Perm:       Perm,
	}

	if User == nil || userType == nil {
		return nil, errors.New("either User or User Type is nil")
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}

// ReadOnly gets Read-Only Permissions
// intakes a user type string and user string to bundle permissions
func ReadOnly(userType *string, User *string) ([]byte, error) {
	Perm := Permission{
		UseAccounts:      true,
		RetrieveAccounts: true,
		ListAccounts:     true,
	}

	if User == nil || userType == nil {
		return nil, errors.New("either User or User Type is nil")
	}

	userBlock := Member{
		Member:     User,
		MemberType: userType,
		Perm:       Perm,
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}

// Approver gets Approver Permissions
// intakes a user type string and user string to bundle permissions
func Approver(userType *string, User *string) ([]byte, error) {
	Perm := Permission{
		UseAccounts:       true,
		RetrieveAccounts:  true,
		ListAccounts:      true,
		ViewSafeMembers:   true,
		ManageSafeMembers: true,
	}

	if User == nil || userType == nil {
		return nil, errors.New("either User or User Type is nil")
	}

	userBlock := Member{
		Member:     User,
		MemberType: userType,
		Perm:       Perm,
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}

// Manager gets Safe Manager Permissions
// intakes a user type string and user string to bundle permissions
func Manager(userType *string, User *string) ([]byte, error) {
	Perm := Permission{
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

	if User == nil || userType == nil {
		return nil, errors.New("either User or User Type is nil")
	}

	userBlock := Member{
		Member:     User,
		MemberType: userType,
		Perm:       Perm,
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}

// ConjurSync gets Conjur Component User Permissions
func ConjurSync() ([]byte, error) {
	Perm := Permission{
		UseAccounts:               true,
		RetrieveAccounts:          true,
		ListAccounts:              true,
		AccessWithoutConfirmation: true,
	}

	US := "ConjurSync"
	UT := "User"

	userBlock := Member{
		Member:     &US,
		MemberType: &UT,
		Perm:       Perm,
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}

// SecretsHub gets Secrets Hub Component User Permissions
func SecretsHub() ([]byte, error) {
	Perm := Permission{
		ViewSafeMembers:           true,
		RetrieveAccounts:          true,
		ListAccounts:              true,
		AccessWithoutConfirmation: true,
	}

	US := "SecretsHub"
	UT := "User"

	userBlock := Member{
		Member:     &US,
		MemberType: &UT,
		Perm:       Perm,
	}

	thisBlock, err := json.Marshal(userBlock)
	if err != nil {
		return nil, err
	}

	return thisBlock, nil
}
