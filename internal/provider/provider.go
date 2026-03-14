package provider

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	ssmtypes "github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/sudabon/dotenv_cryption_personal/internal/config"
	cryptoutil "github.com/sudabon/dotenv_cryption_personal/internal/crypto"
)

const masterKeySize = cryptoutil.DataKeySize

type MasterKeyProvider interface {
	GetMasterKey() ([]byte, error)
	CreateMasterKey() error
	DeleteMasterKey() error
}

type ParameterStoreClient interface {
	GetParameter(ctx context.Context, name string) (string, error)
	PutSecureParameter(ctx context.Context, name string, value string) error
	DeleteParameter(ctx context.Context, name string) error
}

type ParameterStoreProvider struct {
	region        string
	parameterName string
	client        ParameterStoreClient
}

func New(cfg config.Config) (MasterKeyProvider, error) {
	return newParameterStoreProvider(cfg.AWS)
}

func newParameterStoreProvider(cfg config.AWSConfig) (MasterKeyProvider, error) {
	client, err := newParameterStoreClient(context.Background(), cfg.Region)
	if err != nil {
		return nil, wrapAWSClientError(err)
	}

	return &ParameterStoreProvider{
		region:        cfg.Region,
		parameterName: cfg.ParameterName,
		client:        client,
	}, nil
}

func (p *ParameterStoreProvider) GetMasterKey() ([]byte, error) {
	value, err := p.client.GetParameter(context.Background(), p.parameterName)
	if err != nil {
		return nil, wrapAWSError(p.parameterName, "retrieve", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(value))
	if err != nil {
		return nil, fmt.Errorf("invalid master key in aws parameter %q: failed to decode base64: %w", p.parameterName, err)
	}
	if err := validateMasterKey(decoded, fmt.Sprintf("aws parameter %q", p.parameterName)); err != nil {
		return nil, err
	}

	return append([]byte(nil), decoded...), nil
}

func (p *ParameterStoreProvider) CreateMasterKey() error {
	key, err := cryptoutil.GenerateMasterKey()
	if err != nil {
		return err
	}

	encoded := base64.StdEncoding.EncodeToString(key)
	if err := p.client.PutSecureParameter(context.Background(), p.parameterName, encoded); err != nil {
		return wrapAWSError(p.parameterName, "create", err)
	}

	return nil
}

func (p *ParameterStoreProvider) DeleteMasterKey() error {
	if err := p.client.DeleteParameter(context.Background(), p.parameterName); err != nil {
		return wrapAWSError(p.parameterName, "delete", err)
	}

	return nil
}

type ssmSDKClient struct {
	client *ssm.Client
}

func newParameterStoreClient(ctx context.Context, region string) (ParameterStoreClient, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &ssmSDKClient{
		client: ssm.NewFromConfig(cfg),
	}, nil
}

func (c *ssmSDKClient) GetParameter(ctx context.Context, name string) (string, error) {
	withDecryption := true

	resp, err := c.client.GetParameter(ctx, &ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &withDecryption,
	})
	if err != nil {
		return "", err
	}
	if resp.Parameter == nil || resp.Parameter.Value == nil {
		return "", errors.New("parameter has no value")
	}

	return *resp.Parameter.Value, nil
}

func (c *ssmSDKClient) PutSecureParameter(ctx context.Context, name string, value string) error {
	overwrite := false

	_, err := c.client.PutParameter(ctx, &ssm.PutParameterInput{
		Name:      &name,
		Value:     &value,
		Type:      ssmtypes.ParameterTypeSecureString,
		Overwrite: &overwrite,
	})
	return err
}

func (c *ssmSDKClient) DeleteParameter(ctx context.Context, name string) error {
	_, err := c.client.DeleteParameter(ctx, &ssm.DeleteParameterInput{
		Name: &name,
	})
	return err
}

func validateMasterKey(key []byte, source string) error {
	if len(key) != masterKeySize {
		return fmt.Errorf("invalid master key from %s: expected %d bytes", source, masterKeySize)
	}

	return nil
}

func wrapAWSClientError(err error) error {
	message := strings.ToLower(err.Error())
	switch {
	case isAWSAuthLike(message):
		return fmt.Errorf("aws authentication failed: configure AWS credentials (for example `aws configure` or AWS_PROFILE) and confirm the target region is accessible: %w", err)
	default:
		return fmt.Errorf("failed to initialize aws systems manager client: %w", err)
	}
}

func wrapAWSError(parameterName, action string, err error) error {
	var notFound *ssmtypes.ParameterNotFound
	if errors.As(err, &notFound) {
		return fmt.Errorf("aws parameter %q not found: %w", parameterName, err)
	}

	var exists *ssmtypes.ParameterAlreadyExists
	if errors.As(err, &exists) {
		return fmt.Errorf("aws parameter %q already exists: %w", parameterName, err)
	}

	message := strings.ToLower(err.Error())
	switch {
	case strings.Contains(message, "already exists"):
		return fmt.Errorf("aws parameter %q already exists: %w", parameterName, err)
	case strings.Contains(message, "parameter not found"),
		strings.Contains(message, "not found"):
		return fmt.Errorf("aws parameter %q not found: %w", parameterName, err)
	case isAWSAuthLike(message):
		return fmt.Errorf("aws authentication failed: configure AWS credentials (for example `aws configure` or AWS_PROFILE) and ensure %s permission on %q: %w", requiredPermission(action), parameterName, err)
	default:
		switch action {
		case "create":
			return fmt.Errorf("failed to create master key in aws systems manager parameter store: %w", err)
		case "delete":
			return fmt.Errorf("failed to delete master key from aws systems manager parameter store: %w", err)
		default:
			return fmt.Errorf("failed to retrieve master key from aws systems manager parameter store: %w", err)
		}
	}
}

func isAWSAuthLike(message string) bool {
	return strings.Contains(message, "credential") ||
		strings.Contains(message, "accessdenied") ||
		strings.Contains(message, "access denied") ||
		strings.Contains(message, "expiredtoken") ||
		strings.Contains(message, "unauthorized") ||
		strings.Contains(message, "not authorized") ||
		strings.Contains(message, "signature")
}

func requiredPermission(action string) string {
	switch action {
	case "create":
		return "ssm:PutParameter"
	case "delete":
		return "ssm:DeleteParameter"
	default:
		return "ssm:GetParameter"
	}
}
