# config-management Specification

## Purpose
TBD - created by archiving change build-aws-parameter-store-cli. Update Purpose after archive.
## Requirements
### Requirement: dotenv.yaml の読み込み

The CLI SHALL load `dotenv.yaml` from the current working directory at startup before executing commands that require configuration.

#### Scenario: 設定ファイルが存在する

- **WHEN** the user runs a command in a directory that contains a valid `dotenv.yaml`
- **THEN** the CLI loads AWS, crypto, and file-naming settings from that file

#### Scenario: 設定ファイルが存在しない

- **WHEN** the user runs a command that requires configuration and `dotenv.yaml` is missing
- **THEN** the CLI reports `dotenv.yaml not found` and exits with status code 1

### Requirement: AWS 設定のバリデーション

The configuration loader SHALL require `aws.region` and `aws.parameter_name`, default `crypto.algorithm` to `aes-256-gcm` when omitted, and reject unsupported algorithm values.

#### Scenario: AWS の必須フィールドが不足している

- **WHEN** `dotenv.yaml` omits either `aws.region` or `aws.parameter_name`
- **THEN** the CLI reports the missing field names and exits with status code 1

#### Scenario: 暗号アルゴリズムのデフォルトを適用する

- **WHEN** `dotenv.yaml` omits `crypto.algorithm`
- **THEN** the CLI uses `aes-256-gcm` as the effective algorithm

#### Scenario: サポート外の暗号アルゴリズム

- **WHEN** `dotenv.yaml` sets `crypto.algorithm` to a value other than `aes-256-gcm`
- **THEN** the CLI reports an unsupported algorithm error and exits with status code 1

### Requirement: 出力ファイル名の導出

The configuration layer SHALL derive encrypted and decrypted file paths from the input file path and `files.encrypted_prefix`.

#### Scenario: デフォルトの暗号化出力パス

- **WHEN** `files.encrypted_prefix` is unset and the user encrypts `.env`
- **THEN** the output path is `.env.enc`

#### Scenario: カスタムプレフィックスの暗号化出力パス

- **WHEN** `files.encrypted_prefix` is set to `enc.` and the user encrypts `.env`
- **THEN** the output path is `enc..env`

#### Scenario: 復号出力パスを導出できない

- **WHEN** the user decrypts a file whose name matches neither the configured prefix form nor the default `.enc` suffix form
- **THEN** the CLI reports that it cannot derive the decrypted output path

