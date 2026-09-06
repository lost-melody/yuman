package configui

import (
	"context"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/ui"
)

var _ ui.Model = (*App)(nil)

type App struct {
	ctx           context.Context
	err           error
	size          tea.WindowSizeMsg
	lastKey       string
	lastIndex     int
	editorFocused bool
	quitting      bool
	Modal         tea.Model
	List          list.Model
	Editor        *EditorModel
	Help          help.Model
	KeyMap        KeyMap
}

func NewApp(
	ctx context.Context,
	optionList list.Model, editor *EditorModel,
) App {
	app := App{
		ctx:       ctx,
		lastIndex: -1,
		List:      optionList,
		Editor:    editor,
		Help:      help.New(),
		KeyMap:    NewKeyMap(),
	}
	app.UpdateFocus()
	return app
}

// Init implements [tea.Model].
func (app *App) Init() tea.Cmd {
	var cmds ui.Cmds
	cmds.Add(app.Editor.Init())
	return cmds.Batch()
}

// Update implements [tea.Model].
func (app *App) Update(message tea.Msg) (model tea.Model, cmd tea.Cmd) {
	model = app
	var cmds ui.Cmds

	switch msg := message.(type) {
	case tea.WindowSizeMsg:
		app.size = msg
		app.List.SetSize(app.listViewWidth()-2, app.viewHeight()-1)
		app.Editor.SetWidth(app.editorViewWidth() - 2)

	case MsgError:
		if msg.Err != nil {
			app.err = msg.Err
			cmds.Add(tea.Quit)
		}

	case ui.MsgModal:
		app.Modal = msg.Modal
		cmds.Add(msg.Modal.Init())

	case tea.KeyPressMsg:
		app.lastKey = msg.Keystroke()
		if app.Modal != nil {
			var c tea.Cmd
			app.Modal, c = app.Modal.Update(msg)
			cmds.Add(c)
			break
		}
		switch {
		case key.Matches(msg, app.KeyMap.Focus):
			app.editorFocused = true
			app.UpdateFocus()
		case key.Matches(msg, app.KeyMap.Unfocus):
			app.editorFocused = false
			app.UpdateFocus()
		case key.Matches(msg, app.KeyMap.Save):
			cmds.Add(app.SaveOptions())
		case key.Matches(msg, app.KeyMap.Help):
			app.Help.ShowAll = !app.Help.ShowAll
		case key.Matches(msg, app.KeyMap.Quit):
			app.quitting = true
			cmds.Add(tea.Quit)
		default:
			if app.editorFocused {
				editor, c := app.Editor.Update(message)
				app.Editor = editor.(*EditorModel)
				cmds.Add(c)
			} else {
				var c tea.Cmd
				app.List, c = app.List.Update(message)
				cmds.Add(c)
			}
		}

	default:
		var c tea.Cmd
		if app.Modal != nil {
			app.Modal, c = app.Modal.Update(msg)
			cmds.Add(c)
			break
		}

		app.List, c = app.List.Update(message)
		cmds.Add(c)

		editor, c := app.Editor.Update(message)
		app.Editor = editor.(*EditorModel)
		cmds.Add(c)
	}

	if app.lastIndex != app.List.GlobalIndex() {
		app.lastIndex = app.List.GlobalIndex()
		selectedItem := app.List.SelectedItem().(*ListItem)
		cmds.Add(app.Editor.SetSelectedItem(selectedItem))
	}

	cmd = cmds.Batch()
	return
}

