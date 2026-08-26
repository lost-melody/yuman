package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var (
	MsgYumeVersionCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeVersionCmdDesc",
			Other: "Print the version of the installed yume engine",
		},
	}
	MsgYumeNotInstalled = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeNotInstalled",
			Other: "yume is not installed",
		},
	}
)

// yumeVersionCmd represents the version command
var yumeVersionCmd = &cobra.Command{
	Use:   "version",
	Short: tr.Localize(&MsgYumeVersionCmdShort),
	Args:  cobra.NoArgs,
	RunE:  runYumeVersion,
}

func init() {
	yumeCmd.AddCommand(yumeVersionCmd)
}

func runYumeVersion(cmd *cobra.Command, args []string) (err error) {
	v, found := yume.InstalledYumeVersion()
	if !found {
		fmt.Println(tr.Localize(&MsgYumeNotInstalled))
		return
	}

	flagJSON, _ := cmd.Root().PersistentFlags().GetBool(flags.JSON)
	if flagJSON {
		var data []byte
		data, err = json.Marshal(v)
		if err != nil {
			err = fmt.Errorf("marshaling version message: %w", err)
			return
		}
		fmt.Printf("%s\n", data)
		return
	}

	fmt.Printf("%+v\n", v)
	return
}
