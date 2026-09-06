package yumeui

import (
	"context"
	"slices"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
	configui "github.com/lost-melody/yuman/ui/config"
	"github.com/lost-melody/yuman/yume"
	"github.com/samber/lo"
)

var (
	_ configui.EditorHandler = (*HiddenSchemas)(nil)
	_ help.KeyMap            = (*HiddenSchemasKeys)(nil)
)

type MsgSchemasLoaded struct{ Schemas []*HiddenSchemaItem }

type HiddenSchemas struct {
	options  map[string]any
	option   *fcitx5.ConfigOption
	value    *configui.OptionValue
	Spinner  spinner.Model
	Ellipsis spinner.Model
	Keys     HiddenSchemasKeys
	Schemas  []*HiddenSchemaItem
	Index    int
}

type HiddenSchemaItem struct {
	ID     string
	Name   string
	Preset bool
	Hidden bool
}

type HiddenSchemasKeys struct {
	Toggle key.Binding
	Prev   key.Binding
	Next   key.Binding
}

func NewHiddenSchemas() *HiddenSchemas {
	return &HiddenSchemas{
		Spinner:  spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		Ellipsis: spinner.New(spinner.WithSpinner(spinner.Ellipsis)),
		Keys:     NewHiddenSchemasKeys(),
	}
}

func NewHiddenSchemasKeys() HiddenSchemasKeys {
	return HiddenSchemasKeys{
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle"),
		),
		Prev: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "previous"),
		),
		Next: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "next"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *HiddenSchemas) Predicate(option *fcitx5.ConfigOption) bool {
	return configui.OptionType(option.Type) == configui.OptionTypeStrings &&
		(option.Name == "HiddenSchemas" || option.Name == "HiddenSchemes")
}

// SetOption implements [yumeui.EditorHandler].
func (h *HiddenSchemas) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *configui.OptionValue) {
	h.options = options
	h.option = option
	h.value = value
	h.Schemas = nil
	h.Index = 0
}

// KeyMap implements [yumeui.EditorHandler].
func (h *HiddenSchemas) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *HiddenSchemas) Init() tea.Cmd {
	var cmds ui.Cmds
	cmds.Add(h.Spinner.Tick)
	cmds.Add(h.Ellipsis.Tick)
	cmds.Add(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		schemas, err := yume.ListCustomSchemas(ctx, false)
		if err != nil {
			field := huh.NewNote().
				Title("Error").
				Description(err.Error())
			modal := ui.NewHuhModal(field)
			return ui.CmdModal(modal)
		}

		customSchemas := lo.Map(schemas, func(schema yume.CustomSchema, _ int) *HiddenSchemaItem {
			return &HiddenSchemaItem{
				ID:     schema.ID,
				Name:   schema.Name,
				Preset: false,
				Hidden: false,
			}
		})

		return MsgSchemasLoaded{
			Schemas: append(PresetSchemas(), customSchemas...),
		}
	})
	return cmds.Batch()
}

// Update implements [yumeui.EditorHandler].
func (h *HiddenSchemas) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, h.Keys.Toggle):
			if h.Index < len(h.Schemas) {
				h.Schemas[h.Index].Hidden = !h.Schemas[h.Index].Hidden
			}
			h.ApplyStates()
		case key.Matches(msg, h.Keys.Prev):
			if h.Index > 0 {
				h.Index--
			}
		case key.Matches(msg, h.Keys.Next):
			if h.Index < len(h.Schemas)-1 {
				h.Index++
			}
		default:
		}

	case MsgSchemasLoaded:
		h.Schemas = msg.Schemas
		h.LoadStates()

	default:
		if h.Schemas == nil {
			var c tea.Cmd
			h.Spinner, c = h.Spinner.Update(msg)
			cmds.Add(c)
			h.Ellipsis, c = h.Ellipsis.Update(msg)
			cmds.Add(c)
		}
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *HiddenSchemas) View() tea.View {
	if h.Schemas == nil {
		view := lg.Sprintf(" %s Loading%s", h.Spinner.View(), h.Ellipsis.View())
		return tea.NewView(ui.StyleDimmed.Render(view))
	}
	values := lo.Map(h.Schemas, func(schema *HiddenSchemaItem, idx int) string {
		style := lg.NewStyle().Faint(schema.Hidden)
		var indicator, hidden, suffix string
		if idx == h.Index {
			indicator = ">"
			style = style.Foreground(lg.Magenta)
		}
		if schema.Hidden {
			hidden = "*"
		}
		if schema.Preset {
			suffix = "@" + schema.ID
		} else {
			suffix = "custom." + schema.ID
		}
		return style.Render(lg.Sprintf(" %1s [%1s] %s %s", indicator, hidden, schema.Name, suffix))
	})
	valueView := lg.NewStyle().Padding(0, 1).Render(lg.JoinVertical(lg.Left, values...))
	return tea.NewView(valueView)
}

func (h *HiddenSchemas) LoadStates() {
	value := h.value.Strings
	for _, schema := range h.Schemas {
		schema.Hidden = slices.Index(value.Value, schema.FullID()) >= 0
	}
}

func (h *HiddenSchemas) ApplyStates() {
	value := &h.value.Strings
	value.Value = lo.FilterMap(h.Schemas, func(schema *HiddenSchemaItem, _ int) (string, bool) {
		return schema.FullID(), schema.Hidden
	})
	dict, _ := h.options[h.option.Name].(map[string]any)
	if dict == nil {
		dict = map[string]any{}
		h.options[h.option.Name] = dict
	}
	clear(dict)
	value.Apply(h.options, h.option, 0, len(value.Value))
}

// ShortHelp implements [help.KeyMap].
func (k *HiddenSchemasKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Prev, k.Next}
}

// FullHelp implements [help.KeyMap].
func (k *HiddenSchemasKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Toggle, k.Prev, k.Next}}
}

func (schema *HiddenSchemaItem) FullID() string {
	id := schema.ID
	if !schema.Preset {
		id = "custom." + id
	}
	return id
}

func PresetSchemas() []*HiddenSchemaItem {
	return []*HiddenSchemaItem{
		{
			ID:     "lingming",
			Name:   "靈明",
			Preset: true,
			Hidden: false,
		},
		{
			ID:     "qingyun",
			Name:   "卿雲",
			Preset: true,
			Hidden: false,
		},
		{
			ID:     "xingchen",
			Name:   "星陳",
			Preset: true,
			Hidden: false,
		},
		{
			ID:     "riyue",
			Name:   "日月",
			Preset: true,
			Hidden: false,
		},
	}
}
