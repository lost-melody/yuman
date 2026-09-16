package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeteCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeteCmdDesc",
		Other: "Manages the yumete installation",
	},
}

// yumeteCmd represents the yumete command
var yumeteCmd = &cobra.Command{
	Use:   "yumete",
	Short: tr.Localize(&MsgYumeteCmdShort),
}

func init() {
	rootCmd.AddCommand(yumeteCmd)
}
