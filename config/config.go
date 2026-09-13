package config

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"gopkg.in/yaml.v3"
)

type Settings struct {
	Ignore          []string
	Keybindings     *keymap.Bindings
	Locale          *locale.Translator
	Theme           theme.Theme
	HistoryMetadata bool
	Document        Document
}

type HistoryMode string

const (
	HistoryMetadata HistoryMode = "metadata"
	HistoryFull     HistoryMode = "full"
)

type Document struct {
	Language    string                       `yaml:"language"`
	Ignore      []string                     `yaml:"ignore,omitempty"`
	Languages   map[string]map[string]string `yaml:"languages,omitempty"`
	Keybindings map[string][]string          `yaml:"keybindings"`
	Theme       theme.Config                 `yaml:"theme"`
	History     HistoryMode                  `yaml:"history"`
}

func DefaultPath() (string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}
	return filepath.Join(homeDirectory, ".config", "lazyrest", "config.yml"), nil
}

// HistoryPath returns the legacy shared history file. New sessions use
// ProjectHistoryPath; this remains available for callers that need to locate or
// remove data written before histories were isolated.
func HistoryPath() (string, error) {
	configPath, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(configPath), "history.json"), nil
}

// ProjectHistoryPath returns the private history file for a canonical project
// root. The path itself contains only a stable hash of that root.
func ProjectHistoryPath(rootDirectory string) (string, error) {
	configPath, err := DefaultPath()
	if err != nil {
		return "", err
	}
	absoluteRoot, err := filepath.Abs(rootDirectory)
	if err != nil {
		return "", fmt.Errorf("resolve project root for history: %w", err)
	}
	canonicalRoot := filepath.Clean(absoluteRoot)
	if evaluated, evaluateErr := filepath.EvalSymlinks(canonicalRoot); evaluateErr == nil {
		canonicalRoot = evaluated
	}
	projectID := sha256.Sum256([]byte(canonicalRoot))
	return filepath.Join(filepath.Dir(configPath), "history", fmt.Sprintf("%x.json", projectID)), nil
}

func ProjectPath(rootDirectory string) string {
	return filepath.Join(rootDirectory, ".lazyrest.yml")
}

func DefaultDocument() Document {
	return Document{Language: "en", Keybindings: keymap.Default().Map(), Theme: theme.DefaultConfig(), History: HistoryMetadata}
}

func LoadDefault() (Settings, string, error) {
	path, err := DefaultPath()
	if err != nil {
		return Settings{}, "", err
	}
	settings, err := Load(path)
	return settings, path, err
}

func Load(path string) (Settings, error) {
	return LoadFiles([]string{path})
}

func LoadFiles(paths []string) (Settings, error) {
	document := DefaultDocument()
	for _, path := range paths {
		layer, err := read(path)
		if err != nil {
			return Settings{}, err
		}
		merge(&document, layer)
	}
	bindings, err := keymap.New(document.Keybindings)
	if err != nil {
		return Settings{}, fmt.Errorf("validate configuration: %w", err)
	}
	translator, err := locale.New(document.Language, document.Languages)
	if err != nil {
		return Settings{}, fmt.Errorf("validate configuration: %w", err)
	}
	uiTheme, err := theme.FromConfig(document.Theme)
	if err != nil {
		return Settings{}, fmt.Errorf("validate configuration: %w", err)
	}
	if document.History != HistoryMetadata && document.History != HistoryFull {
		return Settings{}, fmt.Errorf("validate configuration: history must be %q or %q", HistoryMetadata, HistoryFull)
	}
	return Settings{
		Ignore:          document.Ignore,
		Keybindings:     bindings,
		Locale:          translator,
		Theme:           uiTheme,
		HistoryMetadata: document.History == HistoryMetadata,
		Document:        document,
	}, nil
}

func Marshal(document Document) ([]byte, error) {
	contents, err := yaml.Marshal(document)
	if err != nil {
		return nil, fmt.Errorf("render configuration: %w", err)
	}
	return contents, nil
}

func Generate(path string) error {
	contents, err := Marshal(DefaultDocument())
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create config %s: %w", path, err)
	}
	if _, err := file.Write(contents); err != nil {
		_ = file.Close()
		return fmt.Errorf("write config %s: %w", path, err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close config %s: %w", path, err)
	}
	return nil
}

