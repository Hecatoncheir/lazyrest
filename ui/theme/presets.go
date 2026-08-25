package theme

import (
	"embed"
	"fmt"
	"path"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed presets/*.yml
var presetFiles embed.FS

var presets, presetNames = mustLoadPresets()

func mustLoadPresets() (map[string]Config, []string) {
	entries, err := presetFiles.ReadDir("presets")
	if err != nil {
		panic(fmt.Errorf("read embedded theme presets: %w", err))
	}
	loaded := make(map[string]Config, len(entries))
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || path.Ext(entry.Name()) != ".yml" {
			continue
		}
		contents, err := presetFiles.ReadFile("presets/" + entry.Name())
		if err != nil {
			panic(fmt.Errorf("read embedded theme preset %s: %w", entry.Name(), err))
		}
		var config Config
		if err := yaml.Unmarshal(contents, &config); err != nil {
			panic(fmt.Errorf("parse embedded theme preset %s: %w", entry.Name(), err))
		}
		name := strings.TrimSuffix(entry.Name(), path.Ext(entry.Name()))
		loaded[name] = config
		names = append(names, name)
	}
	sort.Strings(names)
	for index, name := range names {
		if name == "gruvbox" {
			names = append([]string{name}, append(names[:index], names[index+1:]...)...)
			break
		}
	}
	return loaded, names
}

func PresetNames() []string {
	return append([]string(nil), presetNames...)
}

func resolvePreset(overrides Config) (Config, error) {
	name := overrides.Preset
	if name == "" {
		name = "gruvbox"
	}
	resolved, ok := presets[name]
	if !ok {
		return Config{}, fmt.Errorf("unknown theme preset %q", name)
	}
	resolved.Preset = name
	mergeThemeOverrides(&resolved, overrides)
	return resolved, nil
}

func mergeThemeOverrides(target *Config, overrides Config) {
	values := []*string{&target.Background, &target.PanelBackground, &target.PanelFocus, &target.Foreground, &target.Muted, &target.Accent,
		&target.Border, &target.BorderFocus, &target.SelectionBackground, &target.SelectionForeground, &target.Progress, &target.ProgressForeground,
		&target.Success, &target.SuccessForeground, &target.Failure, &target.FailureForeground, &target.Breadcrumb, &target.BreadcrumbForeground}
	overrideValues := []string{overrides.Background, overrides.PanelBackground, overrides.PanelFocus, overrides.Foreground, overrides.Muted, overrides.Accent,
		overrides.Border, overrides.BorderFocus, overrides.SelectionBackground, overrides.SelectionForeground, overrides.Progress, overrides.ProgressForeground,
		overrides.Success, overrides.SuccessForeground, overrides.Failure, overrides.FailureForeground, overrides.Breadcrumb, overrides.BreadcrumbForeground}
	for index, value := range overrideValues {
		if value != "" {
			*values[index] = value
		}
	}
}
