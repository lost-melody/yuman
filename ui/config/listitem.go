package configui

import (
	"charm.land/bubbles/v2/list"
	lg "charm.land/lipgloss/v2"
	"github.com/lost-melody/yuman/fcitx5"
)

var (
	_ list.Item        = (*ListItem)(nil)
	_ list.DefaultItem = (*ListItem)(nil)
)

type ListItem struct {
	Option *fcitx5.ConfigOption
}

// FilterValue implements [list.Item].
func (item *ListItem) FilterValue() string {
	return item.Option.Title
}

// Description implements [list.DefaultItem].
func (item *ListItem) Description() string {
	return lg.Sprintf("%s: %s", item.Option.Name, item.Option.Type)
}

// Title implements [list.DefaultItem].
func (item *ListItem) Title() string {
	return item.Option.Title
}
