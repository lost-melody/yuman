package cmd

import (
	"charm.land/huh/v2"
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	MsgYumeInstallCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallCmdDesc",
			Other: "Install yume engine",
		},
	}
	MsgYumeInstallFlagUser = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallFlagUser",
			Other: "Install yume into user directories",
		},
	}
	MsgYumeInstallFlagSystem = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallFlagSystem",
			Other: "Install yume into system directories (root)",
		},
	}
	MsgYumeInstallFlagPackage = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallFlagPackage",
			Other: "Package path of yume release",
		},
	}
	MsgYumeInstallUserQuestion = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallUserQuestion",
			Other: "Where should yume be installed?",
		},
	}
	MsgYumeInstallUserQuestionYes = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallUserQuestionYes",
			Other: "User directories",
		},
	}
	MsgYumeInstallUserQuestionNo = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallUserQuestionNo",
			Other: "System directories (need root)",
		},
	}
	MsgYumeInstallSelectPackage = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeInstallSelectPackage",
			Other: "Select a yume package: (a directory or a '.tar.gz' file)",
		},
	}

	MsgErrFlagsConflict = func(flagA, flagB string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrFlagsConflict",
				Other: "flag '{{.FlagA}}' conflicts with flag '{{.FlagB}}'",
			},
			TemplateData: map[string]any{
				"FlagA": flagA,
				"FlagB": flagB,
			},
		}
	}
	MsgErrMissingParameter = func(parameter string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrMissingParameter",
				Other: "missing parameter '{{.Parameter}}'",
			},
			TemplateData: map[string]any{
				"Parameter": parameter,
			},
		}
	}
)

// yumeInstallCmd represents the install command
var yumeInstallCmd = &cobra.Command{
	Use:   "install",
	Short: tr.Localize(&MsgYumeInstallCmdShort),
	RunE:  runYumeInstall,
}

func init() {
	yumeCmd.AddCommand(yumeInstallCmd)
	yumeInstallCmd.Flags().BoolP(flags.User, "u", false, tr.Localize(&MsgYumeInstallFlagUser))
	yumeInstallCmd.Flags().BoolP(flags.System, "s", false, tr.Localize(&MsgYumeInstallFlagSystem))
	yumeInstallCmd.Flags().StringP(flags.Package, "p", "", tr.Localize(&MsgYumeInstallFlagSystem))
}

func runYumeInstall(cmd *cobra.Command, args []string) (err error) {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	flagUser, _ := cmd.Flags().GetBool(flags.User)
	flagSystem, _ := cmd.Flags().GetBool(flags.System)
	if flagUser && flagSystem {
		err = tr.LocalizeError(MsgErrFlagsConflict(flags.User, flags.System))
		return
	}
	// install into user directories or not.
	userDirs := flagUser
	if !flagUser && !flagSystem {
		userDirs = true
		err = huh.NewConfirm().
			Title(tr.Localize(&MsgYumeInstallUserQuestion)).
			Affirmative(tr.Localize(&MsgYumeInstallUserQuestionYes)).
			Negative(tr.Localize(&MsgYumeInstallUserQuestionNo)).
			Value(&userDirs).
			Run()
		if err != nil {
			return
		}
	}
	// yume package path.
	pkgPath, _ := cmd.Flags().GetString(flags.Package)
	if pkgPath == "" {
		err = huh.NewFilePicker().
			Picking(true).
			Title(tr.Localize(&MsgYumeInstallSelectPackage)).
			CurrentDirectory("..").
			FileAllowed(true).
			DirAllowed(true).
			Value(&pkgPath).
			WithHeight(8).
			Run()
		if err != nil {
			return
		}
		if pkgPath == "" {
			err = tr.LocalizeError(MsgErrMissingParameter(flags.Package))
			return
		}
	}
	err = yume.InstallYume(cmd.Context(), pkgPath, userDirs, flagVerbose)
	return
}
