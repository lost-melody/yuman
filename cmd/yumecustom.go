package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeCustomCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeCustomCmdDesc",
		Other: "Manages yume custom schemas",
	},
}

// yumeCustomCmd represents the custom command
var yumeCustomCmd = &cobra.Command{
	Use:   "custom",
	Short: tr.Localize(&MsgYumeCustomCmdShort),
}

func init() {
	yumeCmd.AddCommand(yumeCustomCmd)
}
