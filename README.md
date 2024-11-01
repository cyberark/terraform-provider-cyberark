## CyberArk Terraform Provider

This topic describes how to integrate Terraform with Secrets Hub using the CyberArk Terraform Provider.

## Certification level
![](https://img.shields.io/badge/Certification%20Level-Certified-28A745?link=https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md)

## Overview

The CyberArk Terraform Provider is open source and available on GitHub.

The CyberArk Terraform Provider can interact with CyberArk resources including Privileged Cloud, PAM Self-Hosted, and Secrets Hub and can create safes, accounts, secret stores and sync policies.

Note: Supported platforms for account creation are AWS, Azure, and MySQL databases.

The CyberArk Terraform Provider includes the following features and benefits:

Configuration in the Terraform manifest

Provider authentication to CyberArk Identity Security Platform Shared Services

A Provider can create Safes and Accounts in CyberArk Privileged Cloud and PAM Self-Hosted, as well as Secret Stores and Sync Policies in Secrets Hub

Terraform's `sensitive` variable argument is used by default for sensitive values to keep the value from appearing in logs and on-screen.

## Authentication

The CyberArk Terraform Provider authenticates to CyberArk Identity Security Platform Shared Services with the service account and its credential. When using PAM Self-Hosted, the provider authenticates with the on-prem PVWA instance.

### Set up Service Account

- Log into Identity Administration and navigate to the Users Widget

<img src="img/users-widget.png" width="60%" height="30%">

- Create New User

<img src="img/add-user-widget.png"  width="60%" height="30%">

- Populate User Data

<img src="img/terraform-user.png"  width="60%" height="30%">


## Authorization to access Privilege Cloud and Secrets Hub

Assign the Privilege Cloud Safe Managers Role and the Secrets Manager - Secrets Hub Admin Role to the Service Account.

- Log into Identity Administration and navigate to the Roles Widget

<img src="img/roles-widget.png" width="60%" height="30%">

- Add the new user to the Privilege Cloud Safe Managers Role

<img src="img/priv-safe-manager.png" width="60%" height="30%">

- Search for the Terraform User and Add

<img src="img/add-terraform-user.png" width="60%" height="30%">

- Add the new user to the Secrets Manager - Secrets Hub Admin Role

- Search for the Terraform User and Add

<img src="img/add-terraform-user.png" width="60%" height="30%">

## Requirements

CyberArk Terraform Provider requirements

### Technology

- Go - 1.21
- Terraform - 1.75 or later

### Services

- A tenant with Privilege Cloud and Secrets Hub is required.
- An AWS account with the SecretHub IAM role is necessary.

## Supported platforms

- macOS
- Linux
- Windows

## Install the CyberArk Terraform Provider plugin

You can use any of the following methods to install the CyberArk Terraform Provider plugin:

Install using binaries (Recommended)

Compile source code

Access from the Terraform registry

Install using Homebrew (macOS only)

### Binaries (Recommended)

We recommend installing the CyberArk Terraform Provider plugin (terraform-provider-cyberark) using the appropriate binary distribution for your environment.

In the following examples, replace `$VERSION` with the latest release for your operating system from the GitHub Releases page.

Note: The following example uses a Linux binary.

1. Download the CyberArk Terraform Provider (darwin_amd64 or linux_amd64):

```sh
$  wget https://github.com/cyberark/terraform-provider-cyberark/releases/download/v$VERSION/terraform-provider-cyberark_$VERSION.linux_amd64.zip
```
2. Create a new subdirectory:

```sh
$ mkdir -p ~/.terraform.d/plugins/terraform.example.com/cyberark/cyberark/$VERSION/linux_amd64
```
3. Decompress the binary into the appropriate plugins directory:

```sh
$ unzip terraform-provider-cyberark_$VERSION_linux_amd64.zip ~/.terraform.d/plugins/terraform.example.com/cyberark/cyberark/$VERSION/linux_amd64
```
4. To uninstall or remove the previous version of the plugin, run the following command:

```sh
$ rm -rf ~/.terraform.d/plugins/terraform.example.com/cyberark/cyberark/$VERSION/linux_amd64
```

### Homebrew (MacOS)
To install the CyberArk Terraform Provider using Homebrew:

1. Add and update the CyberArk Tools Homebrew tap:

```sh
$ brew tap cyberark/tools
```

2. Install the CyberArk Terraform Provider and symlink it to Terraform's plugins directory. Symlinking is necessary because Homebrew is sandboxed and cannot write to your home directory.

   Run the following, where $VERSION is the appropriate plugin version:
_Note: Replace `$VERSION` with the appropriate plugin version_

```sh
$ brew install terraform-provider-cyberark

$ mkdir -p ~/.terraform.d/plugins/

$ # If Homebrew is installing somewhere other than `/usr/local/Cellar`, update the path as well.

$ ln -sf /usr/local/Cellar/terraform-provider-cyberark/$VERSION/bin/terraform-provider-cyberark_* \
    ~/.terraform.d/plugins/
```
3. If you have a previously downloaded unversioned plugin, remove it:
```sh
$ brew uninstall terraform-provider-cyberark
$ rm -f ~/.terraform.d/plugins/terraform-provider-cyberark
```
4. Create the Terraform plugins folder if it does not already exist:
```sh
$ mkdir -p ~/.terraform.d/plugins/
```
5. Copy the new binary to the Terraform plugins folder:
```sh
$ mv terraform-provider-cyberark*/terraform-provider-cyberark* ~/.terraform.d/plugins/
```

### Compile from Source

Before you compile the CyberArk Terraform Provider from the source code, make sure you have Go version 1.21 installed on your machine.

To compile the CyberArk Terraform Provider:

macOS/Linux

1. Clone the repository and open the cloned directory:

```sh
$ git clone https://github.com/cyberark/terraform-provider-cyberark.git
$ cd terraform-provider-cyberark
```

2. Build the CyberArk Terraform Provider

```sh
$ mkdir -p ~/.terraform.d/plugins/terraform.example.com/cyberark/cyberark/$VERSION/$platform_reference_in_go
# Example: platform_reference_in_go= darwin_amd64/linux_amd64
# Note: If a static binary is required, use ./bin/build to create the executable
$ go build -o ~/.terraform.d/plugins/terraform.example.com/cyberark/cyberark/$VERSION/$platform_reference_in_go/terraform-provider-cyberark main.go
```

### Terraform registry

To access the CyberArk Terraform Provider from the Terraform registry:

In the main.tf configuration file:

- In the source, use registry.terraform.io/cyberark/cyberark

- In version, provide the latest version

```sh
variable "secret_key" {
  type      = string
  sensitive = true
}

terraform {
    required_providers {
      cyberark = {
        source  = “registry.terraform.io/cyberark/cyberark"version = "~> 0"
      }
    }
  }

provider "cyberark" {
  tenant        = "aarp0000"
  domain        = "example-domain"
  client_id     = "automation@cyberark.cloud.aarp0000"
  client_secret = var.secret_key
}
resource "cyberark_safe" "AAM_Test_Safe" {
  safe_name          = "GEN_BY_TF_abc"
  safe_desc          = "Description for GEN_BY_TF_abc"
  member             = "demo@cyberark.cloud.aarp0000"
  member_type        = "user"
  permission_level   = "read" # full, read, approver, manager
  retention          = 7
  retention_versions = 7
  purge              = false
  cpm_name           = "PasswordManager"
  safe_loc           = ""
}
```

## Caution: Handling Sensitive Files

Important: The Terraform state file and .tfvars files contain sensitive information related to your configurations. It is essential to handle these files with the utmost care to ensure their security.

### Best Practices

- Keep Files Private: Ensure these files are not exposed to unauthorized individuals or systems.
- Restrict Access: Limit access to these files to authorized personnel only.
- Use Encryption: Whenever possible, use encryption for both storage and transmission to protect the contents of these files.

Following these practices helps safeguard your sensitive data.

## Configure CyberArk Terraform Provider

This section describes how to configure the CyberArk Terraform Provider.

### Workflow

Terraform can be executed manually by the user. The CyberArk Terraform Provider reads the provider configuration and authenticates to the tenant using the service account and its credentials.

Once authenticated, it configures the resources according to the main.tf file. After setup, the resources can be viewed in Privilege Cloud and Secrets Hub.

### Use environment variables for Sensitive Variables

In order to use environment variables with Terraform Provider SecrestsHub use the Terraform variables and [standard mechanism]
(https://developer.hashicorp.com/terraform/language/values/variables#environment-variables).

### Example

#### Privileged Cloud

```terraform
variable "secret_key" {
  type      = string
  sensitive = true
}

provider "cyberark" {
  tenant        = "aarp0000"
  domain        = "example-domain"
  client_id     = "automation@cyberark.cloud.aarp0000"
  client_secret = var.secret_key
}

resource "cyberark_safe" "PCloud_Test_Safe" {
  safe_name          = "GEN_BY_TF_abc"
  safe_desc          = "Description for GEN_BY_TF_abc"
  member             = "demo@cyberark.cloud.aarp0000"
  member_type        = "user"
  permission_level   = "read" # full, read, approver, manager
  retention          = 7
  retention_versions = 7
  purge              = false
  cpm_name           = "PasswordManager"
  safe_loc           = ""
}
```

#### PAM Self-Hosted

```terraform
variable "secret_key" {
  type      = string
  sensitive = true
}

variable "pvwa_password" {
  type      = string
  sensitive = true
}

provider "cyberark" {
  tenant        = "aarp0000"
  domain        = "example-domain"
  client_id     = "automation@cyberark.cloud.aarp0000"
  client_secret = var.secret_key

  # For PAM Self-Hosted
  pvwa_url      = "https://pvwa.example.com"
  pvwa_username = "myUser"
  pvwa_password = var.pvwa_password
}

resource "cyberark_pam_safe" "PAM_Test_Safe" {
  safe_name          = "GEN_BY_TF_abc"
  safe_desc          = "Description for GEN_BY_TF_abc"
  member             = "demo@cyberark.cloud.aarp0000"
  member_type        = "user"
  permission_level   = "read" # full, read, approver, manager
  retention          = 7
  retention_versions = 7
  purge              = false
  cpm_name           = "PasswordManager"
  safe_loc           = ""
  enable_olac        = false
}
```

```sh
$ export TF_VAR_secret_key=my-secret-key
$ export TF_VAR_pvwa_password=my-pvwa-password
$ terraform init
$ terraform plan
```

## Pre-requisties for Provider and Resources

- A tenant with both Privilege Cloud and Secrets Hub is required.
- Create and enable a service account and its associated secret.
- An AWS account with the SecretHub IAM role is necessary.
- Get the Privilege Cloud secret store ID via the API or user interface and insert it into the source_id section of the sync policy.

  1. UI: Log in to the CyberArk tenant with sufficient privileges to view the Privilege Cloud store details.

  2. API : Use the documentation below to make an API call and retrieve the Privilege Cloud StoreID. (https://docs.cyberark.com/secrets-hub-privilege-cloud/Latest/en/Content/Developer/sh-policy-api-tutorial.htm?tocpath=Developer%7CTutorials%7C_____4).

## Documentation

### Provider

[cyberark provider](docs/index.md)

### Data Sources

- [Auth token](docs/data-sources/auth_token.md)

### Resources

- [AWS Account](docs/resources/aws_account.md)
- [AWS Secret Store](docs/resources/aws_secret_store.md)
- [Azure Account](docs/resources/azure_account.md)
- [Azure Secret Store](docs/resources/azure_secret_store.md)
- [DB Account](docs/resources/db_account.md)
- [Safe](docs/resources/safe.md)
- [Sync Policy](docs/resources/sync_policy.md)

## Usage instructions

See [here](examples/) for examples.

## Limitations

The CyberArk Terraform Provider plugin does not support the following features:

- Update safe
- Update account
- Delete account
- Update secret store
- Delete secret store
- Update sync policy
- Delete sync policy
- Rotation of auth token
