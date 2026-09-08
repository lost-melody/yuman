package yumeui

import (
	"fmt"
	"slices"
	"strconv"
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
	_ configui.EditorHandler = (*CommentModes)(nil)
	_ help.KeyMap            = (*CommentModesKeys)(nil)
	_ fmt.Stringer           = (*CommentModeSet)(nil)
	_ fmt.Stringer           = CommentMode(0)
)

const (
	CommentModeDiv CommentMode = 1 << iota
	CommentModeCode
	CommentModeFullDiv
	CommentModePinyin
	CommentModeMeaning
	CommentModeCharset
	CommentModeUnicode
)

type CommentMode int

type CommentModeSet struct {
	Name    string
	Modes   []CommentMode
	Enabled bool
}

type CommentModes struct {
	options map[string]any
	option  *fcitx5.ConfigOption
	value   *configui.OptionValue
	editing int64
	Table   []*CommentModeSet
	Keys    CommentModesKeys
}

type CommentModesKeys struct {
	Toggle   key.Binding
	Edit     key.Binding
	Prev     key.Binding
	Next     key.Binding
	Insert   key.Binding
	Delete   key.Binding
	SwapPrev key.Binding
	SwapNext key.Binding
}

func NewCommentModes() *CommentModes {
	return &CommentModes{
		Keys: NewCommentModesKeys(),
	}
}

func NewCommentModesKeys() CommentModesKeys {
	return CommentModesKeys{
		Toggle: key.NewBinding(
			key.WithKeys("space"),
			key.WithHelp("space", "toggle"),
		),
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

func NewCommentModeSet(line string) *CommentModeSet {
	var set CommentModeSet
	set.Unformat(line)
	return &set
}

// Predicate implements [yumeui.EditorHandler].
func (h *CommentModes) Predicate(option *fcitx5.ConfigOption) bool {
	return configui.OptionType(option.Type) == configui.OptionTypeStrings && option.Name == "CommentModes"
}

// SetOption implements [yumeui.EditorHandler].
func (h *CommentModes) SetOption(options map[string]any, option *fcitx5.ConfigOption, value *configui.OptionValue) {
	h.options = options
	h.option = option
	h.value = value
	h.Table = lo.Map(value.Strings.Value, func(line string, _ int) *CommentModeSet {
		return NewCommentModeSet(line)
	})
}

// KeyMap implements [yumeui.EditorHandler].
func (h *CommentModes) KeyMap() help.KeyMap {
	return &h.Keys
}

// Init implements [yumeui.EditorHandler].
func (h *CommentModes) Init() tea.Cmd {
	return nil
}

// Update implements [yumeui.EditorHandler].
func (h *CommentModes) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = h
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.KeyPressMsg:
		value := &h.value.Strings
		switch {
		case key.Matches(msg, h.Keys.Toggle):
			if len(value.Value) == 0 {
				break
			}
			h.Table[value.Index].Enabled = !h.Table[value.Index].Enabled
			value.Value[value.Index] = h.Table[value.Index].Format()
			value.Apply(h.options, h.option, value.Index, value.Index)
		case key.Matches(msg, h.Keys.Edit):
			if len(value.Value) == 0 {
				break
			}
			nameField := huh.NewInput().
				Title(h.option.Title).
				Prompt("Name: ").
				Value(&h.Table[value.Index].Name)
			field := huh.NewMultiSelect[CommentMode]().
				Options(h.AvailableModeOptions()...).
				Value(&h.Table[value.Index].Modes)
			modal := ui.NewHuhModal(nameField, field)
			h.editing = modal.ModalID()
			cmds.Add(ui.CmdModal(modal))
		case key.Matches(msg, h.Keys.Prev):
			value.Prev(h.options, h.option)
		case key.Matches(msg, h.Keys.Next):
			value.Next(h.options, h.option)
		case key.Matches(msg, h.Keys.Insert):
			h.Table = slices.Insert(h.Table, value.Index, NewCommentModeSet("註解|0"))
			value.Insert(h.options, h.option)
		case key.Matches(msg, h.Keys.Delete):
			if value.Index < len(h.Table) {
				h.Table = slices.Delete(h.Table, value.Index, value.Index+1)
				value.Delete(h.options, h.option)
			}
		case key.Matches(msg, h.Keys.SwapPrev):
			idx := value.Index
			h.Table[idx], h.Table[idx-1] = h.Table[idx-1], h.Table[idx]
			value.SwapPrev(h.options, h.option)
		case key.Matches(msg, h.Keys.SwapNext):
			idx := value.Index
			h.Table[idx], h.Table[idx+1] = h.Table[idx+1], h.Table[idx]
			value.SwapNext(h.options, h.option)

		default:
		}

	case ui.MsgModalDone:
		if h.editing == msg.ModalID() {
			value := &h.value.Strings
			value.Value[value.Index] = h.Table[value.Index].Format()
			value.Apply(h.options, h.option, value.Index, value.Index)
		}

	default:
		_ = msg
	}

	cmd = cmds.Batch()
	return
}

