package locale

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuiltInLanguagesAndOverrides(t *testing.T) {
	for _, language := range []string{"en", "ru", "es"} {
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
