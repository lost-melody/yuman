package yume

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lost-melody/yuman/tr"
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

// compileBinaryName is the yume-compile helper shipped next to yuman.
const compileBinaryName = "yume-compile"

var (
	MsgImportingSchema = func(name string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ImportingSchema",
				Other: "Importing schema '{{.Name}}'...",
			},
			TemplateData: map[string]any{
				"Name": name,
			},
		}
	}
	MsgImportRunning = func(command string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ImportRunning",
				Other: "Running {{.Command}}",
			},
			TemplateData: map[string]any{
				"Command": command,
			},
		}
	}
	MsgImportedSchema = func(id string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ImportedSchema",
				Other: "Imported schema '{{.ID}}'",
			},
			TemplateData: map[string]any{
				"ID": id,
			},
		}
	}
	MsgRemovedSchema = func(id string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "RemovedSchema",
				Other: "Removed schema '{{.ID}}'",
			},
			TemplateData: map[string]any{
				"ID": id,
			},
		}
	}

	MsgErrResolveCompile = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveCompile",
			Other: "locate yume-compile",
		},
	}
	MsgErrResolveCustomRoot = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrResolveCustomRoot",
			Other: "resolve custom schema root",
		},
	}
	MsgErrCompileDivision = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileDivision",
			Other: "compile division",
		},
	}
	MsgErrCompileSlotList = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileSlotList",
			Other: "list schema slots",
		},
	}
	MsgErrCompileSlotCreate = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileSlotCreate",
			Other: "create schema slot",
		},
	}
	MsgErrCompileSlotDir = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileSlotDir",
			Other: "resolve schema slot directory",
		},
	}
	MsgErrCompileCustom = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileCustom",
			Other: "compile custom schema",
		},
	}
	MsgErrCompileSlotSource = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileSlotSource",
			Other: "register schema source",
		},
	}
	MsgErrCompileSlotRemove = i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "ErrCompileSlotRemove",
			Other: "remove schema slot",
		},
	}
	MsgErrSchemaNotFound = func(ref string) *i18n.LocalizeConfig {
		return &i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{
				ID:    "ErrSchemaNotFound",
				Other: "schema '{{.Ref}}' not found",
			},
			TemplateData: map[string]any{
				"Ref": ref,
			},
		}
	}
)

// ImportYume imports a custom schema into yume through the yume-compile helper.
//
// The import requires both a table file and a division file. When divPath is
// empty the division is compiled from /dev/null into a temporary file. The
// helper is located next to the yuman executable, and schema slots live under
// $XDG_DATA_HOME/yume/data/custom (falling back to ~/.local/share/yume/data/custom).
// An existing slot with the same name is reused instead of created anew.
func ImportYume(ctx context.Context, name, tablePath, divPath string, verbose bool) (err error) {
	fmt.Println(tr.Localize(MsgImportingSchema(name)))

	compile, err := resolveCompileBinary()
	if err != nil {
		return wrapError(&MsgErrResolveCompile, err)
	}
	root, err := resolveCustomRoot()
	if err != nil {
		return wrapError(&MsgErrResolveCustomRoot, err)
	}

	divIn, divOut, cleanup, err := divisionPaths(divPath)
	if err != nil {
		return wrapError(&MsgErrCreateTempDir, err)
	}
	defer cleanup()

	if _, err = runCompile(ctx, compile, verbose, "--division", divIn, divOut); err != nil {
		return wrapError(&MsgErrCompileDivision, err)
	}

	id, found, err := findSlotID(ctx, compile, root, name, verbose)
	if err != nil {
		return wrapError(&MsgErrCompileSlotList, err)
	}
	if !found {
		var idOut string
		idOut, err = runCompile(ctx, compile, verbose, "--slot-create", root)
		if err != nil {
			return wrapError(&MsgErrCompileSlotCreate, err)
		}
		id = strings.TrimSpace(idOut)
	}

	slotOut, err := runCompile(ctx, compile, verbose, "--slot-dir", root, id)
	if err != nil {
		return wrapError(&MsgErrCompileSlotDir, err)
	}
	slot := strings.TrimSpace(slotOut)

	customOut, err := runCompile(ctx, compile, verbose, "--custom", name, tablePath, divOut, slot)
	if err != nil {
		return wrapError(&MsgErrCompileCustom, err)
	}
	// TODO: ask follow-up questions for these lines and re-run --custom with
	// the extra arguments the answers provide.
	questions := compileQuestions(customOut)
	for _, question := range questions {
		// print questions for now, before we finishing this todo.
		fmt.Println("Question:", question)
	}

	if _, err = runCompile(ctx, compile, verbose, "--slot-source", slot, name, tablePath); err != nil {
		return wrapError(&MsgErrCompileSlotSource, err)
	}

	fmt.Println(tr.Localize(MsgImportedSchema(id)))
	return nil
}

// resolveCompileBinary returns the yume-compile helper next to the running
// yuman executable.
func resolveCompileBinary() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(exe), compileBinaryName), nil
}

// resolveCustomRoot returns the directory that holds custom schema slots,
// honoring XDG_DATA_HOME and falling back to ~/.local/share.
func resolveCustomRoot() (string, error) {
	dataHome, err := userDataHome()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataHome, "yume", "data", "custom"), nil
}

