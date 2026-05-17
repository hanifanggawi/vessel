/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release [domain]",
	Short: "Release a domain by removing its rule",
	Long: `Release a domain by removing its rule from the configuration
and reconciling the resulting state. For example:

  vessel release example.com`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		err := config.RunReconcile()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		releasedRule, err := config.ReleaseRule(domain)
		if err != nil {
			fmt.Println(err.Error())

		} else {
			fmt.Printf("Released `%s`\n", releasedRule)
		}
		err = config.RunReconcile()
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(releaseCmd)
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// releaseCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// releaseCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
