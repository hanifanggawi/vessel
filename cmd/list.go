/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/hanifanggawi/vessel/internal/config"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured rules",
	Long: `List the rules currently configured, showing each domain along with
its restriction type: permanent, scheduled time windows, or a timer.`,
	Run: func(cmd *cobra.Command, args []string) {
		listStr, err := config.DomainConfigListStr()
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		fmt.Print(listStr)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
