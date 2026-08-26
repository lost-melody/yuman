package cmd

import (
	"context"
	"time"

	"charm.land/huh/v2"
	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	MsgYumeCustomImportCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportCmdDesc",
			Other: "Imports a custom schema into yume",
		},
	}
	MsgYumeCustomImportFlagName = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFlagName",
			Other: "Schema name",
		},
	}
	MsgYumeCustomImportFlagTable = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFlagTable",
			Other: "Path to the table file (.txt)",
		},
	}
	MsgYumeCustomImportFlagDiv = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFlagDiv",
			Other: "Path to the division file (.txt, defaults to /dev/null)",
		},
	}
	MsgYumeCustomImportFormName = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFormName",
			Other: "Schema name",
		},
	}
	MsgYumeCustomImportFormTable = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFormTable",
			Other: "Table file (.txt)",
		},
	}
	MsgYumeCustomImportFormDiv = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomImportFormDiv",
			Other: "Division file (.txt, optional)",
		},
	}
)

// yumeCustomImportCmd represents the import command
var yumeCustomImportCmd = &cobra.Command{
	Use:   "import [flags] [table [division]]",
	Short: tr.Localize(&MsgYumeCustomImportCmdShort),
	Args:  cobra.MaximumNArgs(2),
	RunE:  runYumeCustomImport,
}

func init() {
	yumeCustomCmd.AddCommand(yumeCustomImportCmd)
	yumeCustomImportCmd.Flags().StringP(flags.Name, "n", "", tr.Localize(&MsgYumeCustomImportFlagName))
	yumeCustomImportCmd.Flags().StringP(flags.Table, "t", "", tr.Localize(&MsgYumeCustomImportFlagTable))
	yumeCustomImportCmd.Flags().StringP(flags.Div, "d", "", tr.Localize(&MsgYumeCustomImportFlagDiv))
}

func runYumeCustomImport(cmd *cobra.Command, args []string) (err error) {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	name, _ := cmd.Flags().GetString(flags.Name)
	tablePath, _ := cmd.Flags().GetString(flags.Table)
	divPath, _ := cmd.Flags().GetString(flags.Div)

	err = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title(tr.Localize(&MsgYumeCustomImportFormName)).Value(&name),
			huh.NewInput().Title(tr.Localize(&MsgYumeCustomImportFormTable)).Value(&tablePath),
			huh.NewInput().Title(tr.Localize(&MsgYumeCustomImportFormDiv)).Value(&divPath),
		),
	).Run()
	if err != nil {
		return
	}

	if name == "" {
		err = tr.LocalizeError(MsgErrMissingParameter(flags.Name))
		return
	}
	if tablePath == "" {
		err = tr.LocalizeError(MsgErrMissingParameter(flags.Table))
		return
	}

	if err = yume.ImportYume(cmd.Context(), name, tablePath, divPath, flagVerbose); err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), time.Second)
	defer cancel()
	err = fcitx5.Controller.ReloadAddonConfig(ctx, "yume")
	return
}
