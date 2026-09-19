package cmd

import (
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/selfupdate"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumanUpgradeCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumanUpgradeCmdDesc",
		Other: "Upgrade yuman to the latest release",
	},
}

// yumanUpgradeCmd represents the upgrade command
var yumanUpgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: tr.Localize(&MsgYumanUpgradeCmdShort),
	Args:  cobra.NoArgs,
	RunE:  runYumanUpgrade,
}

func init() {
	rootCmd.AddCommand(yumanUpgradeCmd)
}

func runYumanUpgrade(cmd *cobra.Command, args []string) error {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	return selfupdate.Update(cmd.Context(), flagVerbose)
}