// SetThemePreset updates only theme.preset in a configuration file. Editing
// the YAML node tree keeps unrelated settings, ordering, and comments intact.
func SetThemePreset(path, preset string) error {
	if _, err := theme.FromConfig(theme.Config{Preset: preset}); err != nil {
		return fmt.Errorf("validate theme preset: %w", err)
	}
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		contents = nil
	} else if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if len(contents) > 0 {
		if _, err := read(path); err != nil {
			return err
		}
	}

	document := yaml.Node{Kind: yaml.DocumentNode}
	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	document.Content = []*yaml.Node{root}
	if len(contents) > 0 {
		if err := yaml.Unmarshal(contents, &document); err != nil {
			return fmt.Errorf("parse config %s: %w", path, err)
		}
		if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
			return fmt.Errorf("parse config %s: root must be a mapping", path)
		}
		root = document.Content[0]
	}

	themeNode := mappingValue(root, "theme")
	if themeNode == nil {
		root.Content = append(root.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "theme"},
			&yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"},
		)
		themeNode = root.Content[len(root.Content)-1]
	}
	if themeNode.Kind != yaml.MappingNode {
		return fmt.Errorf("parse config %s: theme must be a mapping", path)
	}
	presetNode := mappingValue(themeNode, "preset")
	if presetNode == nil {
		themeNode.Content = append(themeNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "preset"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: preset},
		)
	} else {
		presetNode.Kind = yaml.ScalarNode
		presetNode.Tag = "!!str"
		presetNode.Value = preset
	}

	var output bytes.Buffer
	encoder := yaml.NewEncoder(&output)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return fmt.Errorf("render config %s: %w", path, err)
	}
	if err := encoder.Close(); err != nil {
		return fmt.Errorf("render config %s: %w", path, err)
	}
	return writeConfigAtomically(path, output.Bytes())
}

// SetThemePresetInFiles updates the highest-priority configuration layer that
// already owns theme.preset. If no layer owns it, fallback receives the user
// preference so merely choosing a theme does not create a project config.
func SetThemePresetInFiles(paths []string, fallback, preset string) (string, error) {
	target := fallback
	for _, path := range paths {
		document, err := read(path)
		if err != nil {
			return "", err
		}
		if document.Theme.Preset != "" {
			target = path
		}
	}
	if target == "" {
		return "", errors.New("theme configuration path is empty")
	}
	if err := SetThemePreset(target, preset); err != nil {
		return "", err
	}
	return target, nil
}

func mappingValue(mapping *yaml.Node, key string) *yaml.Node {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == key {
			return mapping.Content[index+1]
		}
	}
	return nil
}

func writeConfigAtomically(path string, contents []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	temporary, err := os.CreateTemp(directory, ".config-*.tmp")
	if err != nil {
		return fmt.Errorf("create temporary config: %w", err)
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("secure temporary config: %w", err)
	}
	if _, err := temporary.Write(contents); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write temporary config: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close temporary config: %w", err)
	}
	if err := os.Rename(temporaryPath, path); err != nil {
		return fmt.Errorf("replace config %s: %w", path, err)
	}
	return nil
}

func read(path string) (Document, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Document{}, nil
	}
	if err != nil {
		return Document{}, fmt.Errorf("read config %s: %w", path, err)
	}
	// Unknown keys are refused rather than dropped: a typo such as
	// "keybinding" for "keybindings" would otherwise look like it worked.
	decoder := yaml.NewDecoder(bytes.NewReader(contents))
	decoder.KnownFields(true)
	var document Document
	if err := decoder.Decode(&document); err != nil {
		if errors.Is(err, io.EOF) {
			return Document{}, nil
		}
		return Document{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return document, nil
}

func merge(target *Document, source Document) {
	if source.Language != "" {
		target.Language = source.Language
	}
	if source.History != "" {
		target.History = source.History
	}
	// The lists add up: a project says what else to skip without losing what
	// the user chose.
	for _, name := range source.Ignore {
		if !slices.Contains(target.Ignore, name) {
			target.Ignore = append(target.Ignore, name)
		}
	}
	if target.Languages == nil {
		target.Languages = map[string]map[string]string{}
	}
	for language, translations := range source.Languages {
		if target.Languages[language] == nil {
			target.Languages[language] = map[string]string{}
		}
		for key, value := range translations {
			target.Languages[language][key] = value
		}
	}
	if target.Keybindings == nil {
		target.Keybindings = map[string][]string{}
	}
	for action, keys := range source.Keybindings {
		target.Keybindings[action] = append([]string(nil), keys...)
	}
	mergeStringFields(&target.Theme, source.Theme)
}

func mergeStringFields(target, source any) {
	targetValue := reflect.ValueOf(target).Elem()
	sourceValue := reflect.ValueOf(source)
	for index := range sourceValue.NumField() {
		value := sourceValue.Field(index).String()
		if value != "" {
			targetValue.Field(index).SetString(value)
		}
	}
}
