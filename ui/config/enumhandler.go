package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
	"github.com/samber/lo"
)

var (
	_ EditorHandler = (*EnumHandler)(nil)
	_ help.KeyMap   = (*EnumHandlerKeys)(nil)
)

type EnumHandler struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *OptionValue
	Keys    EnumHandlerKeys
}

type EnumHandlerKeys struct {
	Prev key.Binding
	Next key.Binding
}

func NewEnumHandler() *EnumHandler {
	return &EnumHandler{
		Keys: NewEnumHandlerKeys(),
	}
}

func NewEnumHandlerKeys() EnumHandlerKeys {
	return EnumHandlerKeys{
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
func (h *EnumHandler) Predicate(option *fcitx5.ConfigOption) bool {
	return OptionType(option.Type) == OptionTypeEnum
}

// SetOption implements [yumeui.EditorHandler].
func (h *EnumHandler) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue) {
	h.options = options
	h.option = option
	h.value = value
}

// KeyMap implements [yumeui.EditorHandler].
func (h *EnumHandler) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *EnumHandler) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *EnumHandler) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.Enum
		switch {
		case key.Matches(msg, h.Keys.Prev):
			value.Prev(h.options, h.option)
		case key.Matches(msg, h.Keys.Next):
			value.Next(h.options, h.option)

		default:
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *EnumHandler) View() tea.View {
	value := &h.value.Enum
	valueView := lg.JoinVertical(
		lg.Left,
		lo.Map(value.Names, func(name string, idx int) string {
			style := lg.NewStyle()
			var indicator string
			selected := idx == value.Index
			if selected {
				style = style.Foreground(lg.Magenta)
				indicator = ">"
			}
			return style.Render(lg.Sprintf(" %1s %s", indicator, name))
		})...,
	)
	return tea.NewView(valueView)
}

// ShortHelp implements [help.KeyMap].
func (k *EnumHandlerKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Prev, k.Next}
}

// FullHelp implements [help.KeyMap].
func (k *EnumHandlerKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Prev, k.Next},
	}
}
