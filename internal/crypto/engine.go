package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	DataKeySize = 32
	NonceSize   = 12
)

func Encrypt(plaintext, key []byte) ([]byte, []byte, error) {
	if err := validateKeyLength(key); err != nil {
		return nil, nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to initialize gcm: %w", err)
	}

	nonce := make([]byte, NonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, nil, fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := aead.Seal(nil, nonce, plaintext, nil)
	return nonce, ciphertext, nil
}

func Decrypt(nonce, ciphertext, key []byte) ([]byte, error) {
	if err := validateKeyLength(key); err != nil {
		return nil, err
	}
	if len(nonce) != NonceSize {
		return nil, fmt.Errorf("invalid nonce length: expected %d bytes", NonceSize)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cipher: %w", err)
	}

	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize gcm: %w", err)
	}

	plaintext, err := aead.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("gcm authentication failed: %w", err)
	}

	return plaintext, nil
}

func GenerateDataKey() ([]byte, error) {
	return generateRandomKey(DataKeySize, "data key")
}

func GenerateMasterKey() ([]byte, error) {
	return generateRandomKey(DataKeySize, "master key")
}

func WrapKey(dataKey, masterKey []byte) ([]byte, error) {
	nonce, ciphertext, err := Encrypt(dataKey, masterKey)
	if err != nil {
		return nil, fmt.Errorf("failed to wrap data key: %w", err)
	}

	return append(nonce, ciphertext...), nil
}

func UnwrapKey(wrappedKey, masterKey []byte) ([]byte, error) {
	if len(wrappedKey) <= NonceSize {
		return nil, errors.New("invalid wrapped key")
	}

	return Decrypt(wrappedKey[:NonceSize], wrappedKey[NonceSize:], masterKey)
}

func validateKeyLength(key []byte) error {
	if len(key) != DataKeySize {
		return fmt.Errorf("invalid key length: expected %d bytes", DataKeySize)
	}

	return nil
}

func generateRandomKey(size int, label string) ([]byte, error) {
	key := make([]byte, size)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, fmt.Errorf("failed to generate %s: %w", label, err)
	}

	return key, nil
}
