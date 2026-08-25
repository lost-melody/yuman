package cmd

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/lost-melody/yuman/flags"
	"github.com/lost-melody/yuman/tr"
	"github.com/lost-melody/yuman/yume"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
)

var (
	MsgYumeCustomRemoveCmdShort = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomRemoveCmdDesc",
			Other: "Removes a yume custom schema",
		},
	}
	MsgYumeCustomRemoveConfirmYes = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomRemoveConfirmYes",
			Other: "Remove",
		},
	}
	MsgYumeCustomRemoveConfirmNo = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "YumeCustomRemoveConfirmNo",
			Other: "Cancel",
		},
	}
	MsgYumeCustomRemoveConfirm = func(ref string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "YumeCustomRemoveConfirm",
				Other: "Remove schema '{{.Ref}}'?",
			},
			TemplateData: map[string]any{
				"Ref": ref,
			},
		}
	}
)

// yumeCustomRemoveCmd represents the remove command
var yumeCustomRemoveCmd = &cobra.Command{
	Use:               "remove <id|name>",
	Short:             tr.Localize(&MsgYumeCustomRemoveCmdShort),
	Args:              cobra.ExactArgs(1),
	RunE:              runYumeCustomRemove,
	ValidArgsFunction: completeYumeCustomList,
}

func init() {
	yumeCustomCmd.AddCommand(yumeCustomRemoveCmd)
}

func runYumeCustomRemove(cmd *cobra.Command, args []string) (err error) {
	flagVerbose, _ := cmd.Flags().GetBool(flags.Verbose)

	schema, err := yume.FindCustomSchema(cmd.Context(), args[0], flagVerbose)
	if err != nil {
		return
	}

	fmt.Printf("name: %s\nid: %s\nschema_tag: %s\nmax_code_length: %s\nterminators: %s\nannotation_enabled: %s\npartial_annotation: %s\ncan_recompile: %s\n",
		schema.Name, schema.ID, schema.SchemaTag, schema.MaxCodeLength,
		schema.Terminators, schema.AnnotationEnabled, schema.PartialAnnotation,
		schema.CanRecompile)

	confirmed := false
	err = huh.NewConfirm().
		Title(tr.Localize(MsgYumeCustomRemoveConfirm(schema.Name))).
		Affirmative(tr.Localize(&MsgYumeCustomRemoveConfirmYes)).
		Negative(tr.Localize(&MsgYumeCustomRemoveConfirmNo)).
		Value(&confirmed).
		Run()
	if err != nil {
		return
	}
	if !confirmed {
		return
	}

	return yume.RemoveCustomSchema(cmd.Context(), schema.ID, flagVerbose)
}

func completeYumeCustomList(cmd *cobra.Command, args []string, toComplete string) (completions []cobra.Completion, directive cobra.ShellCompDirective) {
	directive = cobra.ShellCompDirectiveNoFileComp
	if len(args) != 0 {
		return
	}
	schemas, err := yume.ListCustomSchemas(cmd.Context(), false)
	if err != nil {
		return
	}
	completions = lo.Map(schemas, func(schema yume.CustomSchema, _ int) string {
		return schema.ID + "\t" + schema.Name
	})
	return
}
