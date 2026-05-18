package cmd

import (
	"github.com/hanifanggawi/vessel/internal/challenge"
	"github.com/hanifanggawi/vessel/internal/config"
	"github.com/spf13/cobra"
)

// runReleaseChallenge presents the release challenge wired to the command's
// in/out streams. Swapping the challenge type later is a single change here.
func runReleaseChallenge(cmd *cobra.Command) error {
	return challenge.Run(challenge.NewBackwards(), cmd.InOrStdin(), cmd.OutOrStdout())
}

var sealCmd = &cobra.Command{
	Use:   "seal",
	Short: "Seal the vessel so releasing a domain requires passing a challenge",
	Long: `Seal the vessel. While sealed, 'vessel release' and 'vessel unseal'
require passing a challenge before they take effect.`,
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sealed, err := config.IsSealed()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		if sealed {
			cmd.Println("Vessel is already sealed.")
			return nil
		}
		if err := config.SetSealed(true); err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("Vessel sealed. Releasing a domain now requires passing a challenge.")
		return nil
	},
}

var unsealCmd = &cobra.Command{
	Use:           "unseal",
	Short:         "Unseal the vessel (requires passing a challenge while sealed)",
	Args:          cobra.NoArgs,
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		sealed, err := config.IsSealed()
		if err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		if !sealed {
			cmd.Println("Vessel is already unsealed.")
			return nil
		}
		if err := runReleaseChallenge(cmd); err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		if err := config.SetSealed(false); err != nil {
			cmd.PrintErrln(err.Error())
			return err
		}
		cmd.Println("Vessel unsealed.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sealCmd)
	rootCmd.AddCommand(unsealCmd)
}
