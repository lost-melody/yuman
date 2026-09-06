// Package ui defines TUI primitives.
package ui

import (
	"context"

	tea "charm.land/bubbletea/v2"
)

type Model interface {
	tea.Model
	HasError() error
}

type Cmds []tea.Cmd

func RunApp(ctx context.Context, model Model) (err error) {
	program := tea.NewProgram(model, tea.WithContext(ctx))
	returnModel, err := program.Run()
	if err != nil {
		return
	}
	if model, _ = returnModel.(Model); model != nil {
		err = model.HasError()
	}
	return
}

func (cmds *Cmds) Add(cmd ...tea.Cmd) {
	for _, c := range cmd {
		if c != nil {
			*cmds = append(*cmds, c)
		}
	}
}

func (cmds *Cmds) Batch() tea.Cmd {
	if len(*cmds) == 0 {
		return nil
	}
	return tea.Batch(*cmds...)
}

func (cmds *Cmds) Sequence() tea.Cmd {
	if len(*cmds) == 0 {
		return nil
	}
	return tea.Sequence(*cmds...)
}
