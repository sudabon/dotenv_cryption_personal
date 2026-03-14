## ADDED Requirements

### Requirement: AES-256-GCM によるデータ暗号化

The encryption engine SHALL use AES-256-GCM to encrypt plaintext data. The nonce SHALL be 12 bytes of cryptographically secure randomness.

#### Scenario: 正常な暗号化

- **WHEN** plaintext data and a valid 32-byte data key are provided
- **THEN** the engine generates a 12-byte random nonce and returns ciphertext authenticated with AES-256-GCM

#### Scenario: 不正なキー長での暗号化

- **WHEN** encryption is attempted with a key that is not 32 bytes long
- **THEN** the engine returns an error

### Requirement: AES-256-GCM によるデータ復号

The encryption engine SHALL use AES-256-GCM to decrypt ciphertext using the nonce and data key that were used for encryption.

#### Scenario: 正常な復号

- **WHEN** a valid nonce, ciphertext, and matching 32-byte data key are provided
- **THEN** the engine returns the original plaintext

#### Scenario: 不正なキーでの復号

- **WHEN** decryption is attempted with a different key than the one used for encryption
- **THEN** the engine returns an authentication error

#### Scenario: 改ざんされた暗号文の復号

- **WHEN** any byte in the authenticated ciphertext is modified before decryption
- **THEN** the engine returns an authentication error

### Requirement: データキーの生成とラップ

The system SHALL generate a new random 32-byte data key for each encryption operation and SHALL wrap and unwrap that data key with the 32-byte master key using AES-256-GCM.

#### Scenario: データキーを生成してラップする

- **WHEN** a file encryption operation starts with a valid 32-byte master key
- **THEN** the system generates a fresh 32-byte data key and produces a wrapped-key payload

#### Scenario: 正常にアンラップする

- **WHEN** a wrapped key and the matching 32-byte master key are provided
- **THEN** the system restores the original 32-byte data key

#### Scenario: 不正なマスター鍵でアンラップする

- **WHEN** a wrapped key is unwrapped with a different master key
- **THEN** the system returns an authentication error
