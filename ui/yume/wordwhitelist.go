package yumeui

import (
	"strings"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	ui "github.com/lost-melody/yuman/ui"
	configui "github.com/lost-melody/yuman/ui/config"
	"github.com/samber/lo"
)

var (
	_ configui.EditorHandler = (*WordWhitelist)(nil)
	_ help.KeyMap            = (*WordWhitelistKeys)(nil)
)

type WordWhitelist struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *configui.OptionValue
	editing int64
	Table   string
	Keys    WordWhitelistKeys
}

type WordWhitelistKeys struct {
	Edit key.Binding
}

func NewWordWhitelist() *WordWhitelist {
	return &WordWhitelist{
		Table: "",
		Keys:  NewWordWhitelistKeys(),
	}
}

func NewWordWhitelistKeys() WordWhitelistKeys {
	return WordWhitelistKeys{
		Edit: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "edit"),
		),
	}
}

// Predicate implements [yumeui.EditorHandler].
func (h *WordWhitelist) Predicate(option *fcitx5.ConfigOption) bool {
	return configui.OptionType(option.Type) == configui.OptionTypeStrings && option.Name == "WordWhitelist"
}

// SetOption implements [yumeui.EditorHandler].
func (h *WordWhitelist) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *configui.OptionValue) {
	h.options = options
	h.option = option
	h.value = value
	h.Table = strings.Join(value.Strings.Value, "\n")
}

// KeyMap implements [yumeui.EditorHandler].
func (h *WordWhitelist) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *WordWhitelist) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *WordWhitelist) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, h.Keys.Edit):
			field := huh.NewText().
				Title(h.option.Title).
				Value(&h.Table)
			modal := ui.NewHuhModal(field)
			h.editing = modal.ModalID()
			cmds.Add(ui.CmdModal(modal))

		default:
		}

	case ui.MsgModalDone:
		if h.editing == msg.ModalID() {
			value := &h.value.Strings
			if h.Table == "" {
				value.Value = nil
			} else {
				value.Value = strings.Split(h.Table, "\n")
			}
			value.Index = 0
			if dict, _ := h.options[h.option.Name].(map[string]any); dict != nil {
				clear(dict)
			}
			value.Apply(h.options, h.option, 0, len(value.Value))
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *WordWhitelist) View() tea.View {
	value := &h.value.Strings
	if len(value.Value) == 0 {
		return tea.NewView(ui.StyleDimmed.Render("  Empty list"))
	}
	values := lo.Map(value.Value, func(word string, idx int) string {
		return lg.Sprintf("%3d %s", idx+1, word)
	})
	valueView := ui.StyleMagenta.Padding(0, 1).Render(lg.JoinVertical(lg.Left, values...))
	return tea.NewView(valueView)
}

// ShortHelp implements [help.KeyMap].
func (k *WordWhitelistKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Edit}
}

// FullHelp implements [help.KeyMap].
func (k *WordWhitelistKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.Edit}}
}
