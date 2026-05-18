/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vessel",
	Short: "Block and limit website usage via hosts-file rules",
	Long: `Vessel is a CLI for blocking and limiting website usage.

It manages rules in a config file and reflects them onto the system's
hosts file. A background daemon continuously reads and syncs the hosts
file so rules stay enforced. Seal the vessel to make rules harder to
remove by requiring a challenge before releasing or unsealing them.

Editing the hosts file requires elevation: run write commands with
sudo on Linux or from an Administrator prompt on Windows.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// SetVersion wires the build-time version into the root command so
// `vessel --version` reports it.
func SetVersion(v string) {
	rootCmd.Version = v
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
