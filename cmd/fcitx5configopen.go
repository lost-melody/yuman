package cmd

import (
	"context"
	"time"

	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5ConfigOpenCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5ConfigOpenCmdDesc",
		Other: "Opens the config GUI of fcitx5 or its addons",
	},
}

// fcitx5ConfigOpenCmd represents the open command
var fcitx5ConfigOpenCmd = &cobra.Command{
	Use:               "open [flags] [addon]",
	Short:             tr.Localize(&MsgFcitx5ConfigOpenCmdShort),
	Args:              cobra.MaximumNArgs(1),
	RunE:              runFcitx5ConfigOpen,
	ValidArgsFunction: completeFcitx5Addons,
}

func init() {
	fcitx5ConfigCmd.AddCommand(fcitx5ConfigOpenCmd)
}

func runFcitx5ConfigOpen(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()

	if len(args) == 0 {
		err = fcitx5.Controller.Configure(ctx)
	} else {
		err = fcitx5.Controller.ConfigureAddon(ctx, args[0])
	}
	return
}
