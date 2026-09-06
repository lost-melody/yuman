package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
)

var (
	_ EditorHandler = (*StringHandler)(nil)
	_ help.KeyMap   = (*StringHandlerKeys)(nil)
)

type StringHandler struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *OptionValue
	editing int64
	Keys    StringHandlerKeys
}

type StringHandlerKeys struct {
	Edit key.Binding
}

func NewStringHandler() *StringHandler {
	return &StringHandler{
		Keys: NewStringHandlerKeys(),
	}
}

func NewStringHandlerKeys() StringHandlerKeys {
	return StringHandlerKeys{
		Edit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *StringHandler) Predicate(option *fcitx5.ConfigOption) bool {
	return OptionType(option.Type) == OptionTypeString
}

// SetOption implements [yumeui.EditorHandler].
func (h *StringHandler) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue) {
	h.options = options
	h.option = option
	h.value = value
}

// KeyMap implements [yumeui.EditorHandler].
func (h *StringHandler) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *StringHandler) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *StringHandler) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.String
		switch {
		case key.Matches(msg, h.Keys.Edit):
			field := huh.NewInput().
				Title(h.option.Title).
				Value(&value.Value)
			modal := ui.NewHuhModal(field)
			h.editing = modal.ModalID()
			cmds.Add(ui.CmdModal(modal))

		default:
		}

	case ui.MsgModalDone:
		if h.editing == msg.ModalID() {
			h.value.String.Apply(h.options, h.option)
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *StringHandler) View() tea.View {
	value := &h.value.String
	valueView := lg.Sprintf("String: %s", value.Value)
	return tea.NewView(valueView)
}

// ShortHelp implements [help.KeyMap].
func (k *StringHandlerKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Edit}
}

// FullHelp implements [help.KeyMap].
func (k *StringHandlerKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Edit},
	}
}
