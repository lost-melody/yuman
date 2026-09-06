// Package yumeui implements the yume config ui.
package yumeui

import (
	"context"

	"github.com/lost-melody/yuman/fcitx5"
	"github.com/lost-melody/yuman/ui"
	configui "github.com/lost-melody/yuman/ui/config"
)

func RunYumeConfigApp(ctx context.Context) (err error) {
	config, err := fcitx5.Controller.GetAddonConfig(ctx, "yume")
	if err != nil {
		return
	}
	if config.Options == nil || len(config.Schemes) == 0 {
		return
	}

	optionList := configui.NewOptionList(config)
	editor := configui.NewEditorModel(config.Options).
		WithHandlers(
			NewCommentModes(),
			NewWordWhitelist(),
			NewHiddenSchemas(),
			configui.NewBoolHandler(),
			configui.NewIntHandler(),
			configui.NewStringHandler(),
			configui.NewEnumHandler(),
			configui.NewStringsHandler(),
		)
	app := configui.NewApp(ctx, optionList, editor)

	err = ui.RunApp(ctx, &app)
	if err != nil {
		return
	}

	return
}
