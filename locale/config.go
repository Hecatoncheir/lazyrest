package locale

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type fileConfig struct {
	Language  string                       `yaml:"language"`
	Languages map[string]map[string]string `yaml:"languages"`
}

func LoadDefault() (*Translator, string, error) {
	homeDirectory, err := os.UserHomeDir()
	if err != nil {
		return nil, "", fmt.Errorf("resolve user home directory: %w", err)
	}
	path := filepath.Join(homeDirectory, ".config", "lazyrest", "config.yml")
	translator, err := Load(path)
	return translator, path, err
}

func Load(path string) (*Translator, error) {
	contents, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return English(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}
	var config fileConfig
	if err := yaml.Unmarshal(contents, &config); err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	translator, err := New(config.Language, config.Languages)
	if err != nil {
		return nil, fmt.Errorf("parse config %s: %w", path, err)
	}
	return translator, nil
}
