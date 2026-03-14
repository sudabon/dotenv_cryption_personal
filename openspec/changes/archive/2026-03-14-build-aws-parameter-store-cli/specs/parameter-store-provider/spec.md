## ADDED Requirements

### Requirement: マスター鍵の取得

The system SHALL retrieve the configured master key from AWS Systems Manager Parameter Store using decryption, base64-decode the stored value, and return exactly 32 bytes to the caller.

#### Scenario: 正常なキー取得

- **WHEN** valid AWS credentials are available and the configured parameter exists as a valid base64-encoded 32-byte key
- **THEN** the provider returns the decoded 32-byte master key

#### Scenario: パラメータが存在しない

- **WHEN** the configured parameter name does not exist in Parameter Store
- **THEN** the provider returns a not found error that includes the parameter name

#### Scenario: 不正な格納値

- **WHEN** the configured parameter value is not valid base64 or does not decode to exactly 32 bytes
- **THEN** the provider returns an invalid master key error

#### Scenario: 認証または認可の失敗

- **WHEN** AWS credentials are missing, expired, or do not permit `ssm:GetParameter`
- **THEN** the provider returns an authentication or authorization error with actionable AWS guidance

### Requirement: マスター鍵の作成

The system SHALL generate a new 32-byte random master key, base64-encode it, and create the configured Parameter Store entry as a `SecureString` without overwriting an existing parameter.

#### Scenario: 新しいマスター鍵を作成する

- **WHEN** the configured parameter name does not exist and the user requests master-key creation
- **THEN** the provider creates a `SecureString` parameter that contains a base64-encoded 32-byte key

#### Scenario: 既存パラメータがある

- **WHEN** the configured parameter name already exists
- **THEN** the provider fails without overwriting the parameter and reports that it already exists

#### Scenario: 認証または認可の失敗

- **WHEN** AWS credentials are missing, expired, or do not permit `ssm:PutParameter`
- **THEN** the provider returns an authentication or authorization error with actionable AWS guidance

### Requirement: マスター鍵の削除

The system SHALL delete the configured master-key parameter from AWS Systems Manager Parameter Store.

#### Scenario: 既存パラメータを削除する

- **WHEN** the configured parameter exists and the user requests deletion
- **THEN** the provider deletes that parameter

#### Scenario: 削除対象が存在しない

- **WHEN** the configured parameter does not exist
- **THEN** the provider returns a not found error that includes the parameter name

#### Scenario: 認証または認可の失敗

- **WHEN** AWS credentials are missing, expired, or do not permit `ssm:DeleteParameter`
- **THEN** the provider returns an authentication or authorization error with actionable AWS guidance
