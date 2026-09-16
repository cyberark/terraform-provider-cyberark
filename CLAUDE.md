# CLAUDE.md

## Overview

This repo is the **CyberArk Terraform Provider** (`terraform-provider-cyberark`), a Terraform plugin that lets Terraform manage CyberArk resources via the CyberArk Identity Security Platform Shared Services and, for on-prem deployments, the PVWA API directly. It is broader than just Conjur/Secrets Hub: it manages Safes and Accounts (AWS, Azure, MySQL/DB) in **Privilege Cloud** and **PAM Self-Hosted**, plus **Secret Stores** and **Sync Policies** in **Secrets Hub**. It fits into the wider CyberArk platform ecosystem as the IaC entry point for provisioning these resources declaratively.

## Tech Stack

- **Go** 1.25 (see `go.mod`; README mentions a lower minimum of 1.21 for building)
- Built on the [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) (`hashicorp/terraform-plugin-framework`, `terraform-plugin-go`, `terraform-plugin-log`)
- `terraform-plugin-docs` for generating docs; `terraform-plugin-testing` + `testify` for tests
- Distributed via GoReleaser (`.goreleaser.yml`) and the Terraform Registry

## Build

- Local build: `go build -o <plugin-dir>/terraform-provider-cyberark main.go` (see README "Compile source code" section for the full `~/.terraform.d/plugins/...` target path Terraform expects)
- `main.go` has `go:generate` directives:
  - `terraform fmt -recursive ./examples/`
  - `tfplugindocs generate -provider-name cyberark` (regenerates `docs/` from `internal/provider` schemas + `examples/`)
- Release builds go through GoReleaser via Docker: `docker compose build goreleaser` then `docker compose run --rm --entrypoint goreleaser goreleaser release --clean --skip=validate --snapshot` (see `bin/build`). Version is taken from a `VERSION` file (`SECRETSHUB_VERSION` env var).

## Test

- **Unit tests**: `go test ./internal/cyberark/...` and `./internal/provider/...` (e.g. `go test --coverprofile=output/c.out -v ./internal/cyberark`, per `bin/test-entrypoint.sh`)
- **Acceptance/integration tests**: run against a real CyberArk tenant using Docker Compose (`docker-compose.test.yml`, `Dockerfile.terraform`, `Dockerfile.test`) and require `TF_VAR_*` env vars (tenant, domain, client_id/secret, AWS/Azure/DB test fixtures, etc.). Driven by `bin/integration-test.sh` and `bin/test-entrypoint.sh`; CI orchestration lives in `Jenkinsfile`. These are not runnable without live CyberArk credentials/infra.
- Mocks for the CyberArk API clients live alongside source: `internal/cyberark/mock_pam.go`, `internal/cyberark/mock_secretshub.go`, `internal/provider/mock_provider.go`.

## Repo Structure

- `main.go` — provider entrypoint (`providerserver.Serve`), registers `registry.terraform.io/cyberark/cyberark`
- `internal/provider/` — Terraform-facing layer: `provider.go` (provider schema/config: `tenant`, `client_id`, `client_secret`, `domain`, `pvwa_*` fields), plus one `resource_*.go`/`data_source_*.go` file per resource/data source (e.g. `resource_aws_account.go`, `resource_safe_object.go`, `resource_sync_policy.go`, `resource_pvwa_*` for on-prem PVWA variants, `data_source_auth_token.go`)
- `internal/cyberark/` — API client layer: `auth.go` (Identity Security Platform auth), `pam.go` (Privilege Cloud/PAM API), `secretshub.go` (Secrets Hub API), `client.go`, `permissions.go`, `validation.go`, `types.go`, `error.go`
- `docs/` — generated Terraform Registry docs (`resources/`, `data-sources/`, `index.md`) — **generated via `tfplugindocs`, do not hand-edit content that a regen would overwrite**
- `examples/` — sample `.tf` configs per resource/data source, used both as documentation and as the source `tfplugindocs` renders into `docs/`
- `test/` — per-resource CRUD test fixtures (`create`/`update` subfolders) used by the Docker-based acceptance tests
- `bin/` — shell scripts for build, test, integration-test, notices update, cleanup
- `.goreleaser.yml`, `Dockerfile.goreleaser`, `Dockerfile.terraform`, `Dockerfile.test`, `docker-compose*.yml` — build/release/test tooling
- `Jenkinsfile` — CI pipeline definition

## Conventions & Gotchas

- Two authentication paths coexist: **cloud** (Identity Security Platform Shared Services: `tenant`/`client_id`/`client_secret`/`domain`) for Privilege Cloud & Secrets Hub, and **on-prem PVWA** (`pvwa_username`/`pvwa_password`/`pvwa_url`/`pvwa_login_method`) — resources are split accordingly (`resource_aws_account.go` vs `resource_pvwa_aws_account.go`, etc.). Know which path a resource targets before editing it.
- Sensitive fields (`client_secret`, passwords, keys) must be marked `Sensitive: true` in the schema so Terraform redacts them from logs/state output — follow this pattern for any new attribute.
- Docs in `docs/` are generated output; update `examples/` and the Go schema `Description` fields instead, then regenerate with `go generate ./...` (runs the `tfplugindocs` command in `main.go`).
- Acceptance tests hit real CyberArk tenants and provision real Safes/Accounts/Secret Stores; they require secrets not present in this repo and are primarily run in CI (Jenkins) with `INFRAPOOL_*` secrets, not expected to run locally without credentials.
- Provider internal package name is `provider` but historically referred to as the "SecretHub provider" in code comments (`package provider implements the SecretHub provider for Terraform`) — the scope has since expanded beyond Secrets Hub to PAM/Privilege Cloud, so don't assume comments always match current scope.
