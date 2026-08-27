package cmd

import (
	"fmt"

	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	MsgYumeUpdateCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUpdateCmdDesc",
			Other: "Update yume engine to the latest release",
		},
	}
	MsgYumeUpdateFlagCheck = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUpdateFlagCheck",
			Other: "Only check for updates without installing",
		},
	}
	MsgYumeUpdateChecking = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeUpdateChecking",
			Other: "Checking for the latest yume release...",
		},
	}
	MsgYumeUpdateAvailable = func(newVersion, newTimestamp, newSize, currentVersion, currentTimestamp string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeUpdateAvailable",
				Other: "Update available: {{.New}}-{{.NewTimestamp}} (size: {{.NewSize}}) (installed {{.Current}}-{{.CurrentTimestamp}})",
			},
			TemplateData: map[string]any{
				"New":              newVersion,
				"NewTimestamp":     newTimestamp,
				"NewSize":          newSize,
				"Current":          currentVersion,
				"CurrentTimestamp": currentTimestamp,
			},
		}
	}
	MsgYumeUpdateUpToDate = func(version, timestamp string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeUpdateUpToDate",
				Other: "yume is already up to date ({{.Version}}-{{.Timestamp}})",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Timestamp": timestamp,
			},
		}
	}
	MsgYumeUpdateLatest = func(version, timestamp string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeUpdateLatest",
				Other: "Latest yume version: {{.Version}}-{{.Timestamp}}",
			},
			TemplateData: map[string]any{
				"Version":   version,
				"Timestamp": timestamp,
			},
		}
	}
)

// yumeUpdateCmd represents the update command
var yumeUpdateCmd = &cobra.Command{
	Use:   "update",
	Short: tr.Localize(&MsgYumeUpdateCmdShort),
	Args:  cobra.NoArgs,
	RunE:  runYumeUpdate,
}

func init() {
	yumeCmd.AddCommand(yumeUpdateCmd)
	yumeUpdateCmd.Flags().BoolP(flags.Check, "c", false, tr.Localize(&MsgYumeUpdateFlagCheck))
}

func runYumeUpdate(cmd *cobra.Command, args []string) (err error) {
	flagCheck, _ := cmd.Flags().GetBool(flags.Check)
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)

	fmt.Println(tr.Localize(&MsgYumeUpdateChecking))

	release, err := yume.LatestYumeRelease(cmd.Context())
	if err != nil {
		return
	}

	installed, installedFound := yume.InstalledYumeVersion()

	if flagCheck {
		return reportUpdateCheck(installed, installedFound, release)
	}

	if installedFound {
		if yume.CompareReleases(installed, release) >= 0 {
			fmt.Println(tr.Localize(MsgYumeUpdateUpToDate(release.Version, release.Timestamp)))
			return nil
		}
		fmt.Println(tr.Localize(MsgYumeUpdateAvailable(
			release.Version, release.Timestamp, yume.FormatSize(release.FileSize),
			installed.Version, installed.Build,
		)))
	}

	userDirs, err := resolveInstallTarget(cmd)
	if err != nil {
		return
	}
	return yume.InstallYume(cmd.Context(), release.AssetURL, userDirs, flagVerbose)
}

// reportUpdateCheck reports whether an update is available without installing.
func reportUpdateCheck(installed yume.YumeVersion, installedFound bool, release yume.YumeRelease) error {
	if !installedFound {
		fmt.Println(tr.Localize(&MsgYumeNotInstalled))
		fmt.Println(tr.Localize(MsgYumeUpdateLatest(release.Version, release.Timestamp)))
		return nil
	}
	if yume.CompareReleases(installed, release) < 0 {
		fmt.Println(tr.Localize(MsgYumeUpdateAvailable(
			release.Version, release.Timestamp, yume.FormatSize(release.FileSize),
			installed.Version, installed.Build,
		)))
	} else {
		fmt.Println(tr.Localize(MsgYumeUpdateUpToDate(release.Version, release.Timestamp)))
	}
	return nil
}
