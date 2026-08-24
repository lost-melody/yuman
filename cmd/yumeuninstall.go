package cmd

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

// Choices offered by the interactive uninstall prompt.
const (
	uninstallChoiceUser   = "user"
	uninstallChoiceSystem = "system"
)

var (
	MsgYumeUninstallCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallCmdDesc",
			Other: "Uninstall yume engine",
		},
	}
	MsgYumeUninstallFlagUser = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallFlagUser",
			Other: "Uninstall yume from user directories",
		},
	}
	MsgYumeUninstallFlagSystem = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallFlagSystem",
			Other: "Uninstall yume from system directories (root)",
		},
	}
	MsgYumeUninstallFlagPurge = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallFlagPurge",
			Other: "Also remove yume user data (~/.local/share/yume)",
		},
	}
	MsgYumeUninstallQuestion = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallQuestion",
			Other: "Which yume installations should be uninstalled?",
		},
	}
	MsgYumeUninstallQuestionUser = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallQuestionUser",
			Other: "User directories",
		},
	}
	MsgYumeUninstallQuestionSystem = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallQuestionSystem",
			Other: "System directories (need root)",
		},
	}
	MsgYumeUninstallConfirmYes = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallConfirmYes",
			Other: "Uninstall",
		},
	}
	MsgYumeUninstallConfirmNo = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallConfirmNo",
			Other: "Cancel",
		},
	}
	MsgYumeUninstallPurgeQuestion = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallPurgeQuestion",
			Other: "Also remove yume user data (~/.local/share/yume)?",
		},
	}
	MsgYumeUninstallPurgeYes = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallPurgeYes",
			Other: "Remove",
		},
	}
	MsgYumeUninstallPurgeNo = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUninstallPurgeNo",
			Other: "Keep",
		},
	}

	MsgYumeUninstallConfirm = func(targets string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeUninstallConfirm",
				Other: "Uninstall yume from {{.Targets}}?",
			},
			TemplateData: map[string]any{
				"Targets": targets,
			},
		}
	}
	MsgYumeUninstallDataKept = func(path string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeUninstallDataKept",
				Other: "Keeping yume user data at '{{.Path}}' (pass --purge to remove it)",
			},
			TemplateData: map[string]any{
				"Path": path,
			},
		}
	}
	MsgErrNothingSelected = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrNothingSelected",
			Other: "nothing selected",
		},
	}
)

// yumeUninstallCmd represents the uninstall command
var yumeUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: tr.Localize(&MsgYumeUninstallCmdShort),
	RunE:  runYumeUninstall,
}

func init() {
	yumeCmd.AddCommand(yumeUninstallCmd)
	yumeUninstallCmd.Flags().BoolP(flags.User, "u", false, tr.Localize(&MsgYumeUninstallFlagUser))
	yumeUninstallCmd.Flags().BoolP(flags.System, "s", false, tr.Localize(&MsgYumeUninstallFlagSystem))
	yumeUninstallCmd.Flags().Bool(flags.Purge, false, tr.Localize(&MsgYumeUninstallFlagPurge))
}

func runYumeUninstall(cmd *cobra.Command, args []string) (err error) {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	flagUser, _ := cmd.Flags().GetBool(flags.User)
	flagSystem, _ := cmd.Flags().GetBool(flags.System)
	flagPurge, _ := cmd.Flags().GetBool(flags.Purge)
	if !flagUser && !flagSystem {
		// Uninstall targets: user directories, system directories, or both.
		var choices []string
		err = huh.NewMultiSelect[string]().
			Title(tr.Localize(&MsgYumeUninstallQuestion)).
			Options(
				huh.NewOption(tr.Localize(&MsgYumeUninstallQuestionUser), uninstallChoiceUser),
				huh.NewOption(tr.Localize(&MsgYumeUninstallQuestionSystem), uninstallChoiceSystem),
			).
			Value(&choices).
			Height(3).
			Run()
		if err != nil {
			return
		}
		for _, choice := range choices {
			switch choice {
			case uninstallChoiceUser:
				flagUser = true
			case uninstallChoiceSystem:
				flagSystem = true
			}
		}
		if !flagUser && !flagSystem {
			err = tr.LocalizeError(&MsgErrNothingSelected)
			return
		}
	}
	// Confirm the uninstallation so an accidental key press cannot delete files.
	var targets []string
	if flagUser {
		targets = append(targets, tr.Localize(&MsgYumeUninstallQuestionUser))
	}
	if flagSystem {
		targets = append(targets, tr.Localize(&MsgYumeUninstallQuestionSystem))
	}
	confirmed := false
	err = huh.NewConfirm().
		Title(tr.Localize(MsgYumeUninstallConfirm(strings.Join(targets, ", ")))).
		Affirmative(tr.Localize(&MsgYumeUninstallConfirmYes)).
		Negative(tr.Localize(&MsgYumeUninstallConfirmNo)).
		Value(&confirmed).
		Run()
	if err != nil {
		return
	}
	if !confirmed {
		return
	}
	// The yume data directory holds user data, so removing it needs either
	// the --purge flag or an explicit confirmation. The cursor starts on
	// "Keep" because huh focuses the button that matches the initial value.
	purge := flagPurge
	if !purge {
		if dataDir, ok := yume.ExistingYumeDataDir(); ok {
			removeData := false
			err = huh.NewConfirm().
				Title(tr.Localize(&MsgYumeUninstallPurgeQuestion)).
				Affirmative(tr.Localize(&MsgYumeUninstallPurgeYes)).
				Negative(tr.Localize(&MsgYumeUninstallPurgeNo)).
				Value(&removeData).
				Run()
			if err != nil {
				return
			}
			if removeData {
				purge = true
			} else {
				fmt.Printf("%s\n", tr.Localize(MsgYumeUninstallDataKept(dataDir)))
			}
		}
	}

	// Uninstall the user installation first, then the system installation.
	if flagUser {
		if err = yume.UninstallYume(cmd.Context(), true, purge, flagVerbose); err != nil {
			return
		}
	}
	if flagSystem {
		err = yume.UninstallYume(cmd.Context(), false, purge, flagVerbose)
	}
	return
}
