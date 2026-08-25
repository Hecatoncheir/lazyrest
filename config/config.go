package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Hecatoncheir/lazyrest/keymap"
	"github.com/Hecatoncheir/lazyrest/locale"
	"github.com/Hecatoncheir/lazyrest/ui/theme"
	"gopkg.in/yaml.v3"
)

type Settings struct {
	Keybindings *keymap.Bindings
	Locale      *locale.Translator
	Theme       theme.Theme
}

type fileConfig struct {
	Language    string                       `yaml:"language"`
	Languages   map[string]map[string]string `yaml:"languages"`
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

func LoadDefault() (Settings, string, error) {
	path, err := DefaultPath()
	if err != nil {
		return Settings{}, "", err
	}
	settings, err := Load(path)
	return settings, path, err
}

func Load(path string) (Settings, error) {
	config := fileConfig{}
	contents, err := os.ReadFile(path)
	if err == nil {
		if err := yaml.Unmarshal(contents, &config); err != nil {
			return Settings{}, fmt.Errorf("parse config %s: %w", path, err)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return Settings{}, fmt.Errorf("read config %s: %w", path, err)
	}

	bindings, err := keymap.New(config.Keybindings)
	if err != nil {
		return Settings{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	translator, err := locale.New(config.Language, config.Languages)
	if err != nil {
		return Settings{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	uiTheme, err := theme.FromConfig(config.Theme)
	if err != nil {
		return Settings{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return Settings{Keybindings: bindings, Locale: translator, Theme: uiTheme}, nil
}
