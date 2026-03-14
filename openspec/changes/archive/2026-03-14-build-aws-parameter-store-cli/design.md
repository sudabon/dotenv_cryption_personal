## Context

This repository is a greenfield personal variant of the existing `dotenv_cryption` CLI. The sibling project already establishes the desired operator experience and the stable parts of the system: AES-256-GCM envelope encryption, the ENVC file format, Cobra-based commands, and Viper-backed `dotenv.yaml` loading.

The main change is scope reduction. Instead of supporting multiple cloud secret backends for team development, this personal tool will target AWS Systems Manager Parameter Store only and use it as the system of record for the master key. Because Parameter Store stores strings rather than raw binary payloads, the design must define a stable serialization format for the 32-byte master key while preserving the existing encryption and file-format behavior.

## Goals / Non-Goals

**Goals:**

- Reuse the reference tool's package boundaries and operator flow so implementation stays familiar and testable
- Support encrypt, decrypt, create master, and delete master workflows backed by AWS Systems Manager Parameter Store
- Preserve AES-256-GCM envelope encryption and ENVC file-format behavior from the reference implementation
- Keep configuration minimal and AWS-focused for personal use

**Non-Goals:**

- Supporting GCP Secret Manager or AWS Secrets Manager in the personal tool
- Implementing interactive prompts, key rotation, or master-key import/export commands in the initial release
- Adding deployment automation, release packaging, or cross-platform installer work in this change

## Decisions

### Mirror the reference CLI architecture

The implementation will reuse the same coarse-grained structure as `dotenv_cryption`: `cmd/` for Cobra commands, `internal/config` for `dotenv.yaml`, `internal/provider` for master-key access, `internal/crypto` for AES-256-GCM utilities, `internal/format` for ENVC marshaling, and `internal/app` for orchestration.

This is preferred over collapsing everything into a single package because the reference layout already separates command parsing, cloud access, and cryptographic behavior in a testable way.

### Keep a provider interface even with one backend

The personal tool will define a master-key provider abstraction and back it with a single AWS Parameter Store implementation.

Calling SSM directly from Cobra commands would reduce files in the short term, but it would make command tests heavier and tie encryption flows to AWS SDK details. A small provider boundary keeps the design extensible and matches the sibling repo's testing model.

### Store the master key as a base64-encoded SecureString parameter

Parameter Store stores strings, not arbitrary binary payloads. The provider will therefore:

- generate a 32-byte random master key
- base64-encode it for storage
- create the configured parameter as `SecureString`
- retrieve it with decryption enabled
- decode and validate it back to exactly 32 bytes before use

Base64 is chosen over hex because it is shorter while still unambiguous. Raw-byte storage is not available in Parameter Store.

### Use an AWS-only configuration schema

`dotenv.yaml` will drop the multi-cloud selector and require only the AWS fields needed by this tool:

- `aws.region`
- `aws.parameter_name`

It will also keep the reference tool's crypto and file settings:

- `crypto.algorithm` with default `aes-256-gcm`
- `files.encrypted_prefix` for output naming customization

Retaining the old `cloud` selector and unused GCP configuration would increase complexity without adding value in a personal-only repository.

### Keep crypto and file-format compatibility with the reference tool

The tool will preserve the ENVC binary envelope format and the existing envelope-encryption flow:

- generate a random data key per file
- encrypt plaintext with AES-256-GCM
- wrap the data key with the master key using AES-256-GCM
- write nonce, wrapped key, and ciphertext into ENVC format

This keeps the implementation aligned with the proven reference design and avoids inventing a new ciphertext format for the personal variant.

## Risks / Trade-offs

- [Binary name collision with the team CLI] -> Acceptable for this change because the team and personal variants are not expected to be installed side by side
- [String serialization bugs in Parameter Store] -> Decode on every read and reject values that are not valid base64 or not exactly 32 bytes
- [AWS credential and IAM failures] -> Wrap SDK errors with actionable guidance that points users to `aws configure`, `AWS_PROFILE`, and missing `ssm:GetParameter` / `ssm:PutParameter` / `ssm:DeleteParameter` permissions
- [Assumed compatibility with the existing tool] -> Document that format and crypto are compatible, but cross-tool decryption requires the same underlying 32-byte master key material

## Migration Plan

This is a greenfield repository, so there is no in-place production migration or rollback requirement for existing code.

For users moving from the Secrets Manager-based tool, migration is operational rather than code-driven: either create a new Parameter Store-backed master key and re-encrypt files, or register the same 32-byte master key bytes as a base64-encoded SecureString parameter before switching tools.

## Open Questions

None. The binary name may remain the same as the team tool because both variants will not be installed side by side, and this change will rely on the default AWS-managed KMS path for Parameter Store `SecureString` values rather than supporting a custom KMS key ID.
