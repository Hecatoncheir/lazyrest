package locale

import "fmt"

type Translator struct {
	language string
	texts    map[string]string
}

func English() *Translator {
	translator, _ := New("en", nil)
	return translator
}

func New(language string, overrides map[string]map[string]string) (*Translator, error) {
	if language == "" {
		language = "en"
	}
	base, ok := builtins[language]
	if !ok {
		if _, configured := overrides[language]; !configured {
			return nil, fmt.Errorf("unknown language %q", language)
		}
		base = map[string]string{}
	}
	texts := clone(builtins["en"])
	merge(texts, base)
	merge(texts, overrides[language])
	return &Translator{language: language, texts: texts}, nil
}

func (translator *Translator) Language() string {
	if translator == nil {
		return "en"
	}
	return translator.language
}

func (translator *Translator) Text(key string) string {
	if translator == nil {
		return builtins["en"][key]
	}
	if value, ok := translator.texts[key]; ok {
		return value
	}
	return key
}

func (translator *Translator) Format(key string, arguments ...any) string {
	return fmt.Sprintf(translator.Text(key), arguments...)
}

func (translator *Translator) PluralDiagnostics(count int) string {
	key := "diagnostics_many"
	switch translator.Language() {
	case "ru":
		lastTwo := count % 100
		last := count % 10
		if last == 1 && lastTwo != 11 {
			key = "diagnostics_one"
		} else if last >= 2 && last <= 4 && (lastTwo < 12 || lastTwo > 14) {
			key = "diagnostics_few"
		}
	default:
		if count == 1 {
			key = "diagnostics_one"
		}
	}
	return translator.Format(key, count)
}

func clone(source map[string]string) map[string]string {
	result := make(map[string]string, len(source))
	merge(result, source)
	return result
}

func merge(target, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}
