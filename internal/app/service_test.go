package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sudabon/dotenv_cryption_personal/internal/config"
	"github.com/sudabon/dotenv_cryption_personal/internal/crypto"
)

func TestServiceEncryptDecryptRoundTrip(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	inputPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(inputPath, []byte("HELLO=world\n"), 0o600); err != nil {
		t.Fatalf("os.WriteFile(%q): %v", inputPath, err)
	}

	cfg := config.Config{
		AWS: config.AWSConfig{
			Region:        "ap-northeast-1",
			ParameterName: "/personal/envcrypt/master-key",
		},
		Crypto: config.CryptoConfig{
			Algorithm: config.AlgorithmAES256GCM,
		},
	}

	service := NewService(staticProvider{key: bytes.Repeat([]byte{7}, crypto.DataKeySize)})
	encryptedPath, err := service.EncryptFile(inputPath, cfg)
	if err != nil {
		t.Fatalf("EncryptFile returned error: %v", err)
	}

	decryptedPath, err := service.DecryptFile(encryptedPath, cfg)
	if err != nil {
		t.Fatalf("DecryptFile returned error: %v", err)
	}

	data, err := os.ReadFile(decryptedPath)
	if err != nil {
		t.Fatalf("os.ReadFile(%q): %v", decryptedPath, err)
	}
	if string(data) != "HELLO=world\n" {
		t.Fatalf("expected original plaintext, got %q", string(data))
	}
}

func TestEncryptFileMissingInput(t *testing.T) {
	t.Parallel()

	service := NewService(staticProvider{key: bytes.Repeat([]byte{7}, crypto.DataKeySize)})

	_, err := service.EncryptFile("missing.env", config.Config{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

type staticProvider struct {
	key []byte
}

func (p staticProvider) GetMasterKey() ([]byte, error) {
	return append([]byte(nil), p.key...), nil
}

func (staticProvider) CreateMasterKey() error {
	return nil
}

func (staticProvider) DeleteMasterKey() error {
	return nil
}
