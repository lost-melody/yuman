package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	restartCmd    = createShorthandFor(fcitx5RestartCmd, "", "fcitx5 restart")
	reloadCmd     = createShorthandFor(fcitx5ConfigReloadCmd, "", "fcitx5 config reload")
	installCmd    = createShorthandFor(yumeInstallCmd, "", "yume install")
	uninstallCmd  = createShorthandFor(yumeUninstallCmd, "", "yume uninstall")
	updateyumeCmd = createShorthandFor(yumeUpdateCmd, "update-yume", "yume update")
	importCmd     = createShorthandFor(yumeCustomImportCmd, "", "yume custom import")
)

var MsgCmdShorthandFor = func(command string) *i18n.LocalizeConfig {
	return &i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "CmdShorthandFor",
			Other: " (shorthand for '{{.Command}}')",
		},
		TemplateData: map[string]any{
			"Command": command,
		},
	}
}

func init() {
	rootCmd.AddCommand(restartCmd)
	rootCmd.AddCommand(reloadCmd)
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(updateyumeCmd)
	rootCmd.AddCommand(importCmd)
}

// createShorthandFor creates a copy for cmd which will later be added to rootCmd or
// other parent commands. Note that cmd should not expect a specific parent command.
func createShorthandFor(cmd *cobra.Command, use, target string) *cobra.Command {
	shorthand := *cmd
	if use != "" {
		shorthand.Use = use
	}
	if target != "" {
		shorthand.Short += tr.Localize(MsgCmdShorthandFor(target))
	}
	return &shorthand
}
