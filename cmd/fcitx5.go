package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5CmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5CmdDesc",
		Other: "Manages fcitx5 service via DBus",
	},
}

// fcitx5Cmd represents the fcitx5 command
var fcitx5Cmd = &cobra.Command{
	Use:   "fcitx5",
	Short: tr.Localize(&MsgFcitx5CmdShort),
}

func init() {
	rootCmd.AddCommand(fcitx5Cmd)
}
