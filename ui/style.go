package ui

import (
	lg "charm.land/lipgloss/v2"
)

var (
	StyleDimmed  = lg.NewStyle().Faint(true)
	StyleRounded = lg.NewStyle().Border(lg.RoundedBorder())
	StyleBold    = lg.NewStyle().Bold(true)
	StyleCyan    = lg.NewStyle().Foreground(lg.Cyan)
	StyleMagenta = lg.NewStyle().Foreground(lg.Magenta)
)
