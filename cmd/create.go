package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateCmd(deps Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create envcrypt managed resources",
	}

	cmd.AddCommand(newCreateMasterCmd(deps))

	return cmd
}

func newCreateMasterCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "master",
		Short: "Create the configured master parameter",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := deps.LoadConfig()
			if err != nil {
				return err
			}

			masterKeyProvider, err := deps.ProviderFactory(cfg)
			if err != nil {
				return err
			}

			if err := masterKeyProvider.CreateMasterKey(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "created master parameter: %s\n", parameterName(cfg))
			return nil
		},
	}
}
