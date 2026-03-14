package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sudabon/dotenv_cryption_personal/internal/app"
)

func newEncryptCmd(deps Dependencies) *cobra.Command {
	var filePath string

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt a dotenv file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := deps.LoadConfig()
			if err != nil {
				return err
			}

			masterKeyProvider, err := deps.ProviderFactory(cfg)
			if err != nil {
				return err
			}

			service := app.NewService(masterKeyProvider)
			outputPath, err := service.EncryptFile(filePath, cfg)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s -> %s\n", filePath, outputPath)
			return nil
		},
	}

	cmd.Flags().StringVar(&filePath, "file", ".env", "path to the dotenv file")

	return cmd
}
