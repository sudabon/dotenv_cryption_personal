package provider

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

func TestParameterStoreProviderGetMasterKey(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			accessValue: base64.StdEncoding.EncodeToString(bytesOfLength(masterKeySize)),
		},
	}

	key, err := p.GetMasterKey()
	if err != nil {
		t.Fatalf("GetMasterKey returned error: %v", err)
	}
	if len(key) != masterKeySize {
		t.Fatalf("expected %d byte key, got %d", masterKeySize, len(key))
	}
}

func TestParameterStoreProviderReturnsNotFoundError(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			accessErr: &ssmtypes.ParameterNotFound{Message: awsString("missing")},
		},
	}

	_, err := p.GetMasterKey()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestParameterStoreProviderRejectsInvalidBase64(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			accessValue: "not-base64",
		},
	}

	_, err := p.GetMasterKey()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "failed to decode base64") {
		t.Fatalf("expected base64 decode error, got %v", err)
	}
}

func TestParameterStoreProviderRejectsInvalidDecodedLength(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			accessValue: base64.StdEncoding.EncodeToString([]byte("short")),
		},
	}

	_, err := p.GetMasterKey()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "invalid master key") {
		t.Fatalf("expected invalid master key error, got %v", err)
	}
}

func TestParameterStoreProviderReturnsAuthGuidance(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			accessErr: errors.New("AccessDeniedException: not authorized"),
		},
	}

	_, err := p.GetMasterKey()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "aws configure") {
		t.Fatalf("expected aws auth guidance, got %v", err)
	}
	if !strings.Contains(err.Error(), "ssm:GetParameter") {
		t.Fatalf("expected permission guidance, got %v", err)
	}
}

func TestParameterStoreProviderCreateMasterKey(t *testing.T) {
	t.Parallel()

	client := &mockParameterStoreClient{}
	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client:        client,
	}

	if err := p.CreateMasterKey(); err != nil {
		t.Fatalf("CreateMasterKey returned error: %v", err)
	}
	if client.createdParameterName != "/personal/envcrypt/master-key" {
		t.Fatalf("expected parameter name to be recorded, got %q", client.createdParameterName)
	}

	decoded, err := base64.StdEncoding.DecodeString(client.createdValue)
	if err != nil {
		t.Fatalf("expected base64-encoded value, got %v", err)
	}
	if len(decoded) != masterKeySize {
		t.Fatalf("expected %d byte key, got %d", masterKeySize, len(decoded))
	}
}

func TestParameterStoreProviderCreateMasterKeyReturnsAlreadyExists(t *testing.T) {
	t.Parallel()

	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client: &mockParameterStoreClient{
			createErr: &ssmtypes.ParameterAlreadyExists{Message: awsString("exists")},
		},
	}

	err := p.CreateMasterKey()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected already exists error, got %v", err)
	}
}

func TestParameterStoreProviderDeleteMasterKey(t *testing.T) {
	t.Parallel()

	client := &mockParameterStoreClient{}
	p := &ParameterStoreProvider{
		region:        "ap-northeast-1",
		parameterName: "/personal/envcrypt/master-key",
		client:        client,
	}

	if err := p.DeleteMasterKey(); err != nil {
		t.Fatalf("DeleteMasterKey returned error: %v", err)
	}
	if client.deletedParameterName != "/personal/envcrypt/master-key" {
		t.Fatalf("expected parameter name to be recorded, got %q", client.deletedParameterName)
	}
}

type mockParameterStoreClient struct {
	accessValue          string
	accessErr            error
	createErr            error
	deleteErr            error
	createdParameterName string
	createdValue         string
	deletedParameterName string
}

func (c *mockParameterStoreClient) GetParameter(context.Context, string) (string, error) {
	if c.accessErr != nil {
		return "", c.accessErr
	}
	return c.accessValue, nil
}

func (c *mockParameterStoreClient) PutSecureParameter(_ context.Context, name string, value string) error {
	c.createdParameterName = name
	c.createdValue = value
	if c.createErr != nil {
		return c.createErr
	}
	return nil
}

func (c *mockParameterStoreClient) DeleteParameter(_ context.Context, name string) error {
	c.deletedParameterName = name
	if c.deleteErr != nil {
		return c.deleteErr
	}
	return nil
}

func bytesOfLength(size int) []byte {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i + 1)
	}
	return data
}

func awsString(value string) *string {
	return &value
}
