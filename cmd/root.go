// Package cmd registers subcommands and flags.
package cmd

import (
	"os"
	"runtime"

	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	Version   string = "dev"
	GitCommit string = "unknown"
	GoVersion string = runtime.Version()
)

var (
	MsgRootCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RootCmdDesc",
			Other: "The command line helper for the yume engine",
		},
	}
	MsgRootFlagVerbose = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RootFlagVerbose",
			Other: "Print verbose information",
		},
	}
	MsgRootFlagJSON = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "RootFlagJSON",
			Other: "Print messages in JSON format",
		},
	}
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "yuman",
	Short: tr.Localize(&MsgRootCmdShort),
	RunE:  runRoot,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().BoolP(flags.Verbose, "v", false, tr.Localize(&MsgRootFlagVerbose))
	rootCmd.PersistentFlags().BoolP(flags.JSON, "j", false, tr.Localize(&MsgRootFlagJSON))
	rootCmd.Flags().BoolP(flags.Version, "V", false, versionCmd.Short)
}

func runRoot(cmd *cobra.Command, args []string) error {
	flagVersion, _ := cmd.Flags().GetBool("version")
	if flagVersion {
		return versionCmd.RunE(versionCmd, nil)
	}
	return cmd.Help()
}
