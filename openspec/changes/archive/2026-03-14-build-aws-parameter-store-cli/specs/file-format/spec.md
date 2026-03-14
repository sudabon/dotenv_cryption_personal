## ADDED Requirements

### Requirement: ENVC バイナリフォーマットの書き込み

Encrypted files SHALL be written with the following binary structure:

- MAGIC: 4 bytes (`ENVC`)
- VERSION: 1 byte (`0x01`)
- NONCE_LEN: 1 byte
- WRAPPED_KEY_LEN: 2 bytes in big-endian order
- NONCE: `NONCE_LEN` bytes
- WRAPPED_KEY: `WRAPPED_KEY_LEN` bytes
- CIPHERTEXT: all remaining bytes

#### Scenario: 暗号化ファイルを書き込む

- **WHEN** nonce, wrapped key, and ciphertext are available after encryption
- **THEN** the formatter writes a binary payload with the ENVC header and the provided sections in order

#### Scenario: バージョン番号を書き込む

- **WHEN** an encrypted file is written
- **THEN** the formatter stores `0x01` in the VERSION field

### Requirement: ENVC バイナリフォーマットの読み込み

The file reader SHALL parse the ENVC header and extract nonce, wrapped key, and ciphertext from the encrypted file.

#### Scenario: 正常なファイルを読み込む

- **WHEN** a valid ENVC file is provided for decryption
- **THEN** the reader extracts nonce, wrapped key, and ciphertext correctly

#### Scenario: マジックバイトが不正

- **WHEN** the first 4 bytes are not `ENVC`
- **THEN** the reader returns `invalid file format: missing ENVC header`

#### Scenario: サポート外のバージョン

- **WHEN** the VERSION field is not `0x01`
- **THEN** the reader returns `unsupported file format version`

#### Scenario: ヘッダーと本文の長さが不整合

- **WHEN** the declared nonce or wrapped-key length exceeds the remaining file data
- **THEN** the reader returns `corrupted file: unexpected end of data`
