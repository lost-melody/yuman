package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/spf13/cobra"
)

var MsgYumeCustomListCmdShort = i18n.LocalizeConfig{
	DefaultMessage: &i18n.Message{
		ID:    "YumeCustomListCmdDesc",
		Other: "Lists yume custom schemas",
	},
}

// yumeCustomListCmd represents the list command
var yumeCustomListCmd = &cobra.Command{
	Use:   "list",
	Short: tr.Localize(&MsgYumeCustomListCmdShort),
	RunE:  runYumeCustomList,
}

func init() {
	yumeCustomCmd.AddCommand(yumeCustomListCmd)
}

func runYumeCustomList(cmd *cobra.Command, args []string) (err error) {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)
	schemas, err := yume.ListCustomSchemas(cmd.Context(), flagVerbose)
	if err != nil {
		return
	}

	flagJSON, _ := cmd.Root().PersistentFlags().GetBool(flags.JSON)
	if flagJSON {
		encoder := json.NewEncoder(os.Stdout)
		err = encoder.Encode(schemas)
		return
	}

	for _, schema := range schemas {
		fmt.Printf("%+v\n", schema)
	}
	return
}
