package cmd

import (
	"github.com/lost-melody/yuman/tr"
	yumeui "github.com/lost-melody/yuman/ui/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeConfigCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeConfigCmdDesc",
		Other: "Open config TUI for yume engine",
	},
}

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: tr.Localize(&MsgYumeConfigCmdShort),
	RunE:  runYumeConfig,
}

func init() {
	yumeCmd.AddCommand(configCmd)
}

func runYumeConfig(cmd *cobra.Command, args []string) (err error) {
	err = yumeui.RunYumeConfigApp(cmd.Context())
	return
}
