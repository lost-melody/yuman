package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgVersionCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "VersionCmdDesc",
		Other: "Print the version number of yuman",
	},
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: tr.Localize(&MsgVersionCmdShort),
	RunE:  runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func runVersion(cmd *cobra.Command, args []string) (err error) {
	flagJSON, _ := cmd.Root().PersistentFlags().GetBool(flags.JSON)
	if !flagJSON {
		fmt.Printf("%s (%s) %s\n", Version, GitCommit, GoVersion)
		return
	}

	type VersionMsg struct {
		Version   string `json:"version"`
		GitCommit string `json:"commit"`
		GoVersion string `json:"go_version"`
	}
	data, err := json.Marshal(&VersionMsg{
		Version:   Version,
		GitCommit: GitCommit,
		GoVersion: GoVersion,
	})
	if err != nil {
		err = fmt.Errorf("marshaling version message: %w", err)
		return
	}
	fmt.Printf("%s\n", data)

	return
}