// CustomSchema is one row of a yume-compile --slot-list output.
type CustomSchema struct {
	ID                string `json:"id"`
	SchemaTag         string `json:"schema_tag"`
	Name              string `json:"name"`
	MaxCodeLength     string `json:"max_code_length"`
	Terminators       string `json:"terminators"`
	AnnotationEnabled string `json:"annotation_enabled"`
	PartialAnnotation string `json:"partial_annotation"`
	CanRecompile      string `json:"can_recompile"`
}

// ListCustomSchemas lists the custom schemas under the custom root.
func ListCustomSchemas(ctx context.Context, verbose bool) ([]CustomSchema, error) {
	compile, err := resolveCompileBinary()
	if err != nil {
		return nil, wrapError(&MsgErrResolveCompile, err)
	}
	root, err := resolveCustomRoot()
	if err != nil {
		return nil, wrapError(&MsgErrResolveCustomRoot, err)
	}
	out, err := runCompile(ctx, compile, verbose, "--slot-list", root)
	if err != nil {
		return nil, wrapError(&MsgErrCompileSlotList, err)
	}
	return parseSlotList(out), nil
}

// FindCustomSchema returns the custom schema matching ref, which may be either
// a schema id or name.
func FindCustomSchema(ctx context.Context, ref string, verbose bool) (*CustomSchema, error) {
	schemas, err := ListCustomSchemas(ctx, verbose)
	if err != nil {
		return nil, err
	}
	for i := range schemas {
		if schemas[i].ID == ref || schemas[i].Name == ref {
			return &schemas[i], nil
		}
	}
	return nil, tr.LocalizeError(MsgErrSchemaNotFound(ref))
}

// RemoveCustomSchema removes the custom schema with the given id.
func RemoveCustomSchema(ctx context.Context, id string, verbose bool) error {
	compile, err := resolveCompileBinary()
	if err != nil {
		return wrapError(&MsgErrResolveCompile, err)
	}
	root, err := resolveCustomRoot()
	if err != nil {
		return wrapError(&MsgErrResolveCustomRoot, err)
	}
	if _, err = runCompile(ctx, compile, verbose, "--slot-remove", root, id); err != nil {
		return wrapError(&MsgErrCompileSlotRemove, err)
	}
	fmt.Println(tr.Localize(MsgRemovedSchema(id)))
	return nil
}

// findSlotID returns the slot id of an existing schema named name, using the
// yume-compile --slot-list output. found is false when no such schema exists.
func findSlotID(ctx context.Context, compile, root, name string, verbose bool) (id string, found bool, err error) {
	out, err := runCompile(ctx, compile, verbose, "--slot-list", root)
	if err != nil {
		return "", false, err
	}
	for _, schema := range parseSlotList(out) {
		if schema.Name == name {
			return schema.ID, true, nil
		}
	}
	return "", false, nil
}

// parseSlotList parses a --slot-list stdout. Each row is tab-separated:
// id, schema tag, name, max code length, terminators, annotation enabled,
// partial annotation, can recompile.
func parseSlotList(out string) []CustomSchema {
	schemas := []CustomSchema{}
	for line := range strings.SplitSeq(out, "\n") {
		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}
		schemas = append(schemas, CustomSchema{
			ID:                strings.TrimSpace(fields[0]),
			SchemaTag:         strings.TrimSpace(fields[1]),
			Name:              strings.TrimSpace(fields[2]),
			MaxCodeLength:     strings.TrimSpace(fields[3]),
			Terminators:       strings.TrimSpace(fields[4]),
			AnnotationEnabled: strings.TrimSpace(fields[5]),
			PartialAnnotation: strings.TrimSpace(fields[6]),
			CanRecompile:      strings.TrimSpace(fields[7]),
		})
	}
	return schemas
}

// divisionPaths maps a division file to its yume-compile input and output
// paths. An empty divPath compiles /dev/null into a temporary output file so
// the later --custom step has a real division artifact to consume. The caller
// must invoke the returned cleanup function.
func divisionPaths(divPath string) (input, output string, cleanup func(), err error) {
	if divPath != "" {
		return divPath, strings.TrimSuffix(divPath, filepath.Ext(divPath)) + ".ydiv", func() {}, nil
	}

	dir, mkErr := os.MkdirTemp("", "yume-import-*")
	if mkErr != nil {
		return "", "", nil, mkErr
	}
	return "/dev/null", filepath.Join(dir, "division.ydiv"), func() { _ = os.RemoveAll(dir) }, nil
}

// runCompile executes the yume-compile helper and captures its stdout.
func runCompile(ctx context.Context, compile string, verbose bool, args ...string) (string, error) {
	if verbose {
		cmdline := append([]string{compile}, args...)
		fmt.Println(tr.Localize(MsgImportRunning(strings.Join(shellQuoteAll(cmdline), " "))))
	}

	cmd := exec.CommandContext(ctx, compile, args...)
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	err := cmd.Run()
	return stdout.String(), err
}

// compileQuestions collects the trimmed lines from a yume-compile --custom
// stdout that start with "?". These ask for extra parameters and need a
// follow-up interaction before re-running the command.
func compileQuestions(stdout string) []string {
	var questions []string
	for line := range strings.SplitSeq(stdout, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "?") {
			questions = append(questions, line)
		}
	}
	return questions
}
