package locale

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltInLanguagesAndOverrides(t *testing.T) {
	for _, language := range []string{"en", "ru", "es", "zh"} {
		translator, err := New(language, nil)
		if err != nil {
			t.Fatal(err)
		}
		if translator.Text("files") == "" {
			t.Fatalf("missing files translation for %s", language)
		}
	}
	translator, err := New("ru", map[string]map[string]string{"ru": {"files": "Мои файлы"}})
	if err != nil {
		t.Fatal(err)
	}
	if translator.Text("files") != "Мои файлы" || translator.Text("suite") == "Suite" {
		t.Fatal("override or built-in Russian translation was not applied")
	}
}

func TestBuiltInLanguagesContainEveryEnglishKey(t *testing.T) {
	for language, translations := range builtins {
		for key := range builtins["en"] {
			if _, ok := translations[key]; !ok {
				t.Errorf("language %s is missing translation %q", language, key)
			}
		}
	}
}

func TestChineseDiagnosticCounter(t *testing.T) {
	translator, err := New("zh", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := translator.PluralDiagnostics(2); got != "2 条诊断" {
		t.Fatalf("unexpected Chinese diagnostic counter: %q", got)
	}
}

func TestLoadSelectsLanguageAndFallsBackToEnglish(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yml")
	contents := "language: es\nlanguages:\n  es:\n    files: Mis archivos\n"
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	translator, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if translator.Language() != "es" || translator.Text("files") != "Mis archivos" {
		t.Fatal("selected language override was not loaded")
	}
	if translator.Text("running") != "Ejecutando" || translator.Text("loading_environment") == "" {
		t.Fatal("built-in translation or English fallback is missing")
	}
}

func TestRussianDiagnosticPluralForms(t *testing.T) {
	translator, err := New("ru", nil)
	if err != nil {
		t.Fatal(err)
	}
	tests := map[int]string{1: "1 диагностика", 2: "2 диагностики", 5: "5 диагностик", 11: "11 диагностик", 21: "21 диагностика"}
	for count, want := range tests {
		if got := translator.PluralDiagnostics(count); got != want {
			t.Fatalf("unexpected plural for %d: got %q, want %q", count, got, want)
		}
	}
}
