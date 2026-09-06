package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
)

var (
	_ EditorHandler = (*BoolHandler)(nil)
	_ help.KeyMap   = (*BoolHandlerKeys)(nil)
)

type BoolHandler struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *OptionValue
	Keys    BoolHandlerKeys
}

type BoolHandlerKeys struct {
	Toggle key.Binding
}

func NewBoolHandler() *BoolHandler {
	return &BoolHandler{
		Keys: NewBoolHandlerKeys(),
	}
}

func NewBoolHandlerKeys() BoolHandlerKeys {
	return BoolHandlerKeys{
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *BoolHandler) Predicate(option *fcitx5.ConfigOption) bool {
	return OptionType(option.Type) == OptionTypeBool
}

// SetOption implements [yumeui.EditorHandler].
func (h *BoolHandler) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue) {
	h.options = options
	h.option = option
	h.value = value
}

// KeyMap implements [yumeui.EditorHandler].
func (h *BoolHandler) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *BoolHandler) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *BoolHandler) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.Bool
		switch {
		case key.Matches(msg, h.Keys.Toggle):
			value.Toggle(h.options, h.option)

		default:
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *BoolHandler) View() tea.View {
	value := &h.value.Bool
	var valueView string
	if value.Value {
		valueView = ui.StyleMagenta.Render(" [*] Enabled")
	} else {
		valueView = ui.StyleDimmed.Render(" [ ] Disabled")
	}
	return tea.NewView(valueView)
}

// ShortHelp implements [help.KeyMap].
func (k *BoolHandlerKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle}
}

// FullHelp implements [help.KeyMap].
func (k *BoolHandlerKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Toggle},
	}
}
