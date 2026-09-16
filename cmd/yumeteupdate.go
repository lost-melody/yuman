package cmd

import (
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yumete"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeteUpdateCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeteUpdateCmdDesc",
		Other: "Update yumete to the latest release",
	},
}

// yumeteUpdateCmd represents the update command
var yumeteUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: tr.Localize(&MsgYumeteUpdateCmdShort),
	Args:  cobra.NoArgs,
	RunE:  runYumeteUpdate,
}

func init() {
	yumeteCmd.AddCommand(yumeteUpdateCmd)
}

func runYumeteUpdate(cmd *cobra.Command, args []string) error {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	return yumete.Update(cmd.Context(), flagVerbose)
}
