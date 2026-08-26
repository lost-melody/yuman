package cmd

var (
	restartCmd    = *fcitx5RestartCmd
	reloadCmd     = *fcitx5ConfigReloadCmd
	installCmd    = *yumeInstallCmd
	uninstallCmd  = *yumeUninstallCmd
	updateyumeCmd = *yumeUpdateCmd
	importCmd     = *yumeCustomImportCmd
)

func init() {
	restartCmd.Short += " (shorthand for fcitx5 restart)"
	reloadCmd.Short += " (shorthand for fcitx5 config reload)"
	installCmd.Short += " (shorthand for yume install)"
	uninstallCmd.Short += " (shorthand for yume uninstall)"
	updateyumeCmd.Use = "update-yume"
	updateyumeCmd.Short += " (shorthand for yume update)"
	importCmd.Short += " (shorthand for yume custom import)"

	rootCmd.AddCommand(&restartCmd)
	rootCmd.AddCommand(&reloadCmd)
	rootCmd.AddCommand(&installCmd)
	rootCmd.AddCommand(&uninstallCmd)
	rootCmd.AddCommand(&updateyumeCmd)
	rootCmd.AddCommand(&importCmd)
}
