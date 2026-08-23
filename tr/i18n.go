// Package tr serves as i18n localizer.
package tr

import (
	"embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var (
	Bundle    *i18n.Bundle
	Localizer *i18n.Localizer
)

//go:embed locales/active.*.toml
var LocalesFS embed.FS

func init() {
	Bundle = i18n.NewBundle(language.English)
	Bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	for _, lang := range []string{"zh-CN", "zh-TW"} {
		_, err := Bundle.LoadMessageFileFS(LocalesFS, fmt.Sprintf("locales/active.%s.toml", lang))
		if err != nil {
			panic(fmt.Errorf("loading locale %s: %w", lang, err))
		}
	}
	Localizer = i18n.NewLocalizer(Bundle, getLocaleFromEnv())
}

func Localize(lc *i18n.LocalizeConfig) string {
	msg, _ := Localizer.Localize(lc)
	return msg
}

func LocalizeError(lc *i18n.LocalizeConfig) error {
	return errors.New(Localize(lc))
}

func getLocaleFromEnv() string {
	lang, _ := os.LookupEnv("LANG")
	if lang != "" {
		return strings.ReplaceAll(strings.Split(lang, ".")[0], "_", "-")
	}
	return "en-US"
}
