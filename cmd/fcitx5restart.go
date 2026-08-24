package cmd

import (
	"context"
	"time"

	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	MsgFcitx5RestartCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "Fcitx5RestartCmdDesc",
			Other: "Restart fcitx5 service",
		},
	}
	MsgErrFcitx5CannotRestart = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrFcitx5CannotRestart",
			Other: "Cannot restart fcitx5 service",
		},
	}
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: tr.Localize(&MsgFcitx5RestartCmdShort),
	RunE:  runFcitx5Restart,
}

func init() {
	fcitx5Cmd.AddCommand(restartCmd)
}

func runFcitx5Restart(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()

	canRestart, err := fcitx5.Controller.CanRestart(ctx)
	if err != nil {
		return
	}
	if !canRestart {
		err = tr.LocalizeError(&MsgErrFcitx5CannotRestart)
		return
	}

	err = fcitx5.Controller.Restart(ctx)
	return
}
