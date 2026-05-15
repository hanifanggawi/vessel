package cmd

import (
	"fmt"
	"os"

	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/spf13/cobra"
)

var reconcileCmd = &cobra.Command{
	Use:    "reconcile",
	Short:  "Manually trigger a reconcile of the hosts file",
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.RunReconcile(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Reconcile complete.")
	},
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}
