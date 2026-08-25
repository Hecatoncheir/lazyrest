package theme

import "fmt"

var presetNames = []string{"gruvbox", "catppuccin-mocha", "tokyo-night", "dracula", "nord", "monokai"}

var presets = map[string]Config{
	"gruvbox": {
		Background: "#1d2021", PanelBackground: "#504945", PanelFocus: "#3c3836", Foreground: "#fbf1c7", Muted: "#bdae93", Accent: "#83a598",
		Border: "#bdae93", BorderFocus: "#fbf1c7", SelectionBackground: "#fbf1c7", SelectionForeground: "#282828", Progress: "#fabd2f",
		ProgressForeground: "#3c3836", Success: "#b8bb26", SuccessForeground: "#3c3836", Failure: "#d65d0e", FailureForeground: "#fbf1c7",
		Breadcrumb: "#bdae93", BreadcrumbForeground: "#504945",
	},
	"catppuccin-mocha": {
		Background: "#11111b", PanelBackground: "#1e1e2e", PanelFocus: "#313244", Foreground: "#cdd6f4", Muted: "#a6adc8", Accent: "#89b4fa",
		Border: "#6c7086", BorderFocus: "#cdd6f4", SelectionBackground: "#89b4fa", SelectionForeground: "#11111b", Progress: "#f9e2af",
		ProgressForeground: "#11111b", Success: "#a6e3a1", SuccessForeground: "#11111b", Failure: "#f38ba8", FailureForeground: "#11111b",
		Breadcrumb: "#cba6f7", BreadcrumbForeground: "#11111b",
	},
	"tokyo-night": {
		Background: "#1a1b26", PanelBackground: "#24283b", PanelFocus: "#292e42", Foreground: "#c0caf5", Muted: "#9aa5ce", Accent: "#7aa2f7",
		Border: "#565f89", BorderFocus: "#c0caf5", SelectionBackground: "#7aa2f7", SelectionForeground: "#1a1b26", Progress: "#e0af68",
		ProgressForeground: "#1a1b26", Success: "#9ece6a", SuccessForeground: "#1a1b26", Failure: "#f7768e", FailureForeground: "#1a1b26",
		Breadcrumb: "#bb9af7", BreadcrumbForeground: "#1a1b26",
	},
	"dracula": {
		Background: "#282a36", PanelBackground: "#44475a", PanelFocus: "#3b3d4d", Foreground: "#f8f8f2", Muted: "#6272a4", Accent: "#8be9fd",
		Border: "#6272a4", BorderFocus: "#f8f8f2", SelectionBackground: "#bd93f9", SelectionForeground: "#282a36", Progress: "#f1fa8c",
		ProgressForeground: "#282a36", Success: "#50fa7b", SuccessForeground: "#282a36", Failure: "#ff5555", FailureForeground: "#282a36",
		Breadcrumb: "#ff79c6", BreadcrumbForeground: "#282a36",
	},
	"nord": {
		Background: "#2e3440", PanelBackground: "#3b4252", PanelFocus: "#434c5e", Foreground: "#eceff4", Muted: "#d8dee9", Accent: "#88c0d0",
		Border: "#4c566a", BorderFocus: "#eceff4", SelectionBackground: "#88c0d0", SelectionForeground: "#2e3440", Progress: "#ebcb8b",
		ProgressForeground: "#2e3440", Success: "#a3be8c", SuccessForeground: "#2e3440", Failure: "#bf616a", FailureForeground: "#eceff4",
		Breadcrumb: "#b48ead", BreadcrumbForeground: "#2e3440",
	},
	"monokai": {
		Background: "#272822", PanelBackground: "#3e3d32", PanelFocus: "#49483e", Foreground: "#f8f8f2", Muted: "#a59f85", Accent: "#66d9ef",
		Border: "#75715e", BorderFocus: "#f8f8f2", SelectionBackground: "#66d9ef", SelectionForeground: "#272822", Progress: "#e6db74",
		ProgressForeground: "#272822", Success: "#a6e22e", SuccessForeground: "#272822", Failure: "#f92672", FailureForeground: "#f8f8f2",
		Breadcrumb: "#ae81ff", BreadcrumbForeground: "#272822",
	},
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
