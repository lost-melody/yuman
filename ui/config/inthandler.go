package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
)

var (
	_ EditorHandler = (*IntHandler)(nil)
	_ help.KeyMap   = (*IntHandlerKeys)(nil)
)

type IntHandler struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *OptionValue
	Keys    IntHandlerKeys
}

type IntHandlerKeys struct {
	Inc key.Binding
	Dec key.Binding
	Max key.Binding
	Min key.Binding
}

func NewIntHandler() *IntHandler {
	return &IntHandler{
		Keys: NewIntHandlerKeys(),
	}
}

func NewIntHandlerKeys() IntHandlerKeys {
	return IntHandlerKeys{
		Inc: key.NewBinding(
			key.WithKeys("up", "k", "ctrl+a"),
			key.WithHelp("↑/k/ctrl+a", "increase"),
		),
		Dec: key.NewBinding(
			key.WithKeys("down", "j", "ctrl+x"),
			key.WithHelp("↓/j/ctrl+x", "decrease"),
		),
		Min: key.NewBinding(
			key.WithKeys("J"),
			key.WithHelp("J", "minimum"),
		),
		Max: key.NewBinding(
			key.WithKeys("K"),
			key.WithHelp("K", "maximum"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *IntHandler) Predicate(option *fcitx5.ConfigOption) bool {
	return OptionType(option.Type) == OptionTypeInt
}

// SetOption implements [yumeui.EditorHandler].
func (h *IntHandler) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue) {
	h.options = options
	h.option = option
	h.value = value
}

// KeyMap implements [yumeui.EditorHandler].
func (h *IntHandler) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *IntHandler) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *IntHandler) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.Int
		switch {
		case key.Matches(msg, h.Keys.Inc):
			value.Inc(h.options, h.option)
		case key.Matches(msg, h.Keys.Dec):
			value.Dec(h.options, h.option)
		case key.Matches(msg, h.Keys.Min):
			value.SetMin(h.options, h.option)
		case key.Matches(msg, h.Keys.Max):
			value.SetMax(h.options, h.option)

		default:
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *IntHandler) View() tea.View {
	value := &h.value.Int
	valueView := lg.Sprintf("Current: %d", value.Value)
	if value.Min != nil && value.Max != nil {
		valueView += ui.StyleDimmed.Render(lg.Sprintf("\n  Range: [%d, %d]", *value.Min, *value.Max))
	} else if value.Min != nil {
		valueView += ui.StyleDimmed.Render(lg.Sprintf("\n  Minimum: %d", *value.Min))
	} else if value.Max != nil {
		valueView += ui.StyleDimmed.Render(lg.Sprintf("\n  Maximum: %d", *value.Max))
	}
	return tea.NewView(valueView)
}

// ShortHelp implements [help.KeyMap].
func (k *IntHandlerKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Inc, k.Dec, k.Max, k.Min}
}

// FullHelp implements [help.KeyMap].
func (k *IntHandlerKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Inc, k.Dec, k.Max, k.Min},
	}
}
