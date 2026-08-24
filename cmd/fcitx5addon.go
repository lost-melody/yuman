package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5AddonCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5AddonCmdDesc",
		Other: "Manages fcitx5 addons",
	},
}

// fcitx5AddonCmd represents the addon command
var fcitx5AddonCmd = &cobra.Command{
	Use:   "addon",
	Short: tr.Localize(&MsgFcitx5AddonCmdShort),
}

func init() {
	fcitx5Cmd.AddCommand(fcitx5AddonCmd)
}
