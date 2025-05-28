// Package cyberark provides a client for interacting with the CyberArk's SecretsHub APIs.
package cyberark

// Permission represents the safe member permissions
type Permission struct {
	ManageSafe                             bool `json:"manageSafe"`
	ManageSafeMembers                      bool `json:"manageSafeMembers"`
	ViewSafeMembers                        bool `json:"viewSafeMembers"`
	ViewAuditLog                           bool `json:"viewAuditLog"`
	UseAccounts                            bool `json:"useAccounts"`
	RetrieveAccounts                       bool `json:"retrieveAccounts"`
	ListAccounts                           bool `json:"listAccounts"`
	AddAccounts                            bool `json:"addAccounts"`
	UpdateAccountContent                   bool `json:"updateAccountContent"`
	UpdateAccountProperties                bool `json:"updateAccountProperties"`
	RenameAccounts                         bool `json:"renameAccounts"`
	DeleteAccounts                         bool `json:"deleteAccounts"`
	UnlockAccounts                         bool `json:"unlockAccounts"`
	InitiateCPMAccountManagementOperations bool `json:"initiateCPMAccountManagementOperations"`
	SpecifyNextAccountContent              bool `json:"specifyNextAccountContent"`
	BackupSafe                             bool `json:"backupSafe"`
	AccessWithoutConfirmation              bool `json:"accessWithoutConfirmation"`
	CreateFolders                          bool `json:"createFolders"`
	DeleteFolders                          bool `json:"deleteFolders"`
	MoveAccountsAndFolders                 bool `json:"moveAccountsAndFolders"`
	RequestsAuthorizationLevel1            bool `json:"requestsAuthorizationLevel1"`
	RequestsAuthorizationLevel2            bool `json:"requestsAuthorizationLevel2"`
}

// ACL Struct

// Member represents member of a given type with permissions
type Member struct {
	Member     *string    `json:"memberName,omitempty"`
	MemberType *string    `json:"memberType,omitempty"`
	Perm       Permission `json:"permissions,omitempty"`
}

// Shared Services Structs

// IdentityToken represents the token response from the authentication endpoint
type IdentityToken struct {
	AccessToken *string `json:"access_token"`
	TokenType   *string `json:"token_type"`
	ExpiresIn   *int    `json:"expires_in"`
}

// Vault API Structs

// Credential represents the PAM credential
type Credential struct {
	Name       *string           `json:"name"`       // Custom Account Name of the credential
	Address    *string           `json:"address"`    // Address of where the credential is used
	UserName   *string           `json:"userName"`   // Username value
	Platform   *string           `json:"platformId"` // Required: Management platform
	SafeName   *string           `json:"safeName"`   // Required: Target Safe
	SecretType *string           `json:"secretType"` // Type of secret (use password)
	Secret     *string           `json:"secret"`     // Password Value
	SecretMgmt *SecretManagement `json:"secretManagement"`
	Props      *AccountProps     `json:"platformAccountProperties"`
}

// AccountProps represents the properties of the PAM account
type AccountProps struct {

	/*
	 * Generic - Types that map to additional platforms
	 */

	Port *string `json:"port,omitempty"`

	/*
	 * DB Handlers
	 */

	DBName                  *string `json:"database,omitempty"`
	DSN                     *string `json:"dsn,omitempty"`
	SecretNameInSecretStore *string `json:"secretnameinsecretstore,omitempty"`

	/*
	 * AWS
	 */

	AWSKID     *string `json:"AWSAccessKeyID,omitempty"`
	AWSAccount *string `json:"AWSAccountID,omitempty"`
	Alias      *string `json:"AWSAccountAliasName,omitempty"`
	Region     *string `json:"Region,omitempty"`

	/*
	 * Microsoft
	 */

	MAppID       *string `json:"ApplicationID,omitempty"`
	MAppObjectID *string `json:"ApplicationObjectID,omitempty"`
	MKID         *string `json:"KeyID,omitempty"`
	MADID        *string `json:"ActiveDirectoryID,omitempty"`
	MDur         *string `json:"Duration,omitempty"`
	MPop         *string `json:"PopulateIfNotExist,omitempty"`
	MKeyDesc     *string `json:"KeyDescription,omitempty"`
}

// SecretManagement represents the secret management properties
type SecretManagement struct {
	AutomaticManagement    *bool   `json:"automaticManagementEnabled"`
	ManualManagementReason *string `json:"manualManagementReason"`
	ModifiedTime           *int64  `json:"lastModifiedTime,omitempty"`
	Status                 *string `json:"status,omitempty"`
	LastReconcile          *int64  `json:"lastReconciledTime,omitempty"`
	LastVerified           *int64  `json:"lastVerifiedTime,omitempty"`
}

// CredentialResponse represents the credential response from the PAM API
type CredentialResponse struct {
	Name         *string           `json:"name,omitempty"`
	Address      *string           `json:"address,omitempty"`
	UserName     *string           `json:"userName,omitempty"`
	Platform     *string           `json:"platformId,omitempty"`
	SafeName     *string           `json:"safeName,omitempty"`
	SecretType   *string           `json:"secretType,omitempty"`
	Secret       *string           `json:"secret,omitempty"`
	SecretMgmt   *SecretManagement `json:"secretManagement,omitempty"`
	Props        *AccountProps     `json:"platformAccountProperties,omitempty"`
	CredID       *string           `json:"id,omitempty"`
	CreationTime *int              `json:"lastModifiedTime,omitempty"`
}

// CredentialSearchResponse represents the credential search response from the PAM API
type CredentialSearchResponse struct {
	Accounts []*CredentialResponse `json:"value"`
	Count    *int                  `json:"count"`
}