// View implements [tea.Model].
func (app *App) View() tea.View {
	if app.quitting {
		return tea.NewView("")
	}

	width, height := app.size.Width, app.size.Height
	viewHeight := app.viewHeight()
	listViewWidth := app.listViewWidth()
	editorViewWidth := app.editorViewWidth()

	comp := lg.NewCompositor()

	if app.Modal != nil {
		modalView := ui.StyleRounded.Render(app.Modal.View().Content)
		modalLayer := lg.NewLayer(modalView)
		modalLayer.X((width - modalLayer.Width()) / 2).Y((height - modalLayer.Height()) / 2).Z(50)
		comp.AddLayers(modalLayer)
	}

	{
		helpView := app.Help.View(&app.KeyMap)
		var helpLayer *lg.Layer
		if app.Help.ShowAll {
			helpView = lg.NewStyle().
				// Width(width).
				Padding(0, 1).
				BorderForeground(lg.Magenta).
				Border(lg.RoundedBorder()).
				Render(helpView)
			helpLayer = lg.NewLayer(helpView)
			helpLayer.X((width - helpLayer.Width()) / 2).Y(height - helpLayer.Height()).Z(10)
		} else {
			helpLayer = lg.NewLayer(helpView).X(1).Y(height - 1)
		}
		comp.AddLayers(helpLayer)
	}

	{
		listStyle := lg.NewStyle().Width(listViewWidth).Height(viewHeight).Border(lg.RoundedBorder())
		if !app.editorFocused {
			listStyle = listStyle.BorderForeground(lg.Magenta)
		}
		listView := lg.NewStyle().MaxWidth(listViewWidth - 2).MaxHeight(viewHeight - 2).Render(app.List.View())
		listView = listStyle.Render(listView)
		listLayer := lg.NewLayer(listView)
		comp.AddLayers(listLayer)
	}

	{
		listTitleStyle := lg.NewStyle().Padding(0, 1)
		if !app.editorFocused {
			listTitleStyle = listTitleStyle.Foreground(lg.Magenta).Bold(true)
		}
		listTitleLayer := lg.NewLayer(listTitleStyle.Render("Options"))
		listTitleLayer.X((listViewWidth - listTitleLayer.Width()) / 2).Y(0).Z(1)
		comp.AddLayers(listTitleLayer)
	}

	{
		pagerBorder := lg.Border{
			Left:  "[",
			Right: "]",
		}
		pagerStyle := lg.NewStyle().Padding(0, 1).Border(pagerBorder, false, true)
		if !app.editorFocused {
			pagerStyle = pagerStyle.BorderForeground(lg.Magenta)
		}
		pagerView := pagerStyle.Render(app.List.Paginator.View())
		pagerLayer := lg.NewLayer(pagerView)
		pagerLayer.X((listViewWidth - pagerLayer.Width()) / 2).Y(viewHeight - 1).Z(1)
		comp.AddLayers(pagerLayer)
	}

	{
		editorStyle := lg.NewStyle().Padding(0, 1).Width(editorViewWidth).Height(viewHeight).Border(lg.RoundedBorder())
		if app.editorFocused {
			editorStyle = editorStyle.BorderForeground(lg.Magenta)
		}
		editorView := editorStyle.Render(app.Editor.View().Content)
		editorLayer := lg.NewLayer(editorView).X(listViewWidth)
		comp.AddLayers(editorLayer)
	}

	{
		editorTitleStyle := lg.NewStyle().Padding(0, 1)
		if app.editorFocused {
			editorTitleStyle = editorTitleStyle.Foreground(lg.Magenta).Bold(true)
		}
		editorTitleLayer := lg.NewLayer(editorTitleStyle.Render("Editor"))
		editorTitleLayer.X(listViewWidth + (editorViewWidth-editorTitleLayer.Width())/2).Y(0).Z(1)
		comp.AddLayers(editorTitleLayer)
	}

	{
		lastKeyView := ui.StyleDimmed.Padding(0, 1).Render(app.lastKey)
		lastKeyLayer := lg.NewLayer(lastKeyView)
		lastKeyLayer.X(width - lastKeyLayer.Width() - 1).Y(viewHeight - 2).Z(1)
		comp.AddLayers(lastKeyLayer)
	}

	view := tea.NewView(comp.Render())
	view.AltScreen = true
	if app.List.FilterState() == list.Filtering {
		view.Cursor = tea.NewCursor(4, 2)
	}

	return view
}

// HasError implements [ui.Model].
func (app *App) HasError() error {
	return app.err
}

func (app *App) SaveOptions() tea.Cmd {
	ctx, cancel := context.WithTimeout(app.ctx, time.Second)
	options := app.Editor.options
	return func() tea.Msg {
		defer cancel()
		err := fcitx5.Controller.SetAddonConfig(ctx, "yume", options)
		if err != nil {
			return MsgError{Err: err}
		}
		field := huh.NewNote().Title("Options Saved!")
		modal := ui.NewHuhModal(field)
		return ui.MsgModal{Modal: modal}
	}
}

func (app *App) UpdateFocus() {
	if app.editorFocused {
		app.KeyMap.ListKeys = nil
		app.KeyMap.EditorKeys = &app.Editor.KeyMap
	} else {
		app.KeyMap.ListKeys = &app.List.KeyMap
		app.KeyMap.EditorKeys = nil
	}
	app.KeyMap.Focus.SetEnabled(!app.editorFocused)
	app.KeyMap.Unfocus.SetEnabled(app.editorFocused)
}

// listViewWidth returns the width of list view including paddings and borders.
func (app *App) listViewWidth() int {
	return app.size.Width / 3
}

// editorViewWidth returns the width of editor view including paddings and borders.
func (app *App) editorViewWidth() int {
	return app.size.Width - app.listViewWidth()
}

// viewHeight returns the height of list view and editor view including paddings and borders.
func (app *App) viewHeight() int {
	return app.size.Height - 1
}
