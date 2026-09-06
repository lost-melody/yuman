package configui

import (
	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/ui"
)

var _ ui.Model = (*EditorModel)(nil)

type EditorModel struct {
	option   *fcitx5.ConfigOption
	options  map[string]any
	width    int
	Help     help.Model
	KeyMap   EditorKeyMap
	Value    OptionValue
	Handlers []EditorHandler
}

type EditorHandler interface {
	Predicate(option *fcitx5.ConfigOption) bool
	SetOption(options map[string]any, option *fcitx5.ConfigOption, value *OptionValue)
	KeyMap() help.KeyMap
	tea.Model
}

func NewEditorModel(options map[string]any) *EditorModel {
	return &EditorModel{
		option:  &fcitx5.ConfigOption{},
		options: options,
		width:   0,
		Help:    help.Model{},
		KeyMap:  NewEditorKeyMap(),
		Value:   OptionValue{},
	}
}

// Init implements [ui.Model].
func (editor *EditorModel) Init() tea.Cmd {
	return nil
}

// Update implements [ui.Model].
func (editor *EditorModel) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = editor
	var cmds ui.Cmds
	switch msg := message.(type) {
	default:
		c, handled := editor.HandlersUpdate(msg)
		if handled {
			cmds.Add(c)
			break
		}
	}

	cmd = cmds.Batch()
	return
}

// View implements [ui.Model].
func (editor *EditorModel) View() tea.View {
	if editor.option == nil || editor.options == nil {
		return tea.NewView("ERROR")
	}
	option := editor.option

	valueView := "Unknown"
	for _, handler := range editor.Handlers {
		if handler == nil || !handler.Predicate(option) {
			continue
		}
		valueView = handler.View().Content
		break
	}

	view := lg.JoinVertical(
		lg.Left,
		"",
		ui.StyleBold.Foreground(lg.BrightBlue).Render(option.Title),
		ui.StyleDimmed.Render(lg.Sprintf("  %s: %s", option.Name, option.Type)),
		"",
		valueView,
	)

	return tea.NewView(view)
}

// HasError implements [ui.Model].
func (editor *EditorModel) HasError() error {
	return nil
}

func (editor *EditorModel) WithHandlers(handlers ...EditorHandler) *EditorModel {
	editor.Handlers = handlers
	return editor
}

func (editor *EditorModel) SetSelectedItem(item *ListItem) tea.Cmd {
	options := editor.options
	option := item.Option
	value := options[option.Name]

	editor.option = option
	editor.Value.Set(option, value)

	var cmds ui.Cmds
	editor.KeyMap.HandlerKeys = nil
	for _, handler := range editor.Handlers {
		if handler == nil || !handler.Predicate(option) {
			continue
		}

		editor.KeyMap.HandlerKeys = handler.KeyMap()
		handler.SetOption(options, option, &editor.Value)
		cmds.Add(handler.Init())

		break
	}

	return cmds.Batch()
}

func (editor *EditorModel) HandlersUpdate(msg tea.Msg) (cmd tea.Cmd, handled bool) {
	for i, handler := range editor.Handlers {
		if handler == nil || !handler.Predicate(editor.option) {
			continue
		}
		handled = true
		m, c := handler.Update(msg)
		editor.Handlers[i], _ = m.(EditorHandler)
		cmd = c
		break
	}
	return
}

func (editor *EditorModel) Width() int {
	return editor.width
}

func (editor *EditorModel) SetWidth(width int) {
	editor.width = width
}
