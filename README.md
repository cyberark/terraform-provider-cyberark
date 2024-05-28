# terraform-provider-secretshub

The Terraform Provider SecretsHub has the ability to interact with CyberArk Privilege Cloud Resources
and can create safes and accounts.

Note: Supported platforms for account creation are AWS, Azure, and MySQL databases.

## Certification Level
![](https://img.shields.io/badge/Certification%20Level-Community-28A745?link=https://github.com/cyberark/community/blob/master/Conjur/conventions/certification-levels.md#community)

This repo is a **Community** level project.
For more detailed information on our certification levels, see [our community guidelines](https://github.com/cyberark/community/blob/main/Conjur/conventions/certification-levels.md#community).

## Installation

### Binaries (Recommended)
The recommended way to install `terraform-provider-secretshub` is to use the binary distributions from this project's
[GitHub Releases page](https://github.com/cyberark/terraform-provider-secretshub/releases).
The packages are available for Linux, macOS and Windows.

Download and uncompress the latest release for your OS. This example uses the linux binary.

_Note: Replace `$VERSION` with the one you want to use. See [releases](https://github.com/cyberark/terraform-provider-secretshub/releases)
page for available versions._

```sh
$ wget https://github.com/cyberark/terraform-provider-secretshub/releases/download/v$VERSION/terraform-provider-secretshub-$VERSION-linux-amd64.tar.gz
$ tar -xvf terraform-provider-secretshub*.tar.gz
```

If you already have an unversioned plugin that was previously downloaded, we first need
to remove it:

```sh
$ rm -f ~/.terraform.d/plugins/terraform-provider-secretshub
```

Now copy the new binary to the Terraform's plugins folder. If this is your first plugin,
you will need to create the folder first.

```sh
$ mkdir -p ~/.terraform.d/plugins/
$ mv terraform-provider-secretshub*/terraform-provider-secretshub* ~/.terraform.d/plugins/
```

### Homebrew (MacOS)

Add and update the [CyberArk Tools Homebrew tap](https://github.com/cyberark/homebrew-tools).

```sh
$ brew tap cyberark/tools
```

Install the provider and symlink it to Terraform's plugins directory. Symlinking is
necessary because [Homebrew is sandboxed and cannot write to your home directory](https://github.com/Homebrew/brew/issues/2986).

_Note: Replace `$VERSION` with the appropriate plugin version_

```sh
$ brew install terraform-provider-secretshub

$ mkdir -p ~/.terraform.d/plugins/

$ # If Homebrew is installing somewhere other than `/usr/local/Cellar`, update the path as well.
$ ln -sf /usr/local/Cellar/terraform-provider-secretshub/$VERSION/bin/terraform-provider-secretshub_* \
    ~/.terraform.d/plugins/
```

### Compile from Source

If you wish to compile the provider from source code, you will first need Go installed
on your machine (version >=1.21 is required).

Clone repository and go into the cloned directory

```sh
$ git clone https://github.com/cyberark/terraform-provider-secretshub.git
$ cd terraform-provider-secretshub
```

- Build the provider

```sh
$ mkdir -p ~/.terraform.d/plugins/
$ # Note: If a static binary is required, use ./bin/build to create the executable
$ go build -o ~/.terraform.d/plugins/terraform-provider-secretshub main.go
```

## Configuration with Environment Variables

In order to use environment variables with SecretsHub Terraform provider use the Terraform variables and [standard mechanism](https://developer.hashicorp.com/terraform/language/values/variables#environment-variables).

### Example

```terraform
variable "secret_key" {
  type      = string
  sensitive = true
}

provider "secretshub" {
  tenant        = "aarp0000"
  domain        = "example-domain"
  client_id     = "automation@cyberark.cloud.aarp0000"
  client_secret = var.secret_key
}
```

```sh
$ export TF_VAR_secret_key=my-secret-key
$ terraform init
$ terraform plan
```

## Set up Terraform User

- Log into Identity Administration and navigate to the Users Widget

<img src="img/users-widget.png" width="60%" height="30%">

- Create New User

<img src="img/add-user-widget.png"  width="60%" height="30%">

- Populate User Data

<img src="img/terraform-user.png"  width="60%" height="30%">

- Navigate to the Roles Widget

<img src="img/roles-widget.png" width="60%" height="30%">

- Add the new user to the Privilege Cloud Safe Managers Role

<img src="img/priv-safe-manager.png" width="60%" height="30%">

- Search for the Terraform User and Add

<img src="img/add-terraform-user.png" width="60%" height="30%">

## Documentation

### Provider
[SecretsHub provider](docs/index.md)

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

## Contributing
We welcome contributions of all kinds to this repository. For instructions on how to get started and descriptions
of our development workflows, please see our [contributing guide](CONTRIBUTING.md).

## License
This repository is licensed under Apache License 2.0 - see [`LICENSE`](LICENSE) for more details.
