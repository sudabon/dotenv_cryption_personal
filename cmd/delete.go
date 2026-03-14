package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDeleteCmd(deps Dependencies) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete envcrypt managed resources",
	}

	cmd.AddCommand(newDeleteMasterCmd(deps))

	return cmd
}

func newDeleteMasterCmd(deps Dependencies) *cobra.Command {
	return &cobra.Command{
		Use:   "master",
		Short: "Delete the configured master parameter",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := deps.LoadConfig()
			if err != nil {
				return err
			}

			masterKeyProvider, err := deps.ProviderFactory(cfg)
			if err != nil {
				return err
			}

			if err := masterKeyProvider.DeleteMasterKey(); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "deleted master parameter: %s\n", parameterName(cfg))
			return nil
		},
	}
}
