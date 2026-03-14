# cli-commands Specification

## Purpose
TBD - created by archiving change build-aws-parameter-store-cli. Update Purpose after archive.
## Requirements
### Requirement: encrypt コマンド

The CLI SHALL provide an `encrypt` command that encrypts the selected dotenv file and writes an ENVC-formatted output file. If `--file` is omitted, the CLI SHALL use `.env` from the current working directory.

#### Scenario: ファイル指定で暗号化する

- **WHEN** the user runs the `encrypt` command with `--file .env.production`
- **THEN** the CLI reads that file, retrieves the configured master key, and writes the encrypted output to the configured derived path

#### Scenario: デフォルトファイルを暗号化する

- **WHEN** the user runs the `encrypt` command without `--file`
- **THEN** the CLI encrypts `.env` from the current working directory

#### Scenario: 入力ファイルが存在しない

- **WHEN** the selected plaintext file does not exist
- **THEN** the CLI reports that the input file was not found and exits with status code 1

### Requirement: decrypt コマンド

The CLI SHALL provide a `decrypt` command that reads an ENVC-formatted file, restores the original plaintext, and writes the derived output file. If `--file` is omitted, the CLI SHALL use `.env.enc` from the current working directory.

#### Scenario: ファイル指定で復号する

- **WHEN** the user runs the `decrypt` command with `--file .env.production.enc`
- **THEN** the CLI decrypts that file and writes the derived plaintext output path

#### Scenario: デフォルトファイルを復号する

- **WHEN** the user runs the `decrypt` command without `--file`
- **THEN** the CLI decrypts `.env.enc` from the current working directory

#### Scenario: 不正なフォーマットのファイル

- **WHEN** the user attempts to decrypt a file that does not contain a valid ENVC payload
- **THEN** the CLI reports a file-format error and exits with status code 1

### Requirement: create master コマンド

The CLI SHALL provide a `create master` command that creates a new master key in the configured AWS Systems Manager Parameter Store location.

#### Scenario: マスター鍵を作成する

- **WHEN** the configured parameter does not exist and the user runs `create master`
- **THEN** the CLI creates a new master key via the Parameter Store provider

#### Scenario: 既存パラメータがある

- **WHEN** the configured parameter already exists and the user runs `create master`
- **THEN** the CLI reports that the parameter already exists and exits with status code 1

### Requirement: delete master コマンド

The CLI SHALL provide a `delete master` command that removes the configured master-key parameter from AWS Systems Manager Parameter Store.

#### Scenario: マスター鍵を削除する

- **WHEN** the configured parameter exists and the user runs `delete master`
- **THEN** the CLI deletes that parameter via the Parameter Store provider

#### Scenario: 削除対象が存在しない

- **WHEN** the configured parameter does not exist and the user runs `delete master`
- **THEN** the CLI reports a not found error and exits with status code 1

