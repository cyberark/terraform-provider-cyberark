---
name: ai-pr-review
description: Automated audit of code changes before opening a Pull Request against this repo (terraform-provider-cyberark, the Go Terraform provider for the broader CyberArk platform: Privilege Cloud/PAM Self-Hosted Safes & Accounts, and Secrets Hub Secret Stores/Sync Policies). Analyzes the diff against the base branch, checks compliance with CyberArk standards (architecture, security, quality) and this repo's own CLAUDE.md conventions, and generates a PR_REVIEW_REPORT.md ready to paste into the PR description. Use before every commit that finalizes a task, or when the user writes "@ai-pr-review run analysis and generate report".
---

# AI PR Review

Automated audit of code changes before opening a Pull Request. The skill analyzes the diff, checks compliance with the CyberArk and repo-specific conventions used in `terraform-provider-cyberark`, and generates a report ready to paste into the PR description.

## Workflow

### 1. Context Extraction

- Run `git diff` against the base branch (`main`).
- Identify the changed components (see `CLAUDE.md` for the full layout):
  - `internal/provider/` — Terraform-facing resources/data sources (`resource_*.go`/`data_source_*.go`), one file per resource, split by target: cloud (`resource_aws_account.go`, `resource_safe_object.go`, `resource_sync_policy.go`, ...) vs. on-prem PVWA (`resource_pvwa_*`)
  - `internal/cyberark/` — API client layer: `auth.go` (Identity Security Platform auth), `pam.go` (Privilege Cloud/PAM), `secretshub.go` (Secrets Hub), `client.go`, `permissions.go`, `validation.go`, `types.go`, `error.go`
  - `docs/`, `examples/` — generated Registry docs and their source `.tf` examples

### 2. Multi-Persona Analysis

- **Architect**:
  - This provider has **two coexisting auth paths**: cloud (Identity Security Platform Shared Services — `tenant`/`client_id`/`client_secret`/`domain`) for Privilege Cloud & Secrets Hub, and on-prem PVWA (`pvwa_username`/`pvwa_password`/`pvwa_url`/`pvwa_login_method`). Any new resource must be placed in the correct file pattern for the path it targets (`resource_aws_account.go` vs `resource_pvwa_aws_account.go`), not mixed together.
  - Hand-edits to `docs/` are a hard blocker — regenerate via `go generate ./...` (`tfplugindocs`) from `internal/provider` schema `Description` fields and `examples/` instead.
- **Security**:
  - Any new sensitive attribute (secrets, passwords, keys, client secrets) MUST be marked `Sensitive: true` in the Terraform schema so it's redacted from logs/state output — flag any new schema field holding a credential-like value that lacks this.
  - No hardcoded tenant credentials, PVWA credentials, or API tokens in provider code, `examples/`, or `test/` fixtures.
  - Acceptance tests provision **real** Safes/Accounts/Secret Stores against live CyberArk tenants — a PR should not introduce test fixtures that could leak `INFRAPOOL_*` CI secrets or hit production-like tenants by default.
- **Quality**:
  - New logic in `internal/cyberark/` or `internal/provider/` needs unit tests (`go test ./internal/cyberark/...`, `./internal/provider/...`); mocks (`internal/cyberark/mock_pam.go`, `mock_secretshub.go`, `internal/provider/mock_provider.go`) should be extended consistently with existing patterns, not hand-rolled per-test.
  - New resources should add corresponding CRUD fixtures under `test/<resource>/{create,update}` for the Docker-based acceptance suite where applicable.

### 3. Standard Compliance

