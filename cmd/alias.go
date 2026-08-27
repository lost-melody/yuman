package cmd

import (
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
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

func registerShorthands() {
	createShorthandFor(rootCmd, fcitx5RestartCmd, "", "fcitx5 restart")
	createShorthandFor(rootCmd, fcitx5ConfigReloadCmd, "", "fcitx5 config reload")
	createShorthandFor(rootCmd, yumeInstallCmd, "", "yume install")
	createShorthandFor(rootCmd, yumeUninstallCmd, "", "yume uninstall")
	createShorthandFor(rootCmd, yumeUpdateCmd, "update-yume", "yume update")
	createShorthandFor(rootCmd, yumeCustomImportCmd, "", "yume custom import")
}

// createShorthandFor creates a copy for cmd and add it to parent command.
// Note that cmd should not expect a specific parent command.
func createShorthandFor(parent, cmd *cobra.Command, use, target string) *cobra.Command {
	shorthand := *cmd
	if use != "" {
		shorthand.Use = use
	}
	if target != "" {
		shorthand.Short += tr.Localize(MsgCmdShorthandFor(target))
	}
	parent.AddCommand(&shorthand)
	return &shorthand
}
