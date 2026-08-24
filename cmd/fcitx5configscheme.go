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

var MsgFcitx5ConfigSchemeCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "Fcitx5ConfigSchemeCmdDesc",
		Other: "Prints the config scheme of fcitx5 or its addons",
	},
}

// fcitx5ConfigSchemeCmd represents the scheme command
var fcitx5ConfigSchemeCmd = &cobra.Command{
	Use:               "scheme",
	Short:             tr.Localize(&MsgFcitx5ConfigSchemeCmdShort),
	RunE:              runFcitx5ConfigScheme,
	ValidArgsFunction: completeFcitx5Addons,
}

func init() {
	fcitx5ConfigCmd.AddCommand(fcitx5ConfigSchemeCmd)
}

func runFcitx5ConfigScheme(cmd *cobra.Command, args []string) (err error) {
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
		err = encoder.Encode(config.Schemes)
	} else {
		for _, scheme := range config.Schemes {
			fmt.Println("Scheme:", scheme.Name)
			for _, option := range scheme.Options {
				fmt.Printf("    %+v\n", *option)
			}
		}
	}

	return
}