- Go version follows `go.mod` (currently 1.27.3; README documents a lower build-time minimum of 1.21 — don't let these drift further apart without updating both).
- Release versioning comes from the `VERSION` file (`SECRETSHUB_VERSION` env var consumed by GoReleaser) — a release-worthy PR should bump this consistently with `.goreleaser.yml` expectations.
- Don't assume package/comment naming reflects current scope — the `provider` package is still referred to as the "SecretHub provider" in some comments even though scope now spans PAM/Privilege Cloud too; don't propagate that stale naming into new code without checking current intent.

### 4. Report Generation

Generate `PR_REVIEW_REPORT.md` **fully filled in** — never leave a `<placeholder>` in the final file. Derive each field automatically:

- **One-line summary**: one imperative-mood sentence describing what the diff does.
- **Desired Outcome**: if the branch name or commit messages reference a ticket/issue, pull context from it (e.g. `gh issue view <n>` for a `#<n>` reference); otherwise infer the motivating problem directly from what the diff changes.
- **Implemented Changes**: build this directly from the actual `git diff` and the Multi-Persona Analysis findings (step 2) — what changed, why, and how a reviewer should verify it (manual/integration steps if any). State screenshots as "N/A" unless the diff touches UI.
- **Connected Issue/Story**: scan the current branch name and the commit messages on this branch for a ticket pattern like `[A-Z]+-\d+` (e.g. `CP-26179`, `CNJR-1234`) and use the first match as the "CyberArk internal issue ID". If a GitHub issue number is referenced (`#123` in a branch name or commit message), use it for "Resolves #123" — otherwise omit that line entirely rather than leaving a placeholder.
- **Definition of Done** — tick exactly one box per subsection based on what the diff actually contains, never by default:
  - *Changelog*: tick "CHANGELOG has been updated" only if `CHANGELOG.md` (or equivalent) is part of the diff; tick "does not require a CHANGELOG update" only if the change has no user-facing effect.
  - *Test coverage*: tick "includes new tests" only if test files changed in the diff; tick "does not require tests" only if the change is non-functional (docs, config, tooling, skills).
  - *Documentation*: tick "docs were updated" only if README/docs files changed; tick "does not require updating documentation" only if nothing user-facing changed.
  - *Behavior*: tick "no behavior was changed" only if the diff is limited to non-runtime files (skills, docs, CI config, tests); otherwise tick one of the other two options and explain why.
  - *Security*: tick "no security aspects" only if the Security persona (step 2) genuinely found nothing; otherwise tick the appropriate other option and reference the finding.

If a field genuinely cannot be determined, write a short explicit note (e.g. "No linked issue found in branch/commit history") instead of leaving an angle-bracket placeholder.

The output must follow this structure when pasted into a PR description:

```markdown
<filled summary>

### Desired Outcome

<filled>

### Implemented Changes

- <filled bullets>

### Connected Issue/Story

Resolves #<n>  (omit this line if none found)

CyberArk internal issue ID: <filled or note>

### Definition of Done

#### Changelog
- [ ] The CHANGELOG has been updated, or
- [ ] This PR does not include user-facing changes and doesn't require a CHANGELOG update

#### Test coverage
- [ ] This PR includes new unit and integration tests to go with the code changes, or
- [ ] The changes in this PR do not require tests

#### Documentation
- [ ] Docs (e.g. READMEs) were updated in this PR
- [ ] A follow-up issue to update official docs has been filed here: [insert issue ID]
- [ ] This PR does not require updating any documentation

#### Behavior
- [ ] This PR changes product behavior and has been reviewed by a PO, or
- [ ] These changes are part of a larger initiative that will be reviewed later, or
- [ ] No behavior was changed with this PR

#### Security
- [ ] Security architect has reviewed the changes in this PR,
- [ ] These changes are part of a larger initiative with a separate security review, or
- [ ] There are no security aspects to these changes
```

(the checkboxes above are the template shape — in the actual generated report, exactly one box per subsection must be ticked `[x]`, the rest `[ ]`)

## Usage Instructions

Run this command before every commit that finalizes a task:

```
@ai-pr-review run analysis and generate report
```

## Port.io Metadata

- **Maturity Level**: Tier 2 (Repeatable Skill)
- **Service**: Secrets Manager Integrations
