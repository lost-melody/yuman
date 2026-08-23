package cmd

import (
	"fmt"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeInstallCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeInstallCmdDesc",
		Other: "Download and install yume engine",
	},
}

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: tr.Localize(&MsgYumeInstallCmdShort),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("install called")
	},
}

func init() {
	yumeCmd.AddCommand(installCmd)
}
