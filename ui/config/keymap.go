package configui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
)

var (
	_ help.KeyMap = (*KeyMap)(nil)
	_ help.KeyMap = (*EditorKeyMap)(nil)
)

type KeyMap struct {
	Focus      key.Binding
	Unfocus    key.Binding
	Save       key.Binding
	Help       key.Binding
	Quit       key.Binding
	ListKeys   *list.KeyMap
	EditorKeys *EditorKeyMap
}

type EditorKeyMap struct {
	HandlerKeys help.KeyMap
}

func NewKeyMap() KeyMap {
	return KeyMap{
		Focus: key.NewBinding(
			key.WithKeys("tab", "enter"),
			key.WithHelp("tab/enter", "focus editor"),
		),
		Unfocus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus options"),
		),
		Save: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "save"),
		),
		Help: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "more"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func NewEditorKeyMap() EditorKeyMap {
	return EditorKeyMap{}
}

// ShortHelp implements [help.KeyMap].
func (k *KeyMap) ShortHelp() []key.Binding {
	bindings := []key.Binding{k.Quit, k.Focus, k.Unfocus, k.Save, k.Help}
	if k.ListKeys != nil {
		l := k.ListKeys
		bindings = append(bindings, l.CursorUp, l.CursorDown, l.PrevPage, l.NextPage)
	}
	if k.EditorKeys != nil {
		bindings = append(bindings, k.EditorKeys.ShortHelp()...)
	}
	return bindings
}

// FullHelp implements [help.KeyMap].
func (k *KeyMap) FullHelp() [][]key.Binding {
	h := k.Help
	h.SetHelp("ctrl+g", "less")
	bindings := [][]key.Binding{
		{k.Quit, k.Focus, k.Unfocus, k.Save, h},
	}
	if k.ListKeys != nil {
		l := k.ListKeys
		bindings = append(
			bindings,
			[]key.Binding{l.CursorUp, l.CursorDown, l.PrevPage, l.NextPage},
			[]key.Binding{l.GoToStart, l.GoToEnd},
		)
	}
	if k.EditorKeys != nil {
		bindings = append(bindings, k.EditorKeys.FullHelp()...)
	}
	return bindings
}

// ShortHelp implements [help.KeyMap].
func (e *EditorKeyMap) ShortHelp() []key.Binding {
	var bindings []key.Binding
	if e.HandlerKeys != nil {
		bindings = append(bindings, e.HandlerKeys.ShortHelp()...)
	}
	return bindings
}

// FullHelp implements [help.KeyMap].
func (e *EditorKeyMap) FullHelp() [][]key.Binding {
	var bindings [][]key.Binding
	if e.HandlerKeys != nil {
		bindings = append(bindings, e.HandlerKeys.FullHelp()...)
	}
	return bindings
}
