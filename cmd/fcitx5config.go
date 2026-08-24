package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5ConfigCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5ConfigCmdDesc",
		Other: "Manages fcitx5 config",
	},
}

// fcitx5ConfigCmd represents the config command
var fcitx5ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: tr.Localize(&MsgFcitx5ConfigCmdShort),
}

func init() {
	fcitx5Cmd.AddCommand(fcitx5ConfigCmd)
}
