package cmd

import (
	"bytes"
	"crypto/rand"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/sudabon/dotenv_cryption_personal/internal/config"
	"github.com/sudabon/dotenv_cryption_personal/internal/provider"
)

func TestEncryptDecryptRoundTripWithDefaultPaths(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	masterKey := mustRandomKey(t)
	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)
	writeTestFile(t, filepath.Join(dir, ".env"), "HELLO=world\n")

	root := newTestRootCmd(t, dir, masterKey, nil)
	root.SetArgs([]string{"encrypt"})

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	encryptedPath := filepath.Join(dir, ".env.enc")
	assertFileExists(t, encryptedPath)

	root = newTestRootCmd(t, dir, masterKey, nil)
	root.SetArgs([]string{"decrypt"})
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatalf("decrypt failed: %v", err)
	}

	assertFileContent(t, filepath.Join(dir, ".env"), "HELLO=world\n")
	if !strings.Contains(stdout.String(), ".env -> .env.enc") {
		t.Fatalf("expected encrypt output, got %q", stdout.String())
	}
	if !strings.Contains(stdout.String(), ".env.enc -> .env") {
		t.Fatalf("expected decrypt output, got %q", stdout.String())
	}
}

func TestEncryptUsesConfiguredPrefix(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	masterKey := mustRandomKey(t)
	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
files:
  encrypted_prefix: enc.
`)
	writeTestFile(t, filepath.Join(dir, ".env"), "HELLO=world\n")

	root := newTestRootCmd(t, dir, masterKey, nil)
	root.SetArgs([]string{"encrypt"})

	if err := root.Execute(); err != nil {
		t.Fatalf("encrypt failed: %v", err)
	}

	assertFileExists(t, filepath.Join(dir, "enc..env"))
}

func TestEncryptReturnsMissingInputFile(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	masterKey := mustRandomKey(t)
	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)

	root := newTestRootCmd(t, dir, masterKey, nil)
	root.SetArgs([]string{"encrypt"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected encrypt to fail")
	}
	if !strings.Contains(err.Error(), "input file") {
		t.Fatalf("expected missing input file error, got %v", err)
	}
}

func TestDecryptRejectsInvalidFormat(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	masterKey := mustRandomKey(t)
	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)
	writeTestFile(t, filepath.Join(dir, ".env.enc"), "plain text")

	root := newTestRootCmd(t, dir, masterKey, nil)
	root.SetArgs([]string{"decrypt"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected decrypt to fail")
	}
	if !strings.Contains(err.Error(), "invalid file format") {
		t.Fatalf("expected invalid format error, got %v", err)
	}
}

func TestCreateMasterCreatesConfiguredParameter(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)

	provider := &trackingProvider{}
	root := newManagedResourceRootCmd(t, dir, provider)
	root.SetArgs([]string{"create", "master"})

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatalf("create master failed: %v", err)
	}
	if !provider.createCalled {
		t.Fatal("expected CreateMasterKey to be called")
	}
	if !strings.Contains(stdout.String(), "created master parameter: /personal/envcrypt/master-key") {
		t.Fatalf("expected create output, got %q", stdout.String())
	}
}

func TestDeleteMasterDeletesConfiguredParameter(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)

	provider := &trackingProvider{}
	root := newManagedResourceRootCmd(t, dir, provider)
	root.SetArgs([]string{"delete", "master"})

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatalf("delete master failed: %v", err)
	}
	if !provider.deleteCalled {
		t.Fatal("expected DeleteMasterKey to be called")
	}
	if !strings.Contains(stdout.String(), "deleted master parameter: /personal/envcrypt/master-key") {
		t.Fatalf("expected delete output, got %q", stdout.String())
	}
}

func TestCreateMasterPropagatesProviderError(t *testing.T) {
	dir := t.TempDir()
	chdirForTest(t, dir)

	writeTestFile(t, filepath.Join(dir, "dotenv.yaml"), `aws:
  region: ap-northeast-1
  parameter_name: /personal/envcrypt/master-key
`)

	provider := &trackingProvider{createErr: errors.New("aws authentication failed: configure AWS credentials")}
	root := newManagedResourceRootCmd(t, dir, provider)
	root.SetArgs([]string{"create", "master"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected create master to fail")
	}
	if !strings.Contains(err.Error(), "aws authentication failed") {
		t.Fatalf("expected aws auth error, got %v", err)
	}
}

func TestVersionCommand(t *testing.T) {
	root := NewRootCmd(Dependencies{})
	root.SetArgs([]string{"version"})

	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stdout)

	if err := root.Execute(); err != nil {
		t.Fatalf("version failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "dev") {
		t.Fatalf("expected version output, got %q", stdout.String())
	}
}

func newTestRootCmd(t *testing.T, dir string, masterKey []byte, providerErr error) *cobra.Command {
	t.Helper()

	loadConfig := func() (config.Config, error) {
		return config.LoadFromPath(filepath.Join(dir, "dotenv.yaml"))
	}

	providerFactory := func(config.Config) (provider.MasterKeyProvider, error) {
		if providerErr != nil {
			return nil, providerErr
		}
		return staticProvider{key: masterKey}, nil
	}

	return NewRootCmd(Dependencies{
		LoadConfig:      loadConfig,
		ProviderFactory: providerFactory,
	})
}

func newManagedResourceRootCmd(t *testing.T, dir string, masterKeyProvider provider.MasterKeyProvider) *cobra.Command {
	t.Helper()

	loadConfig := func() (config.Config, error) {
		return config.LoadFromPath(filepath.Join(dir, "dotenv.yaml"))
	}

	providerFactory := func(config.Config) (provider.MasterKeyProvider, error) {
		return masterKeyProvider, nil
	}

	return NewRootCmd(Dependencies{
		LoadConfig:      loadConfig,
		ProviderFactory: providerFactory,
	})
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

type trackingProvider struct {
	createCalled bool
	deleteCalled bool
	createErr    error
	deleteErr    error
}

func (p *trackingProvider) GetMasterKey() ([]byte, error) {
	return nil, nil
}

func (p *trackingProvider) CreateMasterKey() error {
	p.createCalled = true
	return p.createErr
}

func (p *trackingProvider) DeleteMasterKey() error {
	p.deleteCalled = true
	return p.deleteErr
}

func mustRandomKey(t *testing.T) []byte {
	t.Helper()

	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatalf("rand.Read: %v", err)
	}

	return key
}
