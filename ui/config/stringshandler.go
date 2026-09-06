package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
	"github.com/samber/lo"
)

var (
	_ EditorHandler = (*StringsHandler)(nil)
	_ help.KeyMap   = (*StringsHandlerKeys)(nil)
)

type StringsHandler struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *OptionValue
	editing int64
	Keys    StringsHandlerKeys
}

type StringsHandlerKeys struct {
	Edit     key.Binding
	Prev     key.Binding
	Next     key.Binding
	Insert   key.Binding
	Delete   key.Binding
	SwapPrev key.Binding
	SwapNext key.Binding
}

func NewStringsHandler() *StringsHandler {
	return &StringsHandler{
		Keys: NewStringsHandlerKeys(),
	}
}

func NewStringsHandlerKeys() StringsHandlerKeys {
	return StringsHandlerKeys{
		Edit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
		Prev: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "previous"),
		),
		Next: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "next"),
		),
		Insert: key.NewBinding(
			key.WithKeys("alt+n"),
			key.WithHelp("alt+n", "insert"),
		),
		Delete: key.NewBinding(
			key.WithKeys("alt+d"),
			key.WithHelp("alt+d", "delete"),
		),
		SwapPrev: key.NewBinding(
			key.WithKeys("alt+up", "alt+k"),
			key.WithHelp("alt+↑/alt+k", "move up"),
		),
		SwapNext: key.NewBinding(
			key.WithKeys("alt+down", "alt+j"),
			key.WithHelp("alt+↓/alt+j", "move down"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *StringsHandler) Predicate(option *fcitx5.ConfigOption) bool {
	return OptionType(option.Type) == OptionTypeStrings
}

// SetOption implements [yumeui.EditorHandler].
func (h *StringsHandler) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue) {
	h.options = options
	h.option = option
	h.value = value
}

// KeyMap implements [yumeui.EditorHandler].
func (h *StringsHandler) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *StringsHandler) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *StringsHandler) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.Strings
		switch {
		case key.Matches(msg, h.Keys.Edit):
			if len(value.Value) == 0 {
				return
			}
			field := huh.NewInput().
				Title(h.option.Title).
				Value(&value.Value[value.Index]).
				Prompt(lg.Sprintf("(%d) ", value.Index)).
				WithHeight(8)
			modal := ui.NewHuhModal(field)
			h.editing = modal.ModalID()
			cmds.Add(ui.CmdModal(modal))
		case key.Matches(msg, h.Keys.Prev):
			value.Prev(h.options, h.option)
		case key.Matches(msg, h.Keys.Next):
			value.Next(h.options, h.option)
		case key.Matches(msg, h.Keys.Insert):
			value.Insert(h.options, h.option)
		case key.Matches(msg, h.Keys.Delete):
			value.Delete(h.options, h.option)
		case key.Matches(msg, h.Keys.SwapPrev):
			value.SwapPrev(h.options, h.option)
		case key.Matches(msg, h.Keys.SwapNext):
			value.SwapNext(h.options, h.option)

		default:
		}

	case ui.MsgModalDone:
		if h.editing == msg.ModalID() {
			idx := h.value.Strings.Index
			h.value.Strings.Apply(h.options, h.option, idx, idx)
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *StringsHandler) View() tea.View {
	value := &h.value.Strings
	if len(value.Value) == 0 {
		valueView := ui.StyleDimmed.Render("  Empty list")
		return tea.NewView(valueView)
	} else {
		valueView := lg.JoinVertical(
			lg.Left,
			lo.Map(value.Value, func(item string, idx int) string {
				style := lg.NewStyle()
				var indicator string
				selected := idx == value.Index
				if selected {
					style = style.Foreground(lg.Magenta)
					indicator = ">"
				}
				return style.Render(lg.Sprintf(" %1s %3d %s", indicator, idx, item))
			})...,
		)
		return tea.NewView(valueView)
	}
}

// ShortHelp implements [help.KeyMap].
func (k *StringsHandlerKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Edit, k.Prev, k.Next, k.Insert, k.Delete}
}

// FullHelp implements [help.KeyMap].
func (k *StringsHandlerKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Edit, k.Prev, k.Next},
		{k.Insert, k.Delete, k.SwapPrev, k.SwapNext},
	}
}
