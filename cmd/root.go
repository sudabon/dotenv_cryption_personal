package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sudabon/dotenv_cryption_personal/internal/config"
	"github.com/sudabon/dotenv_cryption_personal/internal/provider"
	"github.com/sudabon/dotenv_cryption_personal/pkg/version"
)

type Dependencies struct {
	LoadConfig      func() (config.Config, error)
	ProviderFactory func(config.Config) (provider.MasterKeyProvider, error)
}

func Execute() error {
	return NewRootCmd(Dependencies{}).Execute()
}

func NewRootCmd(deps Dependencies) *cobra.Command {
	if deps.LoadConfig == nil {
		deps.LoadConfig = config.Load
	}
	if deps.ProviderFactory == nil {
		deps.ProviderFactory = provider.New
	}

	rootCmd := &cobra.Command{
		Use:           "envcrypt",
		Short:         "Encrypt and decrypt .env files with AWS Parameter Store",
		Version:       version.Version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	rootCmd.AddCommand(newEncryptCmd(deps))
	rootCmd.AddCommand(newDecryptCmd(deps))
	rootCmd.AddCommand(newCreateCmd(deps))
	rootCmd.AddCommand(newDeleteCmd(deps))
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd
}

func parameterName(cfg config.Config) string {
	return cfg.AWS.ParameterName
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the envcrypt version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version.Version)
			return err
		},
	}
}
