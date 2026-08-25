package keymap

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	Keybindings map[string][]string `yaml:"keybindings"`
}

func LoadDefault() (*Bindings, string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("resolve user home directory: %w", err)
	}
	path := filepath.Join(homeDirectory, ".config", "lazyrest", "config.yml")
	bindings, err := Load(path)
	return bindings, path, err
}

func Load(path string) (*Bindings, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return Default(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var config fileConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	bindings, err := New(config.Keybindings)
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return bindings, nil
}