// View implements [yumeui.EditorHandler].
func (h *CommentModes) View() tea.View {
	value := &h.value.Strings
	if len(value.Value) == 0 {
		valueView := ui.StyleDimmed.Render("  Empty list")
		return tea.NewView(valueView)
	} else {
		valueView := lg.JoinVertical(
			lg.Left,
			lo.Map(h.Table, func(set *CommentModeSet, idx int) string {
				style := lg.NewStyle()
				var indicator string
				selected := idx == value.Index
				if h.Table[idx].Enabled {
					style = style.Foreground(lg.Green)
				}
				if selected {
					style = style.Foreground(lg.Magenta)
					indicator = ">"
				}
				return style.Render(lg.Sprintf(" %1s %3d %s", indicator, idx, set.String()))
			})...,
		)
		return tea.NewView(valueView)
	}
}

func (h *CommentModes) AvailableModeOptions() []huh.Option[CommentMode] {
	return huh.NewOptions(AvailableCommentModes()...)
}

// ShortHelp implements [help.KeyMap].
func (k *CommentModesKeys) ShortHelp() []key.Binding {
	return []key.Binding{k.Toggle, k.Edit, k.Prev, k.Next}
}

// FullHelp implements [help.KeyMap].
func (k *CommentModesKeys) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Toggle, k.Edit, k.Prev, k.Next},
		{k.Prev, k.Next, k.SwapPrev, k.SwapNext},
	}
}

func AvailableCommentModes() []CommentMode {
	return []CommentMode{
		CommentModeDiv,
		CommentModeCode,
		CommentModeFullDiv,
		CommentModePinyin,
		CommentModeMeaning,
		CommentModeCharset,
		CommentModeUnicode,
	}
}

// String implements [fmt.Stringer].
func (c CommentMode) String() string {
	switch c {
	case CommentModeDiv:
		return "拆分"
	case CommentModeCode:
		return "編碼"
	case CommentModeFullDiv:
		return "全息拆分"
	case CommentModePinyin:
		return "拼音"
	case CommentModeMeaning:
		return "字義"
	case CommentModeCharset:
		return "字集"
	case CommentModeUnicode:
		return "統一碼碼位"
	default:
		return "Unknown"
	}
}

// String implements [fmt.Stringer].
func (c *CommentModeSet) String() string {
	var indicator string
	if c.Enabled {
		indicator = "+"
	}
	desc := strings.Join(lo.Map(c.Modes, func(mode CommentMode, _ int) string {
		return mode.String()
	}), ", ")
	return fmt.Sprintf("[%1s] %s: %s", indicator, c.Name, desc)
}

func (c *CommentModeSet) Unformat(line string) {
	split := strings.Split(line, "|")
	if len(split) != 2 {
		return
	}
	c.Enabled = !strings.HasPrefix(line, "-")
	c.Name = split[0]
	c.Modes = nil
	if !c.Enabled {
		c.Name = c.Name[1:]
	}
	mask, _ := strconv.ParseInt(split[1], 10, 64)
	for _, mode := range AvailableCommentModes() {
		if mode&CommentMode(mask) != 0 {
			c.Modes = append(c.Modes, mode)
		}
	}
}

func (c *CommentModeSet) Format() string {
	var indicator string
	if !c.Enabled {
		indicator = "-"
	}
	var mask CommentMode
	for _, m := range c.Modes {
		mask |= m
	}
	return fmt.Sprintf("%s%s|%d", indicator, c.Name, mask)
}
