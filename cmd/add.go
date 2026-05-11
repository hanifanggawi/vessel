/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/hanifanggawi/vessel/internal/daemon"
	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [domain]",
	Short: "Add a domain to restrict",
	Long:  `Add a domain to restrict, e.g instagram.com`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		domain := args[0]
		domainRule := config.DomainRule{
			Domain:  domain,
			AddedAt: time.Now(),
			Kind:    "block",
		}
		rules, err := config.AppendRule([]config.DomainRule{domainRule})
		if err != nil {
			fmt.Println(err.Error())
		}
		for _, rule := range rules {
			fmt.Println(rule)
		}
		err = daemon.RunReconcile()
		if err != nil {
			fmt.Println(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
