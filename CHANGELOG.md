# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [0.3.3] - 2025-05-29

### Added
- Added support for the `connector_pool_id` attribute in the `azure_secret_store` resource
- Added the `SecretStoreState` resource to allow users to enable or disable a secret store

### Fixed
- Removed sync policy update support in Terraform (CyberArk SecretsHub limitation).
- Made the transformation parameter optional.
- Updated README to clarify sync policy update limitations.
- Regenerated documentation to reflect the latest code changes.

## [0.3.2] - 2025-04-30

### Fixed
- Fix bug where Read and Destroy operations could fail if nested resource attributes were not set

## [0.3.1] - 2025-04-23

### Fixed
- Provide error messages from the API back to the user
- Make `retention` and `purge` fields for safes `computed` so the API can provide default values if they are not user-provided
- Safe resource update should look at safe members in the plan, not the existing state
- Handle removing or adding the optional `address` property when updating an account
- Fix crash when debug logging is enabled

## [0.3.0] - 2025-04-11

### Added
- Added support for Import, Update, Delete (Destroy) operations for all supported resources
- Removed Secretstore scan after AWS/Azure Secretstore creation to avoid conflicts
- Upgraded Go to 1.23 to resolve gocovmerge dependency issues

## [0.2.2] - 2024-11-22

### Added
- Made the SecretNameInSecretStore parameter optional for PAM accounts

## [0.2.1] - 2024-10-24

### Added
- Updated the README indicating support for PAM Self-Hosted

## [0.2.0] - 2024-10-09

### Added
- Updated the references as per repo name
- Support for PAM Self-Hosted

## [0.1.1] - 2024-09-16

### Added
- Acceptance Testing
- Automated Testing
- Updated the README

## [0.1.0] - 2023-05-19

### Added
- Initial release
