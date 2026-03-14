## 1. Project Setup

- [x] 1.1 Initialize the Go module, dependency set, and CLI entrypoint structure using the reference repository as the baseline
- [x] 1.2 Implement `dotenv.yaml` loading and validation for `aws.region`, `aws.parameter_name`, crypto defaults, and file path helpers
- [x] 1.3 Add an AWS Parameter Store sample configuration and supporting project scaffolding files needed to exercise the CLI locally

## 2. Crypto And File Processing

- [x] 2.1 Implement AES-256-GCM helpers for plaintext encryption/decryption and data-key generation, wrapping, and unwrapping
- [x] 2.2 Implement ENVC file-format marshal/unmarshal helpers with header, version, and length validation
- [x] 2.3 Build the application service that composes config, provider, crypto, and format logic into encrypt/decrypt file workflows

## 3. Parameter Store And Commands

- [x] 3.1 Implement the master-key provider interface and AWS Systems Manager client wrapper for get/create/delete using base64-encoded `SecureString` values
- [x] 3.2 Add Cobra commands for `encrypt`, `decrypt`, `create master`, and `delete master`, wiring config loading and provider/service factories
- [x] 3.3 Add unit tests for config, crypto/format, provider, service, and command flows, including missing-file and AWS error cases

## 4. Documentation And Verification

- [x] 4.1 Update the README with setup, IAM/authentication guidance, command usage, and migration notes from the Secrets Manager-based tool
- [x] 4.2 Run formatting and verification commands, fix any failures, and leave the change ready for implementation review
