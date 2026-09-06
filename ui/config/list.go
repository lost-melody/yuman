package configui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	"github.com/lost-melody/yuman/fcitx5"
	"github.com/samber/lo"
)

func NewOptionList(config *fcitx5.Config) list.Model {
	items := lo.Map(config.Schemes[0].Options, func(option *fcitx5.ConfigOption, _ int) list.Item {
		return &ListItem{
			Option: option,
		}
	})
	delegate := list.NewDefaultDelegate()
	model := list.New(items, delegate, 0, 0)

	model.InfiniteScrolling = true
	model.KeyMap.ShowFullHelp = key.NewBinding()
	model.KeyMap.Filter = key.NewBinding()
	model.KeyMap.Quit = key.NewBinding()

	model.SetShowTitle(false)
	model.SetShowFilter(false)
	model.SetShowStatusBar(false)
	model.SetShowPagination(false)
	model.SetShowHelp(false)

	return model
}
