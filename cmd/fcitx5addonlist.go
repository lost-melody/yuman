package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgFcitx5AddonListCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5AddonListCmdDesc",
		Other: "Lists fcitx5 addons",
	},
}

// fcitx5AddonListCmd represents the list command
var fcitx5AddonListCmd = &cobra.Command{
	Use:   "list",
	Short: tr.Localize(&MsgFcitx5AddonListCmdShort),
	Args:  cobra.NoArgs,
	RunE:  runFcitx5AddonList,
}

func init() {
	fcitx5AddonCmd.AddCommand(fcitx5AddonListCmd)
}

func runFcitx5AddonList(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()

	addons, err := fcitx5.Controller.GetAddons(ctx)
	if err != nil {
		return
	}

	flagJSON, _ := cmd.Root().PersistentFlags().GetBool(flags.JSON)
	if flagJSON {
		encoder := json.NewEncoder(os.Stdout)
		err = encoder.Encode(addons)
	} else {
		for _, addon := range addons {
			fmt.Printf("%v\n", *addon)
		}
	}

	return
}
