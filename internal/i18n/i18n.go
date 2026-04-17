package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	goi18n "github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed locales/*.json
var localesFS embed.FS

const DefaultLang = "en"

var supported = []string{"en", "ru"}

var (
	mu        sync.RWMutex
	localizer *goi18n.Localizer
	active    = DefaultLang
)

func init() {
	_ = Init(DetectLang(""))
}

// Init (re-)initialises the localizer for the given language code.
func Init(lang string) error {
	if !isSupported(lang) {
		lang = DefaultLang
	}
	b := goi18n.NewBundle(language.English)
	b.RegisterUnmarshalFunc("json", json.Unmarshal)
	for _, code := range supported {
		data, err := localesFS.ReadFile("locales/" + code + ".json")
		if err != nil {
			return fmt.Errorf("i18n: load %s: %w", code, err)
		}
		if _, err := b.ParseMessageFileBytes(data, code+".json"); err != nil {
			return fmt.Errorf("i18n: parse %s: %w", code, err)
		}
	}
	mu.Lock()
	localizer = goi18n.NewLocalizer(b, lang)
	active = lang
	mu.Unlock()
	return nil
}

// T returns the translation for key; falls back to key if not found.
func T(key string) string {
	mu.RLock()
	l := localizer
	mu.RUnlock()
	if l == nil {
		return key
	}
	s, err := l.Localize(&goi18n.LocalizeConfig{MessageID: key})
	if err != nil {
		return key
	}
	return s
}

// Tf returns the translation for key with template data.
func Tf(key string, data map[string]any) string {
	mu.RLock()
	l := localizer
	mu.RUnlock()
	if l == nil {
		return key
	}
	s, err := l.Localize(&goi18n.LocalizeConfig{
		MessageID:    key,
		TemplateData: data,
	})
	if err != nil {
		return key
	}
	return s
}

// Active returns the current language code.
func Active() string {
	mu.RLock()
	defer mu.RUnlock()
	return active
}

// Available returns the list of supported language codes.
func Available() []string {
	out := make([]string, len(supported))
	copy(out, supported)
	return out
}

// DetectLang returns lang if supported, otherwise tries $LANG, falls back to DefaultLang.
func DetectLang(lang string) string {
	if isSupported(lang) {
		return lang
	}
	env := os.Getenv("LANG")
	if idx := strings.Index(env, "_"); idx > 0 {
		env = env[:idx]
	}
	if isSupported(env) {
		return env
	}
	return DefaultLang
}

func isSupported(lang string) bool {
	for _, s := range supported {
		if s == lang {
			return true
		}
	}
	return false
}
