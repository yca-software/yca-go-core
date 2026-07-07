package chi_localizer

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

type Localizer interface {
	Translate(lang, key string, data map[string]any) string
}

type localizer struct {
	bundle             *i18n.Bundle
	localizers         map[string]*i18n.Localizer
	supportedLanguages []string
	defaultLanguage    string
}

func New(supportedLanguages []string, defaultLanguage string, localesPath string) Localizer {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	localizers := make(map[string]*i18n.Localizer)

	for _, lang := range supportedLanguages {
		filePath := fmt.Sprintf("%s/%s.json", localesPath, lang)
		if _, err := os.Stat(filePath); err == nil {
			bundle.MustLoadMessageFile(filePath)
		}
		// Pre-cache the localizer for this language
		localizers[lang] = i18n.NewLocalizer(bundle, lang)
	}

	return &localizer{
		bundle:             bundle,
		localizers:         localizers,
		supportedLanguages: supportedLanguages,
		defaultLanguage:    defaultLanguage,
	}
}

func (t *localizer) Translate(lang, key string, data map[string]any) string {
	loc, ok := t.localizers[lang]
	if !ok {
		loc = t.localizers[t.defaultLanguage]
	}

	msg, err := loc.Localize(&i18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key // Fallback to key name so you know what's missing
	}
	return msg
}
