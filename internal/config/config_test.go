package config

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromPathAWS(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "dotenv.yaml")
	writeTestFile(t, path, `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("LoadFromPath returned error: %v", err)
	}

	if cfg.AWS.Region != "ap-northeast-1" {
		t.Fatalf("expected region, got %q", cfg.AWS.Region)
	}
	if cfg.AWS.ParameterName != "/personal/envcrypt/master-key" {
		t.Fatalf("expected parameter name, got %q", cfg.AWS.ParameterName)
	}
	if cfg.Crypto.Algorithm != AlgorithmAES256GCM {
		t.Fatalf("expected default algorithm %q, got %q", AlgorithmAES256GCM, cfg.Crypto.Algorithm)
	}
}

func TestLoadFromPathMissingFile(t *testing.T) {
	t.Parallel()

	_, err := LoadFromPath(filepath.Join(t.TempDir(), "dotenv.yaml"))
	if !errors.Is(err, ErrConfigNotFound) {
		t.Fatalf("expected ErrConfigNotFound, got %v", err)
	}
}

func TestValidateRejectsMissingAWSFields(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "dotenv.yaml")
	writeTestFile(t, path, `aws:
  parameter_name: /personal/envcrypt/master-key
`)

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "region") {
		t.Fatalf("expected missing region error, got %v", err)
	}
}

func TestValidateRejectsUnsupportedAlgorithm(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "dotenv.yaml")
	writeTestFile(t, path, `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
crypto:
  algorithm: chacha20
`)

	_, err := LoadFromPath(path)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unsupported crypto algorithm") {
		t.Fatalf("expected unsupported algorithm error, got %v", err)
	}
}

func TestPathHelpers(t *testing.T) {
	t.Parallel()

	cfg := Config{}
	if got := cfg.EncryptedPath("/tmp/.env"); got != "/tmp/.env.enc" {
		t.Fatalf("expected suffix path, got %q", got)
	}

	cfg.Files.EncryptedPrefix = "enc."
	if got := cfg.EncryptedPath("/tmp/.env"); got != "/tmp/enc..env" {
		t.Fatalf("expected prefixed path, got %q", got)
	}

	decrypted, err := cfg.DecryptedPath("/tmp/enc..env")
	if err != nil {
		t.Fatalf("DecryptedPath returned error: %v", err)
	}
	if decrypted != "/tmp/.env" {
		t.Fatalf("expected decrypted path /tmp/.env, got %q", decrypted)
	}
}

func TestDecryptedPathRejectsUnknownPattern(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Files: FilesConfig{
			EncryptedPrefix: "enc.",
		},
	}

	_, err := cfg.DecryptedPath("/tmp/secret.bin")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "cannot derive output path") {
		t.Fatalf("expected derive path error, got %v", err)
	}
}