// SafeData represents the PAM safe data
type SafeData struct {
	RetentionDays        *int64  `json:"numberOfDaysRetention,omitempty"`
	RetentionVersions    *int64  `json:"numberOfVersionsRetention,omitempty"`
	PurgeEnabled         *bool   `json:"autoPurgeEnabled,omitempty"`
	CPM                  *string `json:"managingCPM,omitempty"`
	Name                 *string `json:"safeName"`
	Description          *string `json:"description,omitempty"`
	Location             *string `json:"location,omitempty"`
	URLID                *string `json:"safeUrlId,omitempty"`
	NUMBER               *int64  `json:"safeNumber,omitempty"`
	Owner                *string `json:"memberName,omitempty"`
	OwnerType            *string `json:"memberType,omitempty"`
	Level                *string `json:"omitempty"`
	LastModificationTime *int64  `json:"lastModificationTime,omitempty"`
	EnableOLAC           *bool   `json:"enableOLAC,omitempty"`
}

// API represents the CyberArk's SecretsHub and PAM API
type API struct {
	PamAPI        PAMAPI
	SecretsHubAPI SecretsHubAPI
	PVWAAPI       PAMAPI
}

// Secret stores API

// AwsAsmData represents the AWS ASM data
type AwsAsmData struct {
	AccountAlias *string `json:"accountAlias"`
	AccountID    *string `json:"accountId"`
	RegionID     *string `json:"regionId"`
	RoleName     *string `json:"roleName"`
}

// AzureAkvData represents the Azure AKV data
type AzureAkvData struct {
	AppClientDirectoryID *string    `json:"appClientDirectoryId"`
	AzureVaultURL        *string    `json:"azureVaultUrl"`
	AppClientID          *string    `json:"appClientId"`
	AppClientSecret      *string    `json:"appClientSecret"`
	Connector            *Connector `json:"connectionConfig"`
	SubscriptionID       *string    `json:"subscriptionId"`
	SubscriptionName     *string    `json:"subscriptionName"`
	ResourceGroupName    *string    `json:"resourceGroupName"`
}

// Connector represents the connector data
type Connector struct {
	ConnectionType  *string `json:"connectionType"`
	ConnectorID     *string `json:"connectorId"`
	ConnectorPoolID *string `json:"connectorPoolId"`
}

// SecretStoreInput represents the secret store input
type SecretStoreInput[T AwsAsmData | AzureAkvData] struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Type        *string `json:"type"`
	Data        *T      `json:"data"`
}

// SecretStoreOutput represents the secret store output
type SecretStoreOutput[T AwsAsmData | AzureAkvData] struct {
	ID          string    `json:"id"`
	Type        *string   `json:"type"`
	Behaviors   []*string `json:"behaviors"`
	CreatedAt   *string   `json:"createdAt"`
	CreatedBy   *string   `json:"createdby"`
	Data        *T        `json:"data"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	UpdatedAt   *string   `json:"updatedAt"`
	UpdatedBy   *string   `json:"updatedby"`
}

// SecretStoresOutput represents the generic secret stores output
type SecretStoresOutput[T AwsAsmData | AzureAkvData] struct {
	SecretStores []*SecretStoreOutput[T] `json:"secretStores"`
}

// Sync policy API

// Source represents the policy source data
type Source struct {
	SourceID string `json:"id"`
}

// Target represents the policy target data
type Target struct {
	TargetID string `json:"id"`
}

// SafeDataFilter represents the safe data filter
type SafeDataFilter struct {
	SafeName *string `json:"safeName"`
}

// Filter represents the policy filter data
type Filter struct {
	Type *string         `json:"type"`
	Data *SafeDataFilter `json:"data"`
}

// TransformationValue represents the policy transformation value
type TransformationValue struct {
	Predefined string `json:"predefined"`
}

// PolicyInput represents the policy input data
type PolicyInput struct {
	Name           *string              `json:"name"`
	Description    *string              `json:"description"`
	Source         *Source              `json:"source"`
	Target         *Target              `json:"target"`
	Filter         *Filter              `json:"filter"`
	Transformation *TransformationValue `json:"transformation"`
}

// FilterResponse represents the policy filter response
type FilterResponse struct {
	ID *string `json:"id"`
}

// State represents the current policy state
type State struct {
	CurrentState string `json:"current"`
}

// PolicyExternalOutput represents the policy external output
type PolicyExternalOutput struct {
	ID             *string              `json:"id"`
	Name           *string              `json:"name"`
	Description    *string              `json:"description"`
	CreatedAt      *string              `json:"createdAt"`
	UpdatedAt      *string              `json:"updatedAt"`
	CreatedBy      *string              `json:"createdBy"`
	UpdatedBy      *string              `json:"updatedBy"`
	Source         *Source              `json:"source"`
	Target         *Target              `json:"target"`
	Filter         *FilterResponse      `json:"filter"`
	Transformation *TransformationValue `json:"transformation"`
	State          *State               `json:"state"`
}

// SyncResponse represents the policy sync response
type SyncResponse struct {
	Count    int32                   `json:"count"`
	Policies []*PolicyExternalOutput `json:"policies"`
}

type SecretFilterOutput struct {
	ID        *string         `json:"id"`
	Type      *string         `json:"type"`
	Data      *SafeDataFilter `json:"data"`
	CreatedAt *string         `json:"createdAt"`
	UpdatedAt *string         `json:"updatedAt"`
	CreatedBy *string         `json:"createdBy"`
	UpdatedBy *string         `json:"updatedBy"`
}
