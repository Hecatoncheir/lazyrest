package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"gopkg.in/yaml.v3"
)

type Settings struct {
	Keybindings *keymap.Bindings
	Locale      *locale.Translator
	Theme       theme.Theme
	Document    Document
}

type Document struct {
	Language    string                       `yaml:"language"`
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

func HistoryPath() (string, error) {
	configPath, err := DefaultPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(configPath), "history.json"), nil
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
	return Settings{Keybindings: bindings, Locale: translator, Theme: uiTheme, Document: document}, nil
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
	defer file.Close()
	if _, err := file.Write(contents); err != nil {
		return fmt.Errorf("write config %s: %w", path, err)
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
