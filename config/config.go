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
	Ignore      []string
	Keybindings *keymap.Bindings
	Locale      *locale.Translator
	Theme       theme.Theme
	Document    Document
}

type Document struct {
	Language    string                       `yaml:"language"`
	Ignore      []string                     `yaml:"ignore,omitempty"`
	Languages   map[string]map[string]string `yaml:"languages,omitempty"`
	Keybindings map[string][]string          `yaml:"keybindings"`
	Theme       theme.Config                 `yaml:"theme"`
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
	return Document{Language: "en", Keybindings: keymap.Default().Map(), Theme: theme.DefaultConfig()}
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
	return Settings{Ignore: document.Ignore, Keybindings: bindings, Locale: translator, Theme: uiTheme, Document: document}, nil
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
