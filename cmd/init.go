package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/spf13/cobra"
)

var initForce bool

var initCmd = &cobra.Command{
	Use:           "init",
	Short:         "Create a domainconfig.toml at the config path",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		path := config.DomainsConfigPath

		if _, err := os.Stat(path); err == nil && !initForce {
			fmt.Fprintf(cmd.ErrOrStderr(), "config already exists at %s\nRe-run with --force to overwrite it.\n", path)
			return fmt.Errorf("config already exists")
		}

		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		if err := os.WriteFile(path, []byte("[rules]\n"), 0644); err != nil {
			return fmt.Errorf("could not write config: %w", err)
		}

		fmt.Printf("Created %s\n", path)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&initForce, "force", false, "overwrite an existing config file")
}
