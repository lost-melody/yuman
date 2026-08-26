package cmd

import (
	"context"
	"time"

	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5ConfigReloadCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5ConfigReloadCmdDesc",
		Other: "Reloads fcitx5 config or its addons config",
	},
}

// fcitx5ConfigReloadCmd represents the reload command
var fcitx5ConfigReloadCmd = &cobra.Command{
	Use:               "reload [flags] [addon]",
	Short:             tr.Localize(&MsgFcitx5ConfigReloadCmdShort),
	Args:              cobra.MaximumNArgs(1),
	RunE:              runFcitx5ConfigReload,
	ValidArgsFunction: completeFcitx5Addons,
}

func init() {
	fcitx5ConfigCmd.AddCommand(fcitx5ConfigReloadCmd)
}

func runFcitx5ConfigReload(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()

	if len(args) == 0 {
		err = fcitx5.Controller.ReloadConfig(ctx)
	} else {
		err = fcitx5.Controller.ReloadAddonConfig(ctx, args[0])
	}
	if err != nil {
		return
	}

	return
}
