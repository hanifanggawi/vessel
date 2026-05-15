/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/hanifanggawi/vessel/internal/config"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		rules, err := config.LoadConfig(config.DomainsConfigPath)
		if err != nil {
			fmt.Println(err.Error())
			return
		}
		if len(rules) == 0 {
			fmt.Println("No rules configured.")
			return
		}
		fmt.Printf("%-40s %-12s %s\n", "DOMAIN", "KIND", "DETAILS")
		fmt.Println(strings.Repeat("-", 72))
		for _, r := range rules {
			details := ruleDetails(r)
			fmt.Printf("%-40s %-12s %s\n", r.Domain, string(r.Kind), details)
		}
	},
}

func ruleDetails(r config.DomainRule) string {
	switch r.Kind {
	case config.RuleTypeTimer:
		if !r.BlockedUntil.IsZero() {
			return "until " + r.BlockedUntil.Format(time.TimeOnly)
		}
	case config.RuleTypeScheduled:
		windows := make([]string, len(r.BlockedWindows))
		for i, w := range r.BlockedWindows {
			windows[i] = w.StartTime + "-" + w.EndTime
		}
		return strings.Join(windows, ", ")
	}
	return ""
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
