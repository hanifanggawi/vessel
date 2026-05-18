/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/hanifanggawi/vessel/internal/daemon"
	"github.com/spf13/cobra"
)

// daemonCmd represents the daemon command
var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("daemon called")
	},
}

var daemonStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the daemon in the background",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Start()
	},
}

var daemonStopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the running daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Stop()
	},
}

var daemonStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show whether the daemon is running",
	RunE: func(cmd *cobra.Command, args []string) error {
		pid, running, err := daemon.Status()
		if err != nil {
			return err
		}
		if running {
			fmt.Printf("daemon is running (PID %d)\n", pid)
		} else {
			fmt.Println("daemon is not running")
		}
		return nil
	},
}

var daemonRestartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Restart()
	},
}

// Hidden — called internally when the binary re-execs itself
var daemonRunCmd = &cobra.Command{
	Use:    "run",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Run()
	},
}

var daemonInstallCmd = &cobra.Command{
	Use:           "install",
	Short:         "Install the daemon as a system service that runs at startup",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Install()
	},
}

var daemonUninstallCmd = &cobra.Command{
	Use:           "uninstall",
	Short:         "Remove the installed startup service",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		return daemon.Uninstall()
	},
}

func init() {
	daemonCmd.AddCommand(daemonRunCmd, daemonStartCmd, daemonStopCmd, daemonStatusCmd, daemonRestartCmd, daemonInstallCmd, daemonUninstallCmd)
	rootCmd.AddCommand(daemonCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// daemonCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// daemonCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
