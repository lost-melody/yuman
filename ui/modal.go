package ui

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/samber/lo"
)

var _ Modal = (*HuhModal)(nil)

var (
	_ MsgModalDone = MsgModalSubmit{}
	_ MsgModalDone = MsgModalCancel{}
)

// Modal represents a modal dialog.
// The Update() methods should return (nil, CmdModalSubmit) on finished.
type Modal interface {
	tea.Model
	ModalID() int64
}

type (
	MsgModal       struct{ Modal Modal }
	MsgModalSubmit struct{ ID int64 }
	MsgModalCancel struct{ ID int64 }
)

// MsgModalDone includes [MsgModalSubmit] and [MsgModalCancel].
type MsgModalDone interface {
	ModalID() int64
}

func CmdModal(modal Modal) tea.Cmd {
	return func() tea.Msg {
		return MsgModal{Modal: modal}
	}
}

func CmdModalSubmit(id int64) tea.Cmd {
	return func() tea.Msg {
		return MsgModalSubmit{ID: id}
	}
}

func CmdModalCancel(id int64) tea.Cmd {
	return func() tea.Msg {
		return MsgModalCancel{ID: id}
	}
}

// HuhModal implements [tea.Model].
type HuhModal struct {
	id     int64
	fields []huh.Field
	Model  huh.Model
}

// ModalID implements [Modal].
func (h *HuhModal) ModalID() int64 {
	return h.id
}

// NewHuhModal wraps some [huh.Field] in a [huh.Form] and returns a [HuhModal].
// [MsgModalSubmit] or [MsgModalCancel] is emit on finished.
func NewHuhModal(fields ...huh.Field) *HuhModal {
	fields = lo.Filter(fields, func(field huh.Field, _ int) bool {
		return field != nil
	})
	if len(fields) == 0 {
		return &HuhModal{}
	}
	form := huh.NewForm(
		huh.NewGroup(
			fields...,
		).WithWidth(64),
	)
	id := time.Now().UnixNano()
	form.SubmitCmd = CmdModalSubmit(id)
	form.CancelCmd = CmdModalCancel(id)
	return &HuhModal{
		id:     id,
		fields: fields,
		Model:  form,
	}
}

// Init implements [tea.Model].
func (h *HuhModal) Init() tea.Cmd {
	var cmds Cmds
	cmds.Add(h.Model.Init())
	if len(h.fields) != 0 {
		cmds.Add(h.fields[0].Focus())
	}
	return cmds.Batch()
}

// Update implements [tea.Model].
func (h *HuhModal) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if h.Model == nil {
		return nil, CmdModalSubmit(0)
	}
	switch m := msg.(type) {
	case MsgModalSubmit:
		return nil, CmdModalSubmit(m.ID)
	case MsgModalCancel:
		return nil, CmdModalCancel(m.ID)
	default:
		var cmd tea.Cmd
		h.Model, cmd = h.Model.Update(msg)
		return h, cmd
	}
}

// View implements [tea.Model].
func (h *HuhModal) View() tea.View {
	if h.Model == nil {
		return tea.NewView("")
	}
	return tea.NewView(h.Model.View())
}

// ModalID implements [MsgModalDone].
func (m MsgModalCancel) ModalID() int64 {
	return m.ID
}

// ModalID implements [MsgModalDone].
func (m MsgModalSubmit) ModalID() int64 {
	return m.ID
}
