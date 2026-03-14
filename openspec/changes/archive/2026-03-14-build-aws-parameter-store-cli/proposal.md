## Why

`dotenv_cryption` is built for team development and carries multi-cloud secret-management concerns that are unnecessary for a personal workflow. A dedicated personal CLI that preserves the familiar `.env` encryption flow while storing the master key in AWS Systems Manager Parameter Store will make setup simpler, cheaper, and easier to automate in a single AWS account.

## What Changes

- Build a Go CLI for personal `.env` encryption and decryption using the same AES-256-GCM envelope-encryption model and ENVC file format as the existing `dotenv_cryption` tool.
- Use AWS Systems Manager Parameter Store as the only master-key backend instead of GCP Secret Manager or AWS Secrets Manager.
- Define an AWS-only `dotenv.yaml` schema for region, parameter name, crypto defaults, and encrypted-file naming behavior.
- Provide `encrypt`, `decrypt`, `create master`, and `delete master` commands, plus README guidance for AWS authentication and Parameter Store setup.

## Capabilities

### New Capabilities

- `config-management`: Load and validate AWS-only CLI configuration from `dotenv.yaml`.
- `parameter-store-provider`: Retrieve, create, and delete the master key in AWS Systems Manager Parameter Store.
- `encryption-engine`: Encrypt and decrypt dotenv payloads with AES-256-GCM envelope encryption.
- `file-format`: Read and write the ENVC binary envelope format for encrypted files.
- `cli-commands`: Expose encryption, decryption, and master-key lifecycle commands for the personal CLI.

### Modified Capabilities

- None.

## Impact

- New Go module and CLI entrypoints under `main.go` and `cmd/`
- Core packages for config loading, encryption, file format, provider integration, and orchestration under `internal/`
- AWS SDK v2 SSM dependency plus Cobra/Viper-based CLI and config handling
- README and sample configuration for AWS Parameter Store-backed personal usage
