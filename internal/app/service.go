package app

import (
	"errors"
	"fmt"
	"os"

	"github.com/sudabon/dotenv_cryption_personal/internal/config"
	cryptoutil "github.com/sudabon/dotenv_cryption_personal/internal/crypto"
	"github.com/sudabon/dotenv_cryption_personal/internal/format"
	"github.com/sudabon/dotenv_cryption_personal/internal/provider"
)

type Service struct {
	masterKeyProvider provider.MasterKeyProvider
}

func NewService(masterKeyProvider provider.MasterKeyProvider) *Service {
	return &Service{masterKeyProvider: masterKeyProvider}
}

func (s *Service) EncryptFile(inputPath string, cfg config.Config) (string, error) {
	plaintext, err := readFile(inputPath)
	if err != nil {
		return "", err
	}

	masterKey, err := s.masterKeyProvider.GetMasterKey()
	if err != nil {
		return "", err
	}

	dataKey, err := cryptoutil.GenerateDataKey()
	if err != nil {
		return "", fmt.Errorf("failed to generate data key: %w", err)
	}

	nonce, ciphertext, err := cryptoutil.Encrypt(plaintext, dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt file: %w", err)
	}

	wrappedKey, err := cryptoutil.WrapKey(dataKey, masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to wrap data key: %w", err)
	}

	payload, err := format.Marshal(format.Envelope{
		Nonce:      nonce,
		WrappedKey: wrappedKey,
		Ciphertext: ciphertext,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal encrypted file: %w", err)
	}

	outputPath := cfg.EncryptedPath(inputPath)
	if err := os.WriteFile(outputPath, payload, 0o600); err != nil {
		return "", fmt.Errorf("failed to write encrypted file: %w", err)
	}

	return outputPath, nil
}

func (s *Service) DecryptFile(inputPath string, cfg config.Config) (string, error) {
	payload, err := readFile(inputPath)
	if err != nil {
		return "", err
	}

	envelope, err := format.Unmarshal(payload)
	if err != nil {
		return "", err
	}

	masterKey, err := s.masterKeyProvider.GetMasterKey()
	if err != nil {
		return "", err
	}

	dataKey, err := cryptoutil.UnwrapKey(envelope.WrappedKey, masterKey)
	if err != nil {
		return "", fmt.Errorf("failed to unwrap data key: %w", err)
	}

	plaintext, err := cryptoutil.Decrypt(envelope.Nonce, envelope.Ciphertext, dataKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt file: %w", err)
	}

	outputPath, err := cfg.DecryptedPath(inputPath)
	if err != nil {
		return "", err
	}

	if err := os.WriteFile(outputPath, plaintext, 0o600); err != nil {
		return "", fmt.Errorf("failed to write decrypted file: %w", err)
	}

	return outputPath, nil
}

func readFile(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err == nil {
		return data, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("input file %q not found", path)
	}

	return nil, fmt.Errorf("failed to read %q: %w", path, err)
}
