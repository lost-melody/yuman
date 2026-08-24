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
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var MsgFcitx5ConfigDumpCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5ConfigDumpCmdDesc",
		Other: "Dumps config of fcitx5 or its addons",
	},
}

// fcitx5ConfigDumpCmd represents the dump command
var fcitx5ConfigDumpCmd = &cobra.Command{
	Use:               "dump",
	Short:             tr.Localize(&MsgFcitx5ConfigDumpCmdShort),
	RunE:              runFcitx5ConfigDump,
	ValidArgsFunction: completeFcitx5Addons,
}

func init() {
	fcitx5ConfigCmd.AddCommand(fcitx5ConfigDumpCmd)
}

func runFcitx5ConfigDump(cmd *cobra.Command, args []string) (err error) {
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()

	var config *fcitx5.Config
	if len(args) == 0 {
		config, err = fcitx5.Controller.GetGlobalConfig(ctx)
	} else {
		config, err = fcitx5.Controller.GetAddonConfig(ctx, args[0])
	}
	if err != nil {
		return
	}

	flagJSON, _ := cmd.Root().PersistentFlags().GetBool(flags.JSON)
	if flagJSON {
		encoder := json.NewEncoder(os.Stdout)
		err = encoder.Encode(config.Options)
	} else {
		for key, value := range config.Options {
			fmt.Printf("%s: %+v\n", key, value)
		}
	}

	return
}

func completeFcitx5Addons(cmd *cobra.Command, args []string, toComplete string) (completions []cobra.Completion, directive cobra.ShellCompDirective) {
	directive = cobra.ShellCompDirectiveNoFileComp
	if len(args) != 0 {
		return
	}
	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()
	// query addons.
	addons, err := fcitx5.Controller.GetAddons(ctx)
	if err != nil {
		return
	}
	completions = lo.FilterMap(addons, func(addon *fcitx5.AddonInfo, _ int) (string, bool) {
		return addon.UniqueName, addon.Configurable
	})
	return
}
