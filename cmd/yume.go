package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeCmdDesc",
		Other: "Manages yume installation, configuration and data",
	},
}

// yumeCmd represents the yume command
var yumeCmd = &cobra.Command{
	Use:   "yume",
	Short: tr.Localize(&MsgYumeCmdShort),
}

func init() {
	rootCmd.AddCommand(yumeCmd)
}
