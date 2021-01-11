package settings

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
	"io/ioutil"
	"log"
	"path"
	"regexp"
	"strings"
)

const DefaultLang = "en"
const DefaultLocalePath = "locales/"

var GlobalBundle *i18n.Bundle

func InitLang(localePath, lang string) {
	if lang == "" {
		lang = DefaultLang
	}
	if localePath == "" {
		localePath = DefaultLocalePath
	}
	GlobalBundle = LoadTranslations(localePath, lang)
}

func LoadTranslations(localePath, lang string) *i18n.Bundle {
	bundle := i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)

	defaultBotLangLoaded := lang == DefaultLang
	files, err := ioutil.ReadDir(localePath)
	if err == nil {
		re := regexp.MustCompile(`^active\.(?P<lang>.*)\.toml$`)
		for _, file := range files {
			if match := re.FindStringSubmatch(file.Name()); match != nil {
				fileLang := match[re.SubexpIndex("lang")]

				if _, err := bundle.LoadMessageFile(path.Join(localePath, file.Name())); err != nil {
					if lang != DefaultLang && fileLang != DefaultLang {
						log.Println("[Locale] Error load message file:", err)
					}
				} else {
					langName, _ := i18n.NewLocalizer(bundle, fileLang).Localize(&i18n.LocalizeConfig{
						DefaultMessage: &i18n.Message{
							ID:    "locale.language.name",
							Other: "English", /* language.Make(fileLang).String() */
						},
					})

					log.Printf("[Locale] Loaded language: %s - %s", fileLang, langName)
					if lang == fileLang {
						defaultBotLangLoaded = true
						log.Printf("[Locale] Selected language is %s \n", lang)
					}
				}
			}
		}
	}
	if !defaultBotLangLoaded {
		log.Printf("[Locale] Localization file with language %s not found. The default lang is set to: %s\n", lang, DefaultLang)
		lang = DefaultLang
	}

	return bundle
}

func LocalizeMessage(args ...interface{}) string {
	if len(args) == 0 {
		return "Noup"
	}

	var templateData map[string]interface{}
	lang := DefaultLang
	message := args[0].(*i18n.Message)
	var pluralCount interface{} = nil

	// omgg, rework this

	// 1
	if len(args[1:]) > 0 {
		if model, ok := args[1].(map[string]interface{}); ok {
			templateData = model
		} else if model, ok := args[1].(string); ok {
			lang = model
		} else if model, ok := args[1].(int); ok {
			pluralCount = model
		}

		// 2
		if len(args[2:]) > 0 {
			if model, ok := args[2].(string); ok {
				lang = model
			} else if model, ok := args[2].(int); ok {
				pluralCount = model
			}

			// 3
			if len(args[3:]) > 0 {
				if model, ok := args[3].(int); ok {
					pluralCount = model
				}
			}
		}
	}

	localizer := i18n.NewLocalizer(GlobalBundle, lang)
	msg, err := localizer.Localize(&i18n.LocalizeConfig{
		DefaultMessage: message,
		TemplateData:   templateData,
		PluralCount:    pluralCount,
	})

	// fix go-i18n extract
	msg = strings.ReplaceAll(msg, "\\n", "\n")
	// log.Printf("[Locale] (%s) %s", lang, msg)

	if err != nil {
		log.Printf("[Locale] Warning: %s", err)
	}

	return msg
}
