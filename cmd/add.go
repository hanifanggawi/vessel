/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/hanifanggawi/vessel/internal/hosts"
	"github.com/spf13/cobra"
)

// errRuleExists signals a conflict whose explanation has already been printed,
// so the top-level handler should not print it again.
var errRuleExists = errors.New("rule already exists")

var (
	addWindows []string
	addFor     time.Duration
	addUntil   string
	addReplace bool
	addHidden  bool
)

func parseWindow(s string) (config.TimeWindow, error) {
	parts := strings.SplitN(s, "-", 2)
	if len(parts) != 2 {
		return config.TimeWindow{}, fmt.Errorf("invalid window %q: expected HH:MM-HH:MM", s)
	}
	return config.TimeWindow{
		StartTime: strings.TrimSpace(parts[0]),
		EndTime:   strings.TrimSpace(parts[1]),
	}, nil
}

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add [domain]",
	Short: "Add a domain to restrict",
	Long: `Add a domain to restrict, e.g instagram.com

The rule type is inferred from the flags you pass:

  vessel add instagram.com                          permanent block
  vessel add instagram.com --window 09:00-17:00     scheduled (repeatable)
  vessel add instagram.com --for 2h                 timer (duration from now)
  vessel add instagram.com --until 18:30            timer (until a time today)

Adding more --window values to an existing scheduled rule merges them in.
For any other change to an existing rule, pass --replace to overwrite it.

Pass --hidden to block the domain as normal while omitting it from
'vessel list'.`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			err := fmt.Errorf("missing domain argument\nUsage: vessel add [domain] [--window HH:MM-HH:MM ...]")
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		if len(args) > 1 {
			err := fmt.Errorf("unexpected arguments: %s\nRun 'vessel add --help' for usage", strings.Join(args[1:], " "))
			fmt.Fprintln(cmd.ErrOrStderr(), err)
			return err
		}
		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := runAdd(args[0])
		if err != nil && !errors.Is(err, errRuleExists) {
			fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		}
		return err
	},
}

func runAdd(domain string) error {

	// Fail early if the hosts file writing permission is denied
	if err := hosts.CheckWritable(); err != nil {
		return err
	}

	windows := make([]config.TimeWindow, 0, len(addWindows))
	for _, w := range addWindows {
		tw, err := parseWindow(w)
		if err != nil {
			return err
		}
		windows = append(windows, tw)
	}

	rule, err := config.BuildRule(config.RuleSpec{
		Domain:  domain,
		Windows: windows,
		For:     addFor,
		Until:   addUntil,
		Hidden:  addHidden,
	})
	if err != nil {
		return err
	}

	if err := config.RunReconcile(); err != nil {
		return err
	}

	result, err := config.AppendRule(rule, addReplace)
	if err != nil {
		return err
	}

	switch result.Outcome {
	case config.OutcomeAdded:
		fmt.Printf("Added rule: %s\n", result.Rule.HumanStr())
	case config.OutcomeReplaced:
		fmt.Printf("Replaced rule for %q\n", domain)
		fmt.Printf("  was: %s\n", result.Existing.HumanStr())
		fmt.Printf("  now: %s\n", result.Rule.HumanStr())
	case config.OutcomeWindowsMerged:
		if result.AddedWindows == 0 {
			fmt.Printf("No change: those window(s) are already set for %q\n", domain)
		} else {
			fmt.Printf("Added %d window(s) to %q\n", result.AddedWindows, domain)
			fmt.Printf("  now: %s\n", result.Rule.HumanStr())
		}
	case config.OutcomeConflict:
		fmt.Printf("A rule for %q already exists:\n", domain)
		fmt.Printf("  %s\n", result.Existing.HumanStr())
		fmt.Printf("Re-run with --replace to overwrite it.\n")
		return errRuleExists
	}

	return config.RunReconcile()
}

func init() {
	rootCmd.AddCommand(addCmd)

	addCmd.Flags().StringArrayVarP(&addWindows, "window", "w", nil, "blocked time window HH:MM-HH:MM (repeatable; implies scheduled)")
	addCmd.Flags().DurationVar(&addFor, "for", 0, "block for a duration from now, e.g. 2h30m (implies timer)")
	addCmd.Flags().StringVar(&addUntil, "until", "", "block until a time of day HH:MM (implies timer)")
	addCmd.Flags().BoolVar(&addReplace, "replace", false, "overwrite an existing rule for the domain")
	addCmd.Flags().BoolVar(&addHidden, "hidden", false, "block the domain but hide it from 'vessel list'")

	addCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		fmt.Fprintln(cmd.ErrOrStderr(), err.Error())
		return err
	})
}
