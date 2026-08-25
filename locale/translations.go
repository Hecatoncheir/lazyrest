package locale

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed languages/*.yml
var languageFiles embed.FS

var builtins = mustLoadBuiltins()

func mustLoadBuiltins() map[string]map[string]string {
	entries, err := languageFiles.ReadDir("languages")
	if err != nil {
		panic(fmt.Errorf("read embedded languages: %w", err))
	}
	result := make(map[string]map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}
		contents, err := languageFiles.ReadFile("languages/" + entry.Name())
		if err != nil {
			panic(fmt.Errorf("read embedded language %s: %w", entry.Name(), err))
		}
		translations := map[string]string{}
		if err := yaml.Unmarshal(contents, &translations); err != nil {
			panic(fmt.Errorf("parse embedded language %s: %w", entry.Name(), err))
		}
		code := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		result[code] = translations
	}
	return result
}
